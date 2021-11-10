// Package config provides configuration variable primitives and the AppConfig object,
// which can be used to load, validate and retrieve configuration items.
package config

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/universal-devs/go-utilities/constants"
)

// Variable represents a single configuration item.
type Variable struct {
	// Value is the actual value of the Variable.
	Value string

	// DefaultValue is the default value of the Variable.
	DefaultValue string

	// Description is the brief description of the Variable.
	Description string

	// Rules are a map of named validation.Rules that should apply to the Variable's Value.
	Rules map[string]validation.Rule
}

// AppConfig is the collection of application configuration items of an application.
type AppConfig struct {
	vars map[string]*Variable
}

// NewConfig creates a new AppConfig with the supplied default Variables.
func NewConfig(defaults map[string]*Variable) *AppConfig {
	conf := &AppConfig{
		vars: make(map[string]*Variable),
	}
	if defaults != nil {
		conf.vars = defaults
	}
	return conf
}

// Setup the Application's Configuration according to the defaults, environment variables and the envfile(s).
// If no env file supplied, only the defaults and environment variables will be checked.
// Return an error if the config file(s) cannot be loaded, or the configurations are invalid.
// On invalid configs the returned error will be the type validation.Errors.
func (appConf *AppConfig) Setup(envfiles ...string) error {
	if err := appConf.loadEnv(envfiles...); err != nil {
		return errors.Wrap(err, "Failed to set Application Configuration")
	}
	return appConf.Validate()
}

// Lookup returns the named Application Configuration Variable's value (or an empty string),
// and a boolean indicating if it was find or not.
func (appConf *AppConfig) Lookup(name string) (string, bool) {
	if val, ok := appConf.vars[name]; ok {
		return val.Value, true
	}
	return "", false
}

// Get returns the named Application Configuration Variable's value. If it is not set, an empty string is returned.
func (appConf *AppConfig) Get(name string) string {
	val, _ := appConf.Lookup(name)
	return val
}

// ValidationErrors applies on each Variable its own validation rules, unifies the errors and returns them.
func (appConf *AppConfig) ValidationErrors() validation.Errors {
	// allErrors collects all validation errors
	allErrors := validation.Errors{}

	// iterate over variables
	for confKey, confVar := range appConf.vars {
		// validationErrors collects all validation error associated with one variable
		validationErrors := validation.Errors{}
		// iterate over rules
		for ruleName, rule := range confVar.Rules {
			// call the rule on the value and collect errors
			if err := rule.Validate(confVar.Value); err != nil {
				validationErrors[ruleName] = err
			}
		}
		// if there were any validation error add them to the top level collection
		if len(validationErrors) > 0 {
			allErrors[fmt.Sprintf("%s = %s", confKey, confVar.Value)] = validationErrors.Filter()
		}
	}

	if len(allErrors) > 0 {
		return allErrors
	}

	return nil
}

// Validate collects all ValidationErrors and filter them into one error.
func (appConf *AppConfig) Validate() error {
	errs := appConf.ValidationErrors()
	if len(errs) > 0 {
		return errs.Filter()
	}
	return nil
}

// loadEnv loads variables from the envfile(s) and the environment, into the AppConfig.
// Variables in the envfile(s) takes precedence over environment variables.
func (appConf *AppConfig) loadEnv(envfiles ...string) error {
	// If any env file is provided try load it.
	if len(envfiles) > 0 {
		// Overload existing environment variables with the ones in the envfile(s).
		if err := godotenv.Overload(envfiles...); err != nil {
			return errors.Wrap(err, "Failed to overload variables with envfile(s)")
		}
	}

	// Iterate over all Variables
	for confKey, confVar := range appConf.vars {
		// Set default
		appConf.vars[confKey].Value = confVar.DefaultValue
		// Check in environment
		if val := os.Getenv(confKey); val != "" {
			appConf.vars[confKey].Value = val
		}
	}

	return nil
}

/////////////////////////////////////////
// Helper Functions                   //
// for easy accessing configurations //
//////////////////////////////////////

// IsDebug returns true if debug mode is enabled.
func (appConf *AppConfig) IsDebug() bool {
	debug, _ := strconv.ParseBool(appConf.Get(constants.APP_DEBUG))
	return debug
}

// IsDev returns true if development environment is set.
func (appConf *AppConfig) IsDev() bool {
	return appConf.Get(constants.APP_ENV) == constants.ENV_DEV
}

// IsTest returns true if development environment is set.
func (appConf *AppConfig) IsTest() bool {
	return appConf.Get(constants.APP_ENV) == constants.ENV_TEST
}

// IsStaging returns true if development environment is set.
func (appConf *AppConfig) IsStaging() bool {
	return appConf.Get(constants.APP_ENV) == constants.ENV_STAGING
}

// IsAcceptance returns true if development environment is set.
func (appConf *AppConfig) IsAcceptance() bool {
	return appConf.Get(constants.APP_ENV) == constants.ENV_ACCEPTANCE
}

// IsProduction returns true if development environment is set.
func (appConf *AppConfig) IsProduction() bool {
	return appConf.Get(constants.APP_ENV) == constants.ENV_PRODUCTION
}

// LogrusLogLevel returns the logging level in logrus.Level format.
func (appConf *AppConfig) LogrusLogLevel() logrus.Level {
	level, err := logrus.ParseLevel(appConf.Get(constants.APP_LOG_LEVEL))
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

// LogLevel returns the logging level as a string.
func (appConf *AppConfig) LogLevel() string {
	return appConf.Get(constants.APP_LOG_LEVEL)
}

// Env returns the app's environment.
func (appConf *AppConfig) Env() string {
	return appConf.Get(constants.APP_ENV)
}

// Port returns the app's TCP/IP port.
func (appConf *AppConfig) Port() string {
	return appConf.Get(constants.APP_PORT)
}

// DBSecretName returns the app's database's secret's name.
func (appConf *AppConfig) DBSecretName() string {
	return appConf.Get(constants.APP_DB_SECRET_NAME)
}

// GetHostName returns the hostname of the machine where the app is running,
// if EC2_ID is set it will be returned instead. If neither can be found,
// "localhost" will be returned.
func GetHostName() string {
	if hostname := os.Getenv(constants.EC2_ID); hostname != "" {
		return hostname
	}
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "localhost"
}

// Hostname returns the hostname of the machine where the app is running,
// if EC2_ID is set it will be returned instead. If neither can be found,
// "localhost" will be returned.
func (appConf *AppConfig) Hostname() string {
	return GetHostName()
}

// Address returns a string in the format ":PORT"
func (appConf *AppConfig) Address() string {
	return fmt.Sprintf(":%s", appConf.Get(constants.APP_PORT))
}

// DumpTable creates a string table with all the config variable names,
// descriptions, constraints and default values
func (appConf *AppConfig) DumpTable() string {
	// Add the config variables to data in alphabetic order
	data := [][]string{}
	keys := []string{}
	for key := range appConf.vars {
		keys = append(keys, key)
	}
	// Sort is needed because maps always return values in random order
	sort.Strings(keys)
	for _, key := range keys {
		elem := appConf.vars[key]
		// Collect constraints
		constraints := []string{}
		for rule := range elem.Rules {
			constraints = append(constraints, rule)
		}
		// Sort is needed because maps always return values in random order
		sort.Strings(constraints)
		constraintList := strings.Join(constraints, ", ")
		data = append(data, []string{key, elem.Description, constraintList, elem.DefaultValue})
	}

	// Create the table
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Variable Name", "Description", "Constraints", "Default Value"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowSeparator("-")
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()

	return tableString.String()
}

// CreateSampleFile creates the .env.sample file based on the AppConfig variables with description and constraints.
func (appConf *AppConfig) CreateSampleFile(filename string) error {
	// Add the config variables to data in alphabetic order
	data := [][]string{}
	keys := []string{}
	for key := range appConf.vars {
		keys = append(keys, key)
	}
	// Sort is needed because maps always return values in random order
	sort.Strings(keys)
	for _, key := range keys {
		elem := appConf.vars[key]
		// Collect constraints
		constraints := []string{}
		for rule := range elem.Rules {
			constraints = append(constraints, rule)
		}
		// Sort is needed because maps always return values in random order
		sort.Strings(constraints)
		constraintList := strings.Join(constraints, ", ")
		data = append(data, []string{key, elem.DefaultValue, elem.Description, constraintList})
	}

	// Open the file for read and write, this will overwrite already existing files
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to create %s file", filename)
	}
	// Defer the proper closure fo the file
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			fmt.Printf("Failed to properly close %s file: %s\n", filename, err)
		}
	}(file)

	// Create a buffer
	datawriter := bufio.NewWriter(file)
	if _, err := datawriter.WriteString("# Automatically created by the application from the config object\n\n"); err != nil {
		return errors.Wrap(err, "Failed to write line into buffer")
	}
	for _, elem := range data {
		// Write description line
		_, err = datawriter.WriteString(fmt.Sprintf("# Description: %s # Constraints: %s\n", elem[2], elem[3]))
		if err != nil {
			return errors.Wrap(err, "Failed to write line into buffer")
		}
		// Write variable line
		_, err = datawriter.WriteString(fmt.Sprintf("%s=%s\n\n", elem[0], elem[1]))
		if err != nil {
			return errors.Wrap(err, "Failed to write line into buffer")
		}
	}
	// Flush the buffer into the file
	if err := datawriter.Flush(); err != nil {
		return errors.Wrapf(err, "Failed to write buffer into %s file", filename)
	}

	return nil
}
