package constants

const (
	LOG_LEVEL_ERROR = "error"

	LOG_LEVEL_WARN = "warn"

	LOG_LEVEL_INFO = "info"

	LOG_LEVEL_DEBUG = "debug"
)

var (
	// ValidLogLevels are the valid logging levels of the application
	ValidLogLevels = []interface{}{
		LOG_LEVEL_DEBUG,
		LOG_LEVEL_INFO,
		LOG_LEVEL_WARN,
		LOG_LEVEL_ERROR,
	}
)
