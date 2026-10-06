package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestConfig writes yaml contents to a unique file in a temp dir and
// returns the file path.
func writeTestConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

// validYaml is a schema-valid config body with plain (non-env-expanded) paths.
func validYaml(port int) string {
	return `
host: "127.0.0.1"
port: ` + strconv.Itoa(port) + `
upload_ttl: 86400
log_level: "info"
log_format: "text"
locations:
  db: "/tmp/veritas-test/db"
  bin: "/tmp/veritas-test/bin"
  env: "/tmp/veritas-test/env"
  logs: "/tmp/veritas-test/logs"
  blobs: "/tmp/veritas-test/blobs"
  uploads: "/tmp/veritas-test/uploads"
  service: "/tmp/veritas-test/service"
`
}

// unsetOverrideVars ensures the VERITAS_* override env vars are cleared for a
// test (restoring previous values after the test finishes).
func unsetOverrideVars(t *testing.T, vars ...string) {
	t.Helper()
	for _, v := range vars {
		t.Setenv(v, "")
	}
}

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

func TestLoad_ValidConfigFile(t *testing.T) {
	unsetOverrideVars(t,
		VERITAS_DB_PATH, VERITAS_BIN_PATH, VERITAS_ENV_PATH, VERITAS_LOG_PATH,
		VERITAS_BLOB_PATH, VERITAS_UPLOAD_PATH, VERITAS_SERVICE_PATH,
		VERITAS_HOST, VERITAS_PORT, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, validYaml(9001))
	cfg, err := Load(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, uint16(9001), cfg.Port)
	assert.Equal(t, uint64(86400), cfg.UploadTTL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.Equal(t, "/tmp/veritas-test/db", cfg.Locations.DbPath)
	assert.Equal(t, "/tmp/veritas-test/uploads", cfg.Locations.UploadPath)
}

func TestLoad_EmptyCmdPath_CreatesDefaultConfig(t *testing.T) {
	unsetOverrideVars(t,
		VERITAS_DB_PATH, VERITAS_BIN_PATH, VERITAS_ENV_PATH, VERITAS_LOG_PATH,
		VERITAS_BLOB_PATH, VERITAS_UPLOAD_PATH, VERITAS_SERVICE_PATH,
		VERITAS_HOST, VERITAS_PORT, VERITAS_UPLOAD_TTL)

	// point the default config dir at a temp dir
	dir := t.TempDir()
	t.Setenv("VERITAS_CONFIG_PATH", dir)

	// $HOME expansion happens at parse time; default paths all use $HOME
	home, ok := os.LookupEnv("HOME")
	require.True(t, ok, "HOME must be set for default config paths")

	cfg, err := Load("")

	require.NoError(t, err)
	require.NotNil(t, cfg)
	// default yaml values from defaultYamlConfig
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, uint16(9001), cfg.Port)
	assert.Equal(t, uint64(86400), cfg.UploadTTL)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.Equal(t, filepath.Join(home, ".local/share/veritas"), cfg.Locations.DbPath)
	assert.Equal(t, filepath.Join(home, ".config/systemd/user"), cfg.Locations.ServicePath)
	// the default config file was actually written
	_, err = os.Stat(filepath.Join(dir, "config.yaml"))
	assert.NoError(t, err)
}

func TestLoad_MissingFile(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	cfg, err := Load(path)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidYaml(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, "host: [unterminated\n")

	cfg, err := Load(path)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_SchemaViolation_MissingRequiredPort(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, `
host: "127.0.0.1"
upload_ttl: 86400
locations:
  db: "/tmp/veritas-test/db"
  bin: "/tmp/veritas-test/bin"
  env: "/tmp/veritas-test/env"
  logs: "/tmp/veritas-test/logs"
  blobs: "/tmp/veritas-test/blobs"
  uploads: "/tmp/veritas-test/uploads"
  service: "/tmp/veritas-test/service"
`)

	cfg, err := Load(path)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_SchemaViolation_InvalidHost(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, strings.Replace(validYaml(9001), `host: "127.0.0.1"`, `host: "not-an-ip"`, 1))

	cfg, err := Load(path)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_SchemaViolation_PortAboveMaximum(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	// 60001 > schema maximum of 49151
	path := writeTestConfig(t, validYaml(60001))

	cfg, err := Load(path)

	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_EnvPathExpansion(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)

	yaml := ``
	yaml += "host: \"127.0.0.1\"\n"
	yaml += "port: 9001\n"
	yaml += "upload_ttl: 86400\n"
	yaml += "locations:\n"
	yaml += "  db: \"$VERITAS_TEST_DIR/db\"\n"
	yaml += "  bin: \"$VERITAS_TEST_DIR/bin\"\n"
	yaml += "  env: \"$VERITAS_TEST_DIR/env\"\n"
	yaml += "  logs: \"$VERITAS_TEST_DIR/logs\"\n"
	yaml += "  blobs: \"$VERITAS_TEST_DIR/blobs\"\n"
	yaml += "  uploads: \"$VERITAS_TEST_DIR/uploads\"\n"
	yaml += "  service: \"$VERITAS_TEST_DIR/service\"\n"
	path := writeTestConfig(t, yaml)

	cfg, err := Load(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, filepath.Join(dir, "db"), cfg.Locations.DbPath)
}

// ---------------------------------------------------------------------------
// load
// ---------------------------------------------------------------------------

func Test_load_ValidFile(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, validYaml(8080))

	var cfg Config
	err := load(path, &cfg)

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, uint16(8080), cfg.Port)
	assert.Equal(t, uint64(86400), cfg.UploadTTL)
}

func Test_load_NonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.yaml")

	var cfg Config
	err := load(path, &cfg)

	assert.ErrorIs(t, err, os.ErrNotExist)
}

func Test_load_InvalidYaml(t *testing.T) {
	path := writeTestConfig(t, "port: \"not-a-number\"\nhost: [")

	var cfg Config
	err := load(path, &cfg)

	assert.Error(t, err)
}

func Test_load_SchemaViolation_PortBelowMinimum(t *testing.T) {
	path := writeTestConfig(t, validYaml(0)) // schema minimum is 1

	var cfg Config
	err := load(path, &cfg)

	assert.Error(t, err)
}

func Test_load_EnvVarPathExpansion(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)

	yaml := "host: \"127.0.0.1\"\nport: 9001\nupload_ttl: 86400\nlocations:\n"
	yaml += "  db: \"$VERITAS_TEST_DIR/db\"\n  bin: \"$VERITAS_TEST_DIR/b\"\n"
	yaml += "  env: \"$VERITAS_TEST_DIR/e\"\n  logs: \"$VERITAS_TEST_DIR/l\"\n"
	yaml += "  blobs: \"$VERITAS_TEST_DIR/bl\"\n  uploads: \"$VERITAS_TEST_DIR/u\"\n"
	yaml += "  service: \"$VERITAS_TEST_DIR/s\"\n"
	path := writeTestConfig(t, yaml)

	var cfg Config
	err := load(path, &cfg)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "db"), cfg.Locations.DbPath)
	assert.Equal(t, filepath.Join(dir, "b"), cfg.Locations.BinPath)
	assert.Equal(t, filepath.Join(dir, "s"), cfg.Locations.ServicePath)
}

func Test_load_ProtectedPortBelow1024_WarnsButSucceeds(t *testing.T) {
	// Schema allows port >= 1; verifiedPort warns (ErrConflictingPort) on
	// < 1024 but load() must still succeed, keeping the port.
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, validYaml(80))

	var cfg Config
	err := load(path, &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint16(80), cfg.Port)
}

func Test_load_PortBoundary1024_NoConflictWarning(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	path := writeTestConfig(t, validYaml(1024))

	var cfg Config
	err := load(path, &cfg)

	require.NoError(t, err)
	assert.Equal(t, uint16(1024), cfg.Port)
}

func Test_load_MissingRequiredKeysInFile(t *testing.T) {
	path := writeTestConfig(t, `
host: "127.0.0.1"
port: 9001
`)
	// locations + upload_ttl are required by schema

	var cfg Config
	err := load(path, &cfg)

	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// setOverrides
// ---------------------------------------------------------------------------

func Test_setOverrides_NoEnvSet_Identity(t *testing.T) {
	unsetOverrideVars(t,
		VERITAS_DB_PATH, VERITAS_BIN_PATH, VERITAS_ENV_PATH, VERITAS_LOG_PATH,
		VERITAS_BLOB_PATH, VERITAS_UPLOAD_PATH, VERITAS_SERVICE_PATH,
		VERITAS_HOST, VERITAS_PORT, VERITAS_UPLOAD_TTL)

	cfg := &Config{
		Host:      "10.0.0.2",
		Port:      9002,
		UploadTTL: 1234,
	}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2", cfg.Host)
	assert.Equal(t, uint16(9002), cfg.Port)
	assert.Equal(t, uint64(1234), cfg.UploadTTL)
}

func Test_setOverrides_PathVariables(t *testing.T) {
	unsetOverrideVars(t,
		VERITAS_DB_PATH, VERITAS_BIN_PATH, VERITAS_ENV_PATH, VERITAS_LOG_PATH,
		VERITAS_BLOB_PATH, VERITAS_UPLOAD_PATH, VERITAS_SERVICE_PATH,
		VERITAS_HOST, VERITAS_PORT, VERITAS_UPLOAD_TTL)

	dir := t.TempDir()
	t.Setenv(VERITAS_DB_PATH, dir)
	t.Setenv(VERITAS_LOG_PATH, filepath.Join(dir, "logs"))

	cfg := &Config{
		Locations: Locations{
			DbPath:  "/original/db",
			LogPath: "/original/logs",
		},
	}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, dir, cfg.Locations.DbPath)
	assert.Equal(t, filepath.Join(dir, "logs"), cfg.Locations.LogPath)
	// untouched fields stay as-is
	assert.Equal(t, "", cfg.Locations.BinPath)
}

func Test_setOverrides_EnvPathExpansionInOverride(t *testing.T) {
	unsetOverrideVars(t, VERITAS_DB_PATH, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)

	t.Setenv(VERITAS_DB_PATH, "$VERITAS_TEST_DIR/db")

	cfg := &Config{}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "db"), cfg.Locations.DbPath)
}

func Test_setOverrides_Host(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_HOST, "10.1.2.3")

	cfg := &Config{Host: "127.0.0.1"}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, "10.1.2.3", cfg.Host)
}

func Test_setOverrides_HostTakesEffectOverYaml(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)
	t.Setenv(VERITAS_HOST, "192.168.1.10")

	yaml := "host: \"127.0.0.1\"\nport: 9001\nupload_ttl: 86400\nlocations:\n"
	yaml += "  db: \"$VERITAS_TEST_DIR/db\"\n  bin: \"$VERITAS_TEST_DIR/b\"\n"
	yaml += "  env: \"$VERITAS_TEST_DIR/e\"\n  logs: \"$VERITAS_TEST_DIR/l\"\n"
	yaml += "  blobs: \"$VERITAS_TEST_DIR/bl\"\n  uploads: \"$VERITAS_TEST_DIR/u\"\n"
	yaml += "  service: \"$VERITAS_TEST_DIR/s\"\n"
	path := writeTestConfig(t, yaml)

	cfg, err := Load(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "192.168.1.10", cfg.Host)
}

func Test_setOverrides_ValidPort(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_PORT, "9099")

	cfg := &Config{Port: 9001}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, uint16(9099), cfg.Port)
}

func Test_setOverrides_ConflictingPortBelow1024_KeptWithWarning(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_PORT, "53")

	cfg := &Config{Port: 9001}

	// ErrConflictingPort is a logged warning, not a failure
	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, uint16(53), cfg.Port)
}

func Test_setOverrides_NonIntegerPort(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_PORT, "not-a-number")

	cfg := &Config{Port: 9001}

	err := setOverrides(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-integer port")
	// original port unchanged
	assert.Equal(t, uint16(9001), cfg.Port)
}

func Test_setOverrides_NegativePort(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_PORT, "-5")

	cfg := &Config{Port: 9001}

	err := setOverrides(cfg)

	assert.ErrorIs(t, err, ErrInvalidPort)
	assert.Equal(t, uint16(9001), cfg.Port)
}

func Test_setOverrides_PortAbove65535(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_PORT, "70000")

	cfg := &Config{Port: 9001}

	err := setOverrides(cfg)

	assert.ErrorIs(t, err, ErrInvalidPort)
	assert.Equal(t, uint16(9001), cfg.Port)
}

func Test_setOverrides_UploadTTL(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_UPLOAD_TTL, "3600")

	cfg := &Config{UploadTTL: 86400}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, uint64(3600), cfg.UploadTTL)
}

func Test_setOverrides_NonIntegerTTL(t *testing.T) {
	unsetOverrideVars(t, VERITAS_PORT, VERITAS_HOST, VERITAS_UPLOAD_TTL)

	t.Setenv(VERITAS_UPLOAD_TTL, "banana")

	cfg := &Config{UploadTTL: 86400}

	err := setOverrides(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-integer ttl")
	assert.Equal(t, uint64(86400), cfg.UploadTTL)
}

func Test_setOverrides_AllPathVariablesTogether(t *testing.T) {
	unsetOverrideVars(t,
		VERITAS_DB_PATH, VERITAS_BIN_PATH, VERITAS_ENV_PATH, VERITAS_LOG_PATH,
		VERITAS_BLOB_PATH, VERITAS_UPLOAD_PATH, VERITAS_SERVICE_PATH,
		VERITAS_HOST, VERITAS_PORT, VERITAS_UPLOAD_TTL)

	dir := t.TempDir()
	t.Setenv(VERITAS_DB_PATH, filepath.Join(dir, "db"))
	t.Setenv(VERITAS_BIN_PATH, filepath.Join(dir, "bin"))
	t.Setenv(VERITAS_ENV_PATH, filepath.Join(dir, "env"))
	t.Setenv(VERITAS_LOG_PATH, filepath.Join(dir, "logs"))
	t.Setenv(VERITAS_BLOB_PATH, filepath.Join(dir, "blobs"))
	t.Setenv(VERITAS_UPLOAD_PATH, filepath.Join(dir, "uploads"))
	t.Setenv(VERITAS_SERVICE_PATH, filepath.Join(dir, "service"))

	cfg := &Config{}

	err := setOverrides(cfg)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "db"), cfg.Locations.DbPath)
	assert.Equal(t, filepath.Join(dir, "bin"), cfg.Locations.BinPath)
	assert.Equal(t, filepath.Join(dir, "env"), cfg.Locations.EnvPath)
	assert.Equal(t, filepath.Join(dir, "logs"), cfg.Locations.LogPath)
	assert.Equal(t, filepath.Join(dir, "blobs"), cfg.Locations.BlobPath)
	assert.Equal(t, filepath.Join(dir, "uploads"), cfg.Locations.UploadPath)
	assert.Equal(t, filepath.Join(dir, "service"), cfg.Locations.ServicePath)
}

// ---------------------------------------------------------------------------
// verifiedPort
// ---------------------------------------------------------------------------

func Test_verifiedPort_ValidStandardPort(t *testing.T) {
	port, err := verifiedPort(8080)

	require.NoError(t, err)
	assert.Equal(t, uint16(8080), port)
}

func Test_verifiedPort_ValidHighPort(t *testing.T) {
	port, err := verifiedPort(65535)

	require.NoError(t, err)
	assert.Equal(t, uint16(65535), port)
}

func Test_verifiedPort_Boundary1024(t *testing.T) {
	port, err := verifiedPort(1024)

	require.NoError(t, err)
	assert.Equal(t, uint16(1024), port)
}

func Test_verifiedPort_ConflictingPort80(t *testing.T) {
	port, err := verifiedPort(80)

	require.ErrorIs(t, err, ErrConflictingPort)
	assert.Equal(t, uint16(80), port)
}

func Test_verifiedPort_ConflictingPortBoundary1023(t *testing.T) {
	port, err := verifiedPort(1023)

	require.ErrorIs(t, err, ErrConflictingPort)
	assert.Equal(t, uint16(1023), port)
}

func Test_verifiedPort_ZeroPort_Conflicting(t *testing.T) {
	// port 0 is in [0, 1023) range -> ErrConflictingPort, not invalid.
	port, err := verifiedPort(0)

	require.ErrorIs(t, err, ErrConflictingPort)
	assert.Equal(t, uint16(0), port)
}

func Test_verifiedPort_NegativePort_Invalid(t *testing.T) {
	port, err := verifiedPort(-1)

	require.ErrorIs(t, err, ErrInvalidPort)
	assert.Equal(t, uint16(0), port)
}

func Test_verifiedPort_Above65535_Invalid(t *testing.T) {
	port, err := verifiedPort(65536)

	require.ErrorIs(t, err, ErrInvalidPort)
	assert.Equal(t, uint16(0), port)
}

func Test_verifiedPort_HugePort_Invalid(t *testing.T) {
	port, err := verifiedPort(1 << 30)

	require.ErrorIs(t, err, ErrInvalidPort)
	assert.Equal(t, uint16(0), port)
}

func Test_verifyPort_ErrorsDistinct(t *testing.T) {
	assert.False(t, errors.Is(ErrInvalidPort, ErrConflictingPort))
	assert.False(t, errors.Is(ErrConflictingPort, ErrInvalidPort))
	assert.ErrorIs(t, ErrConflictingPort, ErrConflictingPort)
	assert.ErrorIs(t, ErrInvalidPort, ErrInvalidPort)
}
