# go-utilities

This is a collection of handy packages and utilities which can be imported into various Go Applications.

## Usage
---

### [Config](config)
The config package provides configuration primitives and the AppConfig object, which can be used to load, validate and retrieve configuration items

---
### [Constants](constants)
The constants package provides constant values that all application should use. These are mainly environment variable names

---
### [Logger](logger)
The logger package provides a common logger which should be used by all services. It requires the service-name, version, environment and hostname to be set. These fields will be added to all log entries. In debug mode every log entry will contain the caller function with filename and line-number.

Use ```github.com/pkg/errors``` to wrap and propagate errors in your application. Use the logger's WithError method to log errors from the application (this will allow the unwrapping of errors, with correct error-trace)

---
