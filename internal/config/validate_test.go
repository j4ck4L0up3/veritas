package config

import (
	"fmt"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compileOrDie compiles the embedded schema or fails the test.
func compileOrDie(t *testing.T) *jsonschema.Schema {
	t.Helper()
	sch, err := compileSchema()
	require.NoError(t, err)
	require.NotNil(t, sch)
	return sch
}

// mustValidate compiles the schema and runs validate against the given YAML,
// returning the validation error (or nil) for the current test.
func mustValidate(t *testing.T, yamlText string) error {
	t.Helper()
	sch := compileOrDie(t)
	return validate(sch, []byte(yamlText))
}

// validFullConfig is a schema-valid config exercising every property.
const validFullConfig = `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_level: "info"
log_format: "json"
locations:
  db: "/data/veritas.db"
  bin: "/usr/local/bin/veritas"
  env: "/etc/veritas/env"
  logs: "/var/log/veritas"
  blobs: "/data/blobs"
  uploads: "/data/uploads"
  service: "$HOME/.config/systemd/user/veritas.service"
`

// ---------------------------------------------------------------------------
// compileSchema
// ---------------------------------------------------------------------------

func Test_compileSchema_ReturnsCompiledSchema(t *testing.T) {
	// Act
	sch, err := compileSchema()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sch)
}

// ---------------------------------------------------------------------------
// validate — valid configs
// ---------------------------------------------------------------------------

func Test_validate_ValidFullConfig(t *testing.T) {
	// Act
	err := mustValidate(t, validFullConfig)

	// Assert
	assert.NoError(t, err)
}

func Test_validate_MinimalConfig_OnlyRequiredKeys(t *testing.T) {
	// Arrange — log_level/log_format are optional
	yamlText := `
host: "192.168.1.10"
port: 1
upload_ttl: 1
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	assert.NoError(t, err)
}

func Test_validate_Ipv6Host(t *testing.T) {
	// Arrange
	yamlText := `
host: "fd00::1"
port: 49151
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	assert.NoError(t, err)
}

func Test_validate_AllowedLogLevels(t *testing.T) {
	// Arrange
	base := `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_level: %q
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act + Assert
	for _, level := range []string{"debug", "info", "warn", "error", "fatal"} {
		err := mustValidate(t, fmt.Sprintf(base, level))
		assert.NoError(t, err, "log_level %q should be accepted", level)
	}
}

func Test_validate_AllowedLogFormats(t *testing.T) {
	// Arrange
	base := `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_format: %q
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act + Assert
	for _, format := range []string{"text", "json"} {
		err := mustValidate(t, fmt.Sprintf(base, format))
		assert.NoError(t, err, "log_format %q should be accepted", format)
	}
}

func Test_validate_PortBoundaries_1And49151(t *testing.T) {
	// Arrange — both port extremes are inclusive
	validFor := `
host: "127.0.0.1"
port: %d
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act + Assert
	assert.NoError(t, mustValidate(t, fmt.Sprintf(validFor, 1)))
	assert.NoError(t, mustValidate(t, fmt.Sprintf(validFor, 49151)))
}

// ---------------------------------------------------------------------------
// validate — malformed input
// ---------------------------------------------------------------------------

func Test_validate_MalformedYaml_ReturnsError(t *testing.T) {
	// Act
	err := mustValidate(t, "host: [127.0.0.1\n  bad indent::")

	// Assert
	require.Error(t, err)
}

func Test_validate_EmptyYaml_MissingRequiredKeys(t *testing.T) {
	// Act
	err := mustValidate(t, "")

	// Assert
	require.Error(t, err)
}

func Test_validate_ScalarInsteadOfObject(t *testing.T) {
	// Act
	err := mustValidate(t, "just-a-string")

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — host
// ---------------------------------------------------------------------------

func Test_validate_Host_Missing_TooFewDigits(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.1"
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Host_NotAnIP(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "localhost"
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Host_QuotedNumericOctetOutOfRange(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.256"
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Host_Number(t *testing.T) {
	// Arrange — unquoted "123.45" is parsed by YAML as a float, not a string
	yamlText := `
host: 123.45
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — port
// ---------------------------------------------------------------------------

func Test_validate_Port_Zero(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 0
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Port_49152_AboveMaximum(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 49152
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Port_Negative(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: -1
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Port_String(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: "9001"
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — upload_ttl
// ---------------------------------------------------------------------------

func Test_validate_UploadTTL_Zero(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 0
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_UploadTTL_Negative(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: -5
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_UploadTTL_String(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: "86400"
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — locations
// ---------------------------------------------------------------------------

func Test_validate_Locations_MissingKey(t *testing.T) {
	// Arrange — no "blobs" key
	yamlText := `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

func Test_validate_Locations_EmptyStringPath(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
locations:
  db: ""
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_Locations_NonStringPath(t *testing.T) {
	// Arrange — unquoted "123" is parsed by YAML as an integer
	yamlText := `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
locations:
  db: 123
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

func Test_validate_Locations_UnknownKey(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
  extra: "/nope"
`)

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — top-level structure
// ---------------------------------------------------------------------------

func Test_validate_MissingHost(t *testing.T) {
	// Arrange
	yamlText := `
port: 9001
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

func Test_validate_MissingPort(t *testing.T) {
	// Arrange
	yamlText := `
host: "127.0.0.1"
upload_ttl: 86400
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

func Test_validate_MissingUploadTTL(t *testing.T) {
	// Arrange
	yamlText := `
host: "127.0.0.1"
port: 9001
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`

	// Act
	err := mustValidate(t, yamlText)

	// Assert
	require.Error(t, err)
}

func Test_validate_MissingLocations(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_UnknownTopLevelProperty(t *testing.T) {
	// Act
	err := mustValidate(t, validFullConfig+"\nsneaky: true\n")

	// Assert
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// validate — log_level / log_format
// ---------------------------------------------------------------------------

func Test_validate_LogLevel_Invalid(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_level: "verbose"
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_LogLevel_WrongType(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_level: 3
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_LogFormat_Invalid(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_format: "yaml"
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert
	require.Error(t, err)
}

func Test_validate_LogLevel_CaseSensitive(t *testing.T) {
	// Act
	err := mustValidate(t, `
host: "127.0.0.1"
port: 9001
upload_ttl: 86400
log_level: "INFO"
locations:
  db: "/db"
  bin: "/bin"
  env: "/env"
  logs: "/logs"
  blobs: "/blobs"
  uploads: "/uploads"
  service: "/service"
`)

	// Assert — "INFO" is not one of the const values
	require.Error(t, err)
}
