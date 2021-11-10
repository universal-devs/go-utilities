// Package logger provides a wrapper around Sirupsen's Logrus with mandatory default fields
// Use the NewCommonLogger constructor to create your application's logger
// Use the NewComponentLogger method to create child loggers for components of your application
// Use Entry WithField WithFields and WithError to create new log entries
package logger

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/universal-devs/go-utilities/constants"
	gormLog "gorm.io/gorm/logger"
)

// BasicJSONFormatter is the basic json log format
var BasicJSONFormatter = &logrus.JSONFormatter{
	TimestampFormat:  time.RFC3339,
	DisableTimestamp: false,
}

// BasicTextFormatter is the basic json log format
var BasicTextFormatter = &logrus.TextFormatter{
	TimestampFormat:  time.RFC3339,
	DisableTimestamp: false,
	FullTimestamp:    true,
}

// inBrackets matches everything inside [brackets]
var inBrackets = regexp.MustCompile(`\[([^]]+)\]`)

// Logger is a wrapper around Logrus FieldLogger with default fields
type Logger struct {
	log           logrus.FieldLogger
	defaultFields logrus.Fields
	formatErrors  bool
	gormConf      *gormLog.Config
}

// NewLogger creates a new logger instance with the supplied Logrus FieldLogger and default fields
func NewLogger(log logrus.FieldLogger, defaultFields logrus.Fields) *Logger {
	return &Logger{
		log:           log,
		defaultFields: defaultFields,
		formatErrors:  isFormatErrors(),
		gormConf: &gormLog.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormLog.Info,
		},
	}
}

func getLogLevel(debug bool) logrus.Level {
	if debug {
		return logrus.DebugLevel
	}
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return logrus.InfoLevel
	}
	return level
}

func isDevLog() bool {
	devLog, _ := strconv.ParseBool(os.Getenv("LOG_DEV"))
	return devLog
}

func isFormatErrors() bool {
	devLog, _ := strconv.ParseBool(os.Getenv("LOG_FORMAT_ERRORS"))
	return devLog
}

// NewCommonLogger (DEPRECATED!, use NewCommonLoggerFromConfiguration instead!)
// service name
// semantic version (+ commit hash)
// environment
// host (EC2 Identifier)
func NewCommonLogger(service, version, env, host string, debug bool) *Logger {
	log := logrus.New()
	log.SetLevel(getLogLevel(debug))
	log.SetOutput(os.Stdout)
	log.SetReportCaller(debug)
	if isDevLog() {
		log.SetFormatter(BasicTextFormatter)
	} else {
		log.SetFormatter(BasicJSONFormatter)
	}
	return NewLogger(log, logrus.Fields{
		"service": service,
		"version": version,
		"env":     env,
		"host":    host,
	})
}

// configGetter is a structure that can get config items by name and tell the hostname
type configGetter interface {
	Get(string) string
	Hostname() string
}

// NewCommonLoggerFromConfiguration is the prefferred way to create the Common Logger
func NewCommonLoggerFromConfiguration(serviceName, serviceVersion string, config configGetter) *Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	ok, _ := strconv.ParseBool(config.Get(constants.APP_DEBUG))
	log.SetReportCaller(ok)

	level, err := logrus.ParseLevel(config.Get(constants.APP_LOG_LEVEL))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	log.SetFormatter(BasicJSONFormatter)
	if ok, _ := strconv.ParseBool(config.Get(constants.APP_LOG_DEV)); ok {
		log.SetFormatter(BasicTextFormatter)
	}

	commonLog := NewLogger(log, logrus.Fields{
		"service": serviceName,
		"version": serviceVersion,
		"env":     config.Get(constants.APP_ENV),
		"host":    config.Hostname(),
	})

	switch level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		commonLog.gormConf.LogLevel = gormLog.Error
	case logrus.WarnLevel:
		commonLog.gormConf.LogLevel = gormLog.Warn
	}

	return commonLog
}

// NewComponentLogger creates a new logger with the loggers default FieldLogger and fields
// and adds a new field 'component' with the supplied componentName.
func (l *Logger) NewComponentLogger(componentName string) *Logger {
	newFields := logrus.Fields{}
	for key, value := range l.defaultFields {
		newFields[key] = value
	}
	newFields["component"] = componentName
	newLogger := NewLogger(l.log, newFields)
	newLogger.gormConf.SlowThreshold = l.gormConf.SlowThreshold
	newLogger.gormConf.LogLevel = l.gormConf.LogLevel
	return newLogger
}

// Entry creates a new log entry with the default fields
// Call .Info .Warn .Error etc. on this Entry
func (l *Logger) Entry() *logrus.Entry {
	return l.log.WithFields(l.defaultFields)
}

// WithField adds an extra field to the default fields
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.log.WithFields(l.defaultFields).WithField(key, value)
}

// WithFields adds a map of fields to the default fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.log.WithFields(l.defaultFields).WithFields(fields)
}

// WithError adds a new field with key "error" and value is the parsed version of the supplied error object
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.log.WithFields(l.defaultFields).WithField("error", l.parseError(err))
}

// parseError tries to unwrap the underlying pkg/errors.Error, and return it as a string.
// If the error cannot be unwrapped the original error string will be returned.
// A nil error will produce "<nil>" string.
// If error formatting is enabled by the LOG_FORMAT_ERROR env var the newlines will be replaced by ---
// and the tabs will be replaced by spaces.
func (l *Logger) parseError(err error) string {
	if err == nil {
		return "<nil>"
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		return err.Error()
	}
	if l.formatErrors {
		str := fmt.Sprintf("%+v", unwrapped)
		re := regexp.MustCompile(`\r?\n`)
		str = re.ReplaceAllString(str, " --- ")
		return strings.ReplaceAll(str, "\t", "")
	}
	return fmt.Sprintf("%+v", unwrapped)
}

func getLogLevelFromGormMsg(msg string) string {
	matches := inBrackets.FindAllStringSubmatch(msg, -1)
	if len(matches) > 0 {
		if len(matches[0]) > 1 {
			switch matches[0][1] {
			case "error":
				return constants.LOG_LEVEL_ERROR
			case "warn":
				return constants.LOG_LEVEL_WARN
			}
		}
	}
	return constants.LOG_LEVEL_INFO
}

// Printf implements the gorm/logger.Writer interface
func (l *Logger) Printf(format string, args ...interface{}) {
	switch getLogLevelFromGormMsg(format) {
	case constants.LOG_LEVEL_ERROR:
		l.Entry().Errorf(strings.Replace(format, "[error] ", "", 1), args...)
	case constants.LOG_LEVEL_WARN:
		l.Entry().Warnf(strings.Replace(format, "[warn] ", "", 1), args...)
	default:
		l.Entry().Infof(strings.Replace(format, "[info] ", "", 1), args...)
	}
}

// NewGormLogger creates a gorm/logger.Interface from the CommonLogger
func (l *Logger) NewGormLogger(componentName string) gormLog.Interface {
	gormLogger := l.NewComponentLogger(componentName)
	return gormLog.New(gormLogger, *gormLogger.gormConf)
}
