package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readFile reads the whole file at path, or fails the test.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

// ---------------------------------------------------------------------------
// getConfigPath
// ---------------------------------------------------------------------------

func Test_getConfigPath_UsesDefaultWhenEnvUnset(t *testing.T) {
	// Arrange
	t.Setenv("VERITAS_CONFIG_PATH", "")
	home, ok := os.LookupEnv("HOME")
	require.True(t, ok, "HOME must be set (used by the default config path)")

	// Act
	got := getConfigPath()

	// Assert
	assert.Equal(t, filepath.Join(home, ".config", "veritas"), got)
}

func Test_getConfigPath_RespectsEnvOverride(t *testing.T) {
	// Arrange
	t.Setenv("VERITAS_CONFIG_PATH", "/some/custom/dir")

	// Act
	got := getConfigPath()

	// Assert
	assert.Equal(t, "/some/custom/dir", got)
}

func Test_getConfigPath_OverridesDefaultWithEnv(t *testing.T) {
	// Arrange
	t.Setenv("VERITAS_CONFIG_PATH", "$VERITAS_TEST_HOME/veritas")
	t.Setenv("VERITAS_TEST_HOME", "/custom/home")

	// Act
	got := getConfigPath()

	// Assert
	assert.Equal(t, "/custom/home/veritas", got)
}

// ---------------------------------------------------------------------------
// setYamlConfig
// ---------------------------------------------------------------------------

func Test_setYamlConfig_FreshDir_WritesDefaultConfig(t *testing.T) {
	// Arrange
	dir := t.TempDir()

	// Act
	err := setYamlConfig(dir)

	// Assert
	require.NoError(t, err)
	contents := readFile(t, filepath.Join(dir, "config.yaml"))
	assert.Contains(t, contents, "host: \"127.0.0.1\"")
	assert.Contains(t, contents, "port: 9001")
	assert.Contains(t, contents, "upload_ttl: 86400")
	assert.Contains(t, contents, "log_level: \"info\"")
	assert.Contains(t, contents, "log_format: \"text\"")
	assert.Contains(t, contents, "blobs: \"$HOME/.local/share/veritas/blobs\"")
	assert.Contains(t, contents, "service: \"$HOME/.config/systemd/user\"")
}

func Test_setYamlConfig_CreatesNestedParentDirs(t *testing.T) {
	// Arrange
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	// Act
	err := setYamlConfig(dir)

	// Assert
	require.NoError(t, err)
	_, statErr := os.Stat(filepath.Join(dir, "config.yaml"))
	assert.NoError(t, statErr)
}

func Test_setYamlConfig_AlreadyExists_ReturnsNil(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("existing"), 0o600))

	// Act
	err := setYamlConfig(dir)

	// Assert — no error, and pre-existing content must be preserved (O_EXCL)
	require.NoError(t, err)
	assert.Equal(t, "existing", readFile(t, filepath.Join(dir, "config.yaml")))
}

func Test_setYamlConfig_Idempotent_SecondCallKeepsFirstContent(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	require.NoError(t, setYamlConfig(dir))
	first := readFile(t, filepath.Join(dir, "config.yaml"))

	// Act
	err := setYamlConfig(dir)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, first, readFile(t, filepath.Join(dir, "config.yaml")))
}

func Test_setYamlConfig_OverriddenByCustomValue_Preserved(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("port: 9999"), 0o600))

	// Act
	err := setYamlConfig(dir)

	// Assert — a user-customized config is never clobbered
	require.NoError(t, err)
	assert.Equal(t, "port: 9999", readFile(t, filepath.Join(dir, "config.yaml")))
}

func Test_setYamlConfig_UnwritablePath_ReturnsError(t *testing.T) {
	// Arrange — a regular file blocks creation of the config directory
	file := t.TempDir() + "/blocked"
	require.NoError(t, os.WriteFile(file, []byte("file"), 0o644))

	// Act
	err := setYamlConfig(file)

	// Assert
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "config dir"), "expected a wrapped 'config dir' error, got: %v", err)
}

// ---------------------------------------------------------------------------
// parseEnvPath
// ---------------------------------------------------------------------------

func Test_parseEnvPath_NoDollar_ReturnsUnchanged(t *testing.T) {
	// Arrange
	got := parseEnvPath("/plain/path")

	// Assert
	assert.Equal(t, "/plain/path", got)
}

func Test_parseEnvPath_EmptyString(t *testing.T) {
	// Arrange
	got := parseEnvPath("")

	// Assert
	assert.Equal(t, "", got)
}

func Test_parseEnvPath_ExpandsSetVariable(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)

	// Act
	got := parseEnvPath("$VERITAS_TEST_DIR/data")

	// Assert
	assert.Equal(t, filepath.Join(dir, "data"), got)
}

func Test_parseEnvPath_AbsolutePrefixPreserved(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_DIR", dir)
	// dir itself is a temp dir without a $ in it, so this exercises the
	// absolute-path ("/") reconstruction branch.

	// Act
	got := parseEnvPath("/$VERITAS_TEST_DIR/data")

	// Assert
	assert.Equal(t, filepath.Join(dir, "data"), got)
}

func Test_parseEnvPath_MultipleVariablesInOnePath(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_A", filepath.Join(dir, "a"))
	t.Setenv("VERITAS_TEST_B", "b")

	// Act
	got := parseEnvPath("$VERITAS_TEST_A/$VERITAS_TEST_B")

	// Assert
	assert.Equal(t, filepath.Join(dir, "a", "b"), got)
}

func Test_parseEnvPath_UnsetVariable_ReturnsInputUnchanged(t *testing.T) {
	// Arrange — VERITAS_TEST_MISSING is never set; os.LookupEnv reports !ok.
	// (t.Setenv is deliberately NOT used, since it would set the var.)

	// Act
	got := parseEnvPath("$VERITAS_TEST_MISSING/data")

	// Assert
	assert.Equal(t, "$VERITAS_TEST_MISSING/data", got)
}

func Test_parseEnvPath_SetEmptyVariable_ExpandsToEmpty(t *testing.T) {
	// Arrange — a variable that *is* set but empty (os.LookupEnv ok=true)
	t.Setenv("VERITAS_TEST_EMPTY", "")

	// Act
	got := parseEnvPath("$VERITAS_TEST_EMPTY/baz")

	// Assert
	assert.Equal(t, "baz", got)
}

func Test_parseEnvPath_MixedSetThenUnset_ReturnsInputUnchanged(t *testing.T) {
	// Arrange — first var set, second genuinely unset (!ok) so parsing bails.
	// VERITAS_TEST_MISSING is deliberately NOT t.Setenv-ed (that would set it).
	dir := t.TempDir()
	t.Setenv("VERITAS_TEST_OK", dir)

	// Act
	got := parseEnvPath("$VERITAS_TEST_OK/$VERITAS_TEST_NOTSET/data")

	// Assert — the set var is NOT expanded because the later unset var bails
	assert.Equal(t, "$VERITAS_TEST_OK/$VERITAS_TEST_NOTSET/data", got)
}

func Test_parseEnvPath_EmptyVarLeadingSlash(t *testing.T) {
	// Arrange
	t.Setenv("VERITAS_TEST_EMPTY", "")

	// Act
	got := parseEnvPath("/$VERITAS_TEST_EMPTY/bar")

	// Assert
	assert.Equal(t, "/bar", got)
}
