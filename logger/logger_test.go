package logger

import (
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
	"github.com/universal-devs/go-utilities/config"
	"github.com/universal-devs/go-utilities/constants"
)

// LoggerSuite extends testify's Suite.
type LoggerSuite struct {
	suite.Suite
}

func (ls *LoggerSuite) TestCreateLogger() {
	l := logrus.New()
	fields := logrus.Fields{
		"Field1": "Value1",
		"Field2": 23,
		"Field3": true,
	}
	log := NewLogger(l, fields)
	ls.NotNil(log, "New logger should be created")
	ls.Equal(fields, log.defaultFields, "Default field should have been set")
}

func (ls *LoggerSuite) TestNewCommonLoggerFromConfiguration() {
	hostname, err := os.Hostname()
	ls.NoError(err, "Hostname should be returned")

	conf := config.NewConfig(map[string]*config.Variable{
		constants.APP_LOG_LEVEL: {
			DefaultValue: constants.LOG_LEVEL_WARN,
		},
		constants.APP_ENV: {
			DefaultValue: constants.ENV_TEST,
		},
		constants.APP_LOG_DEV: {
			DefaultValue: "1",
		},
	})
	err = conf.Setup()
	ls.NoError(err, "Default configs should have been set up")

	commonLog := NewCommonLoggerFromConfiguration("test-service", "v1.2.3", conf)
	ls.NotNil(commonLog, "Common Logger should have been created")
	ls.Equal(logrus.Fields{
		"host":    hostname,
		"service": "test-service",
		"version": "v1.2.3",
		"env":     constants.ENV_TEST,
	}, commonLog.defaultFields, "Default field should have been set")
}

func (ls *LoggerSuite) TestCreateCommonLogger() {
	fields := logrus.Fields{
		"service": "test-service",
		"version": "v1.2.3",
		"env":     "test",
		"host":    "docker",
	}
	commonLog := NewCommonLogger("test-service", "v1.2.3", "test", "docker", false)
	ls.NotNil(commonLog, "Common Logger should have been created")
	ls.Equal(fields, commonLog.defaultFields, "Default field should have been set")

	commonLog = NewCommonLogger("test-service", "v1.2.3", "test", "docker", true)
	ls.NotNil(commonLog, "Common Logger should have been created")
	ls.Equal(fields, commonLog.defaultFields, "Default field should have been set")
}

func (ls *LoggerSuite) TestCreateComponentLogger() {
	l := logrus.New()
	fields := logrus.Fields{
		"Field1": "Value1",
		"Field2": 23,
		"Field3": true,
	}

	log := NewLogger(l, fields)
	ls.NotNil(log, "New logger should be created")
	ls.Equal(fields, log.defaultFields, "Default field should have been set")
	componentLog := log.NewComponentLogger("test-device")
	ls.NotNil(componentLog, "Common Logger should have been created")
	fields["component"] = "test-device"
	ls.Equal(fields, componentLog.defaultFields, "Default field should have been set")
}

func (ls *LoggerSuite) TestExtraField() {
	nullLogger, hook := logrusTest.NewNullLogger()
	fields := logrus.Fields{
		"Field1": "Value1",
		"Field2": 23,
		"Field3": true,
	}
	testLogger := NewLogger(nullLogger, fields)
	testLogger.WithField("extra-field", "extra-value").Info("Info msg")
	fields["extra-field"] = "extra-value"
	ls.Equal(fields, hook.LastEntry().Data, "Extra field should have been added to the log entry")
}

func (ls *LoggerSuite) TestExtraFields() {
	nullLogger, hook := logrusTest.NewNullLogger()
	fields := logrus.Fields{
		"Field1": "Value1",
		"Field2": 23,
		"Field3": true,
	}
	testLogger := NewLogger(nullLogger, fields)
	testLogger.WithFields(logrus.Fields{
		"extra-field1": "extra-value1",
		"extra-field2": 42,
		"extra-field3": false,
	}).Info("Info msg")
	fields["extra-field1"] = "extra-value1"
	fields["extra-field2"] = 42
	fields["extra-field3"] = false
	ls.Equal(fields, hook.LastEntry().Data, "Extra field should have been added to the log entry")
}

func (ls *LoggerSuite) TestWithError() {
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	err := errors.New("Test error")
	testLogger.WithError(err).Error("Something went wrong")
	ls.Equal("Test error", hook.LastEntry().Data["error"], "error field should have been added to the log entry")
}

func getError() error {
	return errors.New("Test Error")
}

func getSomething() error {
	return errors.Wrap(getError(), "Cannot get something")
}

func (ls *LoggerSuite) TestWithErrorAndStack() {
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	err := getSomething()
	testLogger.WithError(err).Error("Something went wrong")
	ls.Contains(
		hook.LastEntry().Data["error"],
		"github.com/universal-devs/go-utilities/logger.getError",
		"error field should have been added to the log entry")
}

func (ls *LoggerSuite) TestWithErrorAndStack_unpacked() {
	ls.NoError(os.Setenv("LOG_FORMAT_ERRORS", "true"), "LOG_FORMAT_ERRORS should have been set to true")
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	err := getSomething()
	testLogger.WithError(err).Error("Something went wrong")
	ls.Contains(
		hook.LastEntry().Data["error"],
		"Test Error --- github.com/universal-devs/go-utilities/logger.getError",
		"error field should have been added to the log entry")
}

func (ls *LoggerSuite) TestNilError() {
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	testLogger.WithError(nil).Error("Something went wrong")
	ls.Equal("<nil>", hook.LastEntry().Data["error"], "<nil> should be returned")
}

func (ls *LoggerSuite) TestEntry() {
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	entry := testLogger.Entry()
	ls.NotNil(entry, "A new log Entry should be created")
	entry.Error("Something went wrong")
	ls.Equal("Something went wrong", hook.LastEntry().Message, "Entry should have been written")
	ls.Equal(logrus.ErrorLevel, hook.LastEntry().Level, "The level of the log entry should be error")
}

func (ls *LoggerSuite) TestGormLogger() {
	nullLogger, hook := logrusTest.NewNullLogger()
	testLogger := NewLogger(nullLogger, nil)
	gormLog := testLogger.NewGormLogger("GORM")

	gormLog.Info(context.TODO(), "GORM Info")
	ls.Contains(hook.LastEntry().Message, "GORM Info", "Entry should have been written")
	ls.Equal(logrus.InfoLevel, hook.LastEntry().Level, "The level of the log entry should be info")

	gormLog.Warn(context.TODO(), "GORM Warn")
	ls.Contains(hook.LastEntry().Message, "GORM Warn", "Entry should have been written")
	ls.Equal(logrus.WarnLevel, hook.LastEntry().Level, "The level of the log entry should be warn")

	gormLog.Error(context.TODO(), "GORM Error")
	ls.Contains(hook.LastEntry().Message, "GORM Error", "Entry should have been written")
	ls.Equal(logrus.ErrorLevel, hook.LastEntry().Level, "The level of the log entry should be error")
}

// TestLogger runs the suite
func TestLogger(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
