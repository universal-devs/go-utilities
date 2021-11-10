package constants

const (
	APP_DEBUG = "APP_DEBUG"

	APP_LOG_LEVEL = "APP_LOG_LEVEL"

	APP_LOG_DEV = "APP_LOG_DEV"

	APP_ENV = "APP_ENV"

	APP_PORT = "APP_PORT"

	APP_DB_SECRET_NAME = "APP_DB_SECRET_NAME"

	APP_LOG_FORMAT_ERRORS = "APP_LOG_FORMAT_ERRORS"

	EC2_ID = "EC2_ID"
)

const (
	ENV_TEST = "test"

	ENV_DEV = "dev"

	ENV_STAGING = "stage"

	ENV_ACCEPTANCE = "acceptance"

	ENV_PRODUCTION = "production"
)

var (
	// ValidEnvironments are the valid environments of the application
	ValidEnvironments = []interface{}{
		ENV_DEV,
		ENV_TEST,
		ENV_STAGING,
		ENV_ACCEPTANCE,
		ENV_PRODUCTION,
	}
)

var (
	TruthyValues = []interface{}{"1", "t", "T", "TRUE", "true", "True", "0", "f", "F", "FALSE"}
)

var (
	// BasicEnvs is the list of environment variables all service should use.
	BasicEnvs = []string{
		EC2_ID,
		APP_ENV,
		APP_PORT,
		APP_LOG_LEVEL,
		APP_LOG_DEV,
		APP_LOG_FORMAT_ERRORS,
		APP_DEBUG,
		APP_DB_SECRET_NAME,
	}
)

// lib/pq ssl modes https://www.postgresql.org/docs/current/libpq-ssl.html
const (
	// SSL_MODE_DISABLE disables the checking of SSL.
	SSL_MODE_DISABLE = "disable"

	// SSL_MODE_ALLOW allows the checking of SSL.
	SSL_MODE_ALLOW = "allow"

	// SSL_MODE_PREFER prefers  the checking of SSL.
	SSL_MODE_PREFER = "prefer"

	// SSL_MODE_REQUIRE requires  the checking of SSL.
	SSL_MODE_REQUIRE = "require"

	// SSL_MODE_REQUIRE requires  the checking of SSL and verifies it with a CA.
	SSL_MODE_VERIFY_CA = "verify-ca"

	// SSL_MODE_REQUIRE requires  the checking of SSL and verifies it with a CA and validates if the domain exists.
	SSL_MODE_VERIFY_FULL = "verify-full"
)

var (
	// ValidSSLModes are the valid SSL modes. Used in validation.
	ValidSSLModes = []interface{}{
		SSL_MODE_DISABLE,
		SSL_MODE_ALLOW,
		SSL_MODE_PREFER,
		SSL_MODE_REQUIRE,
		SSL_MODE_VERIFY_CA,
		SSL_MODE_VERIFY_FULL,
	}
)
