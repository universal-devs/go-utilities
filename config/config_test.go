package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sirupsen/logrus"
	"github.com/universal-devs/go-utilities/constants"

	"github.com/stretchr/testify/suite"
)

///////////
// Suite //
///////////

// ConfigTestSuite extends testify's Suite.
type ConfigTestSuite struct {
	suite.Suite
}

// get default configs
func (cts *ConfigTestSuite) getDefaultConfigs() map[string]*Variable {
	return map[string]*Variable{
		constants.APP_PORT: {
			DefaultValue: "8080",
			Description:  "TCP/IP Port where the application listens",
			Rules: map[string]validation.Rule{
				"Required":   validation.Required,
				"Valid port": is.Port,
			},
		},
		constants.APP_ENV: {
			DefaultValue: constants.ENV_TEST,
			Description:  "The environment of the application",
			Rules: map[string]validation.Rule{
				"Required":          validation.Required,
				"Valid environment": validation.In(constants.ValidEnvironments...),
			},
		},
		constants.APP_DEBUG: {
			DefaultValue: "true",
			Description:  "Debug mode",
			Rules: map[string]validation.Rule{
				"Truthy value": validation.In(constants.TruthyValues...),
			},
		},
		constants.APP_LOG_LEVEL: {
			DefaultValue: constants.LOG_LEVEL_DEBUG,
			Description:  "Level of logging",
			Rules: map[string]validation.Rule{
				"Required":        validation.Required,
				"Valid log level": validation.In(constants.ValidLogLevels...),
			},
		},
		constants.APP_LOG_DEV: {
			Description: "Log development mode (Test formatter instead of JSON)",
			Rules: map[string]validation.Rule{
				"Truthy value": validation.In(constants.TruthyValues...),
			},
		},
		constants.APP_LOG_FORMAT_ERRORS: {
			Description: "Format error log entries by switching newlines to --- and tabs to spaces",
			Rules: map[string]validation.Rule{
				"Truthy value": validation.In(constants.TruthyValues...),
			},
		},
		constants.APP_DB_SECRET_NAME: {
			Description: "The Database's secret's name in AWS SecretsManager",
		},
	}
}

// clears the environment, and creates a new temporary envfile
func (cts *ConfigTestSuite) setupEnvTest(envs ...string) string {
	// Unset all config variables from the environment
	for _, name := range envs {
		cts.NoError(os.Unsetenv(name), "Environment variable should have been unset")
	}

	// Create empty envfile
	tmpfile, err := ioutil.TempFile(os.TempDir(), "devops-testing")
	cts.NoError(err, "Temp file should have been created")
	cts.NoError(tmpfile.Close(), "Temp file should have been closed")

	return tmpfile.Name()
}

func (cts *ConfigTestSuite) writeEnvfile(fileName string, in map[string]string) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	cts.NoError(err, "Envfile should have been opened for write")

	for key, val := range in {
		line := key + "=" + val + "\n"
		_, err := f.WriteString(line)
		cts.NoError(err, "Line should have been written to the envfile")
	}

	cts.NoError(f.Close(), "Envfile should have been closed")
}

func (cts *ConfigTestSuite) setEnvVars(in map[string]string) {
	for key, val := range in {
		cts.NoError(os.Setenv(key, val), "Environment variable should have been set")
	}
}

func (cts *ConfigTestSuite) hostname() string {
	name, err := os.Hostname()
	cts.NoError(err, "Hostname should be returned")
	return name
}

///////////
// TESTS //
///////////

func (cts *ConfigTestSuite) TestNew() {
	conf := NewConfig(nil)
	cts.NotNil(conf, "AppConfig should be created")
}

func (cts *ConfigTestSuite) TestDefaults() {
	envFile := cts.setupEnvTest(constants.BasicEnvs...)
	defer func(fileName string) {
		cts.NoErrorf(os.Remove(fileName), "Temp envfile (%s) should have been removed", fileName)
	}(envFile)

	conf := NewConfig(cts.getDefaultConfigs())

	cts.NoError(conf.loadEnv(), "Defaults and environment variables should have been loaded")
	cts.NoError(conf.Validate(), "The default configs should be valid")
	cts.Equalf(cts.hostname(), conf.Hostname(), "Hostname should return %s", cts.hostname())
	cts.Equal("8080", conf.Port(), "Port should return 8080")
	cts.Equal(":8080", conf.Address(), "The default address should be :8080")
	cts.Equal("test", conf.Env(), "The default environment should be dev")
	cts.True(conf.IsDebug(), "Debug mode should be enabled by default")
	cts.False(conf.IsDev(), "Environment is test IsDev should be false")
	cts.True(conf.IsTest(), "Environment is test IsTest should be true")
	cts.False(conf.IsStaging(), "Environment is test IsStaging should be false")
	cts.False(conf.IsAcceptance(), "Environment is test IsAcceptance should be false")
	cts.False(conf.IsProduction(), "Environment is test IsProduction should be false")
	cts.Equal("debug", conf.LogLevel(), "The default loglevel string should be debug")
	cts.Equal(logrus.DebugLevel, conf.LogrusLogLevel(), "The default loglevel should be logrus.DefaultLevel")
	cts.Equal("", conf.DBSecretName(), "Database secret name should be empty by default")
	val, ok := conf.Lookup("Som-var-which-is-not-set")
	cts.False(ok, "Lookup should return false on nonexisting variable name")
	cts.Empty(val, "On empty string should be returned on failed lookup")

	tab := conf.DumpTable()
	cts.Contains(tab, "The Database's secret's name in AWS SecretsManager")
	cts.Contains(tab, "TCP/IP Port where the application listens", "TCP Port where the application listens should be on the table")
}

func (cts *ConfigTestSuite) TestCreateSampleFile() {
	sampleFile := cts.setupEnvTest(constants.BasicEnvs...)
	cts.T().Logf("sampleFile: %s", sampleFile)
	conf := NewConfig(cts.getDefaultConfigs())
	defer func(fileName string) {
		cts.NoErrorf(os.Remove(fileName), "Temp sampleFile (%s) should have been removed", fileName)
	}(sampleFile)

	cts.NoError(conf.CreateSampleFile(sampleFile), "The sample file should have been created")

	content, err := ioutil.ReadFile(sampleFile)
	cts.NoError(err, "The sample file should be readable")
	for _, clue := range []string{
		"# Automatically created by the application from the config object",
		"# Description: Level of logging # Constraints: Required, Valid log level",
		"APP_LOG_LEVEL=debug",
		"# Description: Log development mode (Test formatter instead of JSON) # Constraints: Truthy value",
	} {
		cts.Containsf(string(content), clue, "The sample file should contain: %s", clue)
	}
}

func (cts *ConfigTestSuite) TestWrongEnvfile() {
	conf := NewConfig(cts.getDefaultConfigs())

	cts.EqualError(
		conf.Setup("AfileThatDoesNotExists"),
		//nolint:lll
		"Failed to set Application Configuration: Failed to overload variables with envfile(s): open AfileThatDoesNotExists: no such file or directory",
	)
}

func (cts *ConfigTestSuite) TestLoadConfig() {
	testCases := map[string]struct {
		defaults         map[string]*Variable
		envFile          map[string]string
		envVars          map[string]string
		validationErrors []string
		boolHelpers      map[string]bool
		stringHelpers    map[string]string
		logLvl           logrus.Level
	}{
		"Default configs": {
			defaults: cts.getDefaultConfigs(),
			boolHelpers: map[string]bool{
				"IsDebug":      true,
				"IsDev":        false,
				"IsTest":       true,
				"IsStaging":    false,
				"IsAcceptance": false,
				"IsProduction": false,
			},
			stringHelpers: map[string]string{
				"LogLevel": "debug",
				"Hostname": cts.hostname(),
				"Env":      "test",
				"Port":     "8080",
				"Address":  ":8080",
			},
			logLvl: logrus.DebugLevel,
		},
		"Debug off": {
			defaults: cts.getDefaultConfigs(),
			envFile: map[string]string{
				"APP_DEBUG": "off",
			},
			boolHelpers: map[string]bool{
				"IsDebug": false,
			},
			logLvl: logrus.DebugLevel,
		},
		"APP_PORT and APP_LOG_LEVEL in envfile": {
			defaults: cts.getDefaultConfigs(),
			envFile: map[string]string{
				"APP_PORT":      "9090",
				"APP_LOG_LEVEL": "error",
			},
			stringHelpers: map[string]string{
				"LogLevel": "error",
				"Address":  ":9090",
			},
			logLvl: logrus.ErrorLevel,
		},
		"APP_PORT and APP_LOG_LEVEL in environment": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_PORT":      "5050",
				"APP_LOG_LEVEL": "warn",
			},
			stringHelpers: map[string]string{
				"LogLevel": "warn",
				"Address":  ":5050",
			},
			logLvl: logrus.WarnLevel,
		},
		"APP_PORT and APP_LOG_LEVEL in both envfile and environment": {
			defaults: cts.getDefaultConfigs(),
			envFile: map[string]string{
				"APP_PORT":      "9090",
				"APP_LOG_LEVEL": "error",
			},
			envVars: map[string]string{
				"APP_PORT":      "5050",
				"APP_LOG_LEVEL": "warn",
			},
			stringHelpers: map[string]string{
				"LogLevel": "error",
				"Address":  ":9090",
			},
			logLvl: logrus.ErrorLevel,
		},
		"Invalid APP_PORT, not a number": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_PORT": "notAportNum",
			},
			validationErrors: []string{
				"PORT",
				"must be a valid port number",
			},
		},
		"Invalid APP_PORT, too low": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_PORT": "-1",
			},
			validationErrors: []string{
				"PORT",
				"must be a valid port number",
			},
		},
		"Invalid APP_PORT, too high": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_PORT": "999999",
			},
			validationErrors: []string{
				"PORT",
				"must be a valid port number",
			},
		},
		"Invalid APP_ENV": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_ENV": "Nasa",
			},
			validationErrors: []string{
				"Valid environment",
				"must be a valid value",
			},
		},
		"Invalid APP_LOG_LEVEL": {
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_LOG_LEVEL": "kernel_panic",
			},
			validationErrors: []string{
				"Valid log level",
				"must be a valid value",
			},
		},
		"EC2_IS is set": {
			logLvl:   logrus.DebugLevel,
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"EC2_ID": "i-asdf12345",
			},
			stringHelpers: map[string]string{
				"Hostname": "i-asdf12345",
			},
		},
		"APP_DB_SECRET_NAME is set": {
			logLvl:   logrus.DebugLevel,
			defaults: cts.getDefaultConfigs(),
			envVars: map[string]string{
				"APP_DB_SECRET_NAME": "super-secret-name",
			},
			stringHelpers: map[string]string{
				"DBSecretName": "super-secret-name",
			},
		},
	}

	for testCaseName, testCase := range testCases {
		cts.T().Logf("Configuration test: %s", testCaseName)

		envFile := cts.setupEnvTest(constants.BasicEnvs...)
		defer func(fileName string) {
			cts.NoErrorf(os.Remove(fileName), "Temp envfile (%s) should have been removed", fileName)
		}(envFile)
		cts.writeEnvfile(envFile, testCase.envFile)
		cts.setEnvVars(testCase.envVars)

		conf := NewConfig(testCase.defaults)

		errs := conf.Setup(envFile)

		if len(testCase.validationErrors) > 0 {
			cts.Error(errs, "Setup should have returned an error")
			for _, validationError := range testCase.validationErrors {
				cts.Contains(errs.Error(), validationError)
			}
			// skip the rest of the test
			continue
		}

		cts.Equalf(testCase.logLvl, conf.LogrusLogLevel(), "Expected log level: %d actual log level: %d", testCase.logLvl, conf.LogrusLogLevel())

		for key, val := range testCase.boolHelpers {
			switch key {
			case "IsDebug":
				cts.Equal(val, conf.IsDebug())
			case "IsDev":
				cts.Equal(val, conf.IsDev())
			case "IsTest":
				cts.Equal(val, conf.IsTest())
			case "IsStaging":
				cts.Equal(val, conf.IsStaging())
			case "IsAcceptance":
				cts.Equal(val, conf.IsAcceptance())
			case "IsProduction":
				cts.Equal(val, conf.IsProduction())
			}
		}

		for key, val := range testCase.stringHelpers {
			switch key {
			case "LogLevel":
				cts.Equal(val, conf.LogLevel())
			case "Hostname":
				cts.Equal(val, conf.Hostname())
			case "Env":
				cts.Equal(val, conf.Env())
			case "Port":
				cts.Equal(val, conf.Port())
			case "Address":
				cts.Equal(val, conf.Address())
			case "DBSecretName":
				cts.Equal(val, conf.DBSecretName())
			}
		}
	}
}

// TestConfig runs the whole test suite
func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
