package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create a dummy judo.properties file
	judoProps := `app_name=TestApp
model_dir=model
app_dir=application`
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Get the actual resolved temp directory path (macOS may symlink /var to /private/var)
	resolvedTempDir, err := filepath.EvalSymlinks(tempDir)
	require.NoError(t, err)

	// Assert that the config is loaded correctly
	assert.Equal(t, "TestApp", cfg.AppName)
	assert.Equal(t, filepath.Join(resolvedTempDir, "model"), cfg.ModelDir)
	assert.Equal(t, filepath.Join(resolvedTempDir, "application"), cfg.AppDir)
}

func TestGetConfig_NoPropertiesFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Reset the config instance
	Reset()

	// Get the config - should use defaults when no properties file exists
	cfg := GetConfig()

	// Assert that default values are used
	assert.Equal(t, filepath.Base(tempDir), cfg.AppName)
	assert.Equal(t, tempDir, cfg.ModelDir)
	assert.Equal(t, filepath.Join(tempDir, "application"), cfg.AppDir)
	assert.Equal(t, "karaf", cfg.Runtime)
	assert.Equal(t, "hsqldb", cfg.DBType)
	assert.Equal(t, 8181, cfg.KarafPort)
	assert.Equal(t, 5432, cfg.PostgresPort)
	assert.Equal(t, 8080, cfg.KeycloakPort)
}

func TestGetConfig_ProfileProperties(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create a profile properties file (should take precedence over judo.properties)
	profileProps := `app_name=ProfileApp
model_dir=profile_model
app_dir=profile_app
runtime=compose
dbtype=postgresql
karaf_port=8282
postgres_port=5433
keycloak_port=8081`
	profilePropsPath := filepath.Join(tempDir, "test.properties")
	err = ioutil.WriteFile(profilePropsPath, []byte(profileProps), 0644)
	require.NoError(t, err)

	// Set the profile
	Profile = "test"
	defer func() { Profile = "" }()

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Assert that profile properties are loaded correctly
	assert.Equal(t, "ProfileApp", cfg.AppName)
	assert.Equal(t, filepath.Join(tempDir, "profile_model"), cfg.ModelDir)
	assert.Equal(t, filepath.Join(tempDir, "profile_app"), cfg.AppDir)
	assert.Equal(t, "compose", cfg.Runtime)
	assert.Equal(t, "postgresql", cfg.DBType)
	assert.Equal(t, 8282, cfg.KarafPort)
	assert.Equal(t, 5433, cfg.PostgresPort)
	assert.Equal(t, 8081, cfg.KeycloakPort)
}

func TestApplyInlineOptions(t *testing.T) {
	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Apply inline options
	ApplyInlineOptions("runtime=compose,dbtype=postgres,karaf_port=8182")

	// Assert that the options are applied correctly
	assert.Equal(t, "compose", cfg.Runtime)
	assert.Equal(t, "postgresql", cfg.DBType)
	assert.Equal(t, 8182, cfg.KarafPort)
}

func TestApplyInlineOptions_AllOptions(t *testing.T) {
	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Apply all possible inline options
	ApplyInlineOptions("runtime=compose,dbtype=hsqldb,model_dir=custom_model,karaf_port=8282,postgres_port=5433,keycloak_port=8081,compose_env=dev,compose_access_ip=192.168.1.100,karaf_enable_admin_user=true,java_compiler=javac")

	// Assert that all options are applied correctly
	assert.Equal(t, "compose", cfg.Runtime)
	assert.Equal(t, "hsqldb", cfg.DBType)
	assert.Equal(t, filepath.Join(cfg.ModelDir, "custom_model"), cfg.ModelDir)
	assert.Equal(t, 8282, cfg.KarafPort)
	assert.Equal(t, 5433, cfg.PostgresPort)
	assert.Equal(t, 8081, cfg.KeycloakPort)
	assert.Equal(t, "dev", cfg.ComposeEnv)
	assert.Equal(t, "192.168.1.100", cfg.ComposeAccessIP)
	assert.True(t, cfg.KarafEnableAdminUser)
	assert.Equal(t, "javac", cfg.JavaCompiler)
}

func TestApplyInlineOptions_InvalidPorts(t *testing.T) {
	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Apply inline options with invalid port numbers
	ApplyInlineOptions("karaf_port=invalid,postgres_port=abc,keycloak_port=def")

	// Assert that invalid ports are ignored and defaults are preserved
	assert.Equal(t, 8181, cfg.KarafPort)
	assert.Equal(t, 5432, cfg.PostgresPort)
	assert.Equal(t, 8080, cfg.KeycloakPort)
}

func TestReadProperties(t *testing.T) {
	// Create a temporary properties file
	tempFile, err := ioutil.TempFile("", "test-properties")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test properties
	properties := `# This is a comment
app_name=TestApp
model_dir=model
app_dir=application
; Another comment
runtime=karaf
empty_key=
=empty_value`
	_, err = tempFile.WriteString(properties)
	require.NoError(t, err)
	tempFile.Close()

	// Read the properties
	props := readProperties(tempFile.Name())

	// Assert that properties are parsed correctly
	assert.Equal(t, "TestApp", props["app_name"])
	assert.Equal(t, "model", props["model_dir"])
	assert.Equal(t, "application", props["app_dir"])
	assert.Equal(t, "karaf", props["runtime"])
	assert.Equal(t, "", props["empty_key"])
	assert.Empty(t, props[""])
	assert.NotContains(t, props, "# This is a comment")
	assert.NotContains(t, props, "; Another comment")
}

func TestReadProperties_NonExistentFile(t *testing.T) {
	// Try to read from a non-existent file
	props := readProperties("/non/existent/path.properties")

	// Should return nil
	assert.Nil(t, props)
}

func TestReadProperties_EmptyFile(t *testing.T) {
	// Create an empty temporary file
	tempFile, err := ioutil.TempFile("", "empty-properties")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Read the empty properties file
	props := readProperties(tempFile.Name())

	// Should return empty map
	assert.Empty(t, props)
}

func TestIsProjectInitialized(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Test with no properties files
	assert.False(t, IsProjectInitialized())

	// Create judo.properties file
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte("app_name=TestApp"), 0644)
	require.NoError(t, err)

	// Should be considered initialized
	assert.True(t, IsProjectInitialized())

	// Remove judo.properties and create judo-version.properties
	os.Remove(judoPropsPath)
	judoVersionPropsPath := filepath.Join(tempDir, "judo-version.properties")
	err = ioutil.WriteFile(judoVersionPropsPath, []byte("version=1.0.0"), 0644)
	require.NoError(t, err)

	// Should still be considered initialized
	assert.True(t, IsProjectInitialized())
}

func TestLoadProperties(t *testing.T) {
	// This function is called in PersistentPreRun and should not panic
	// We'll just test that it can be called without error
	assert.NotPanics(t, func() {
		LoadProperties()
	})
}

func TestSetupEnvironment(t *testing.T) {
	// This function exists for compatibility and should not panic
	assert.NotPanics(t, func() {
		SetupEnvironment()
	})
}

func TestConfig_LoadProperties_RelativePaths(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create a judo.properties file with relative paths
	judoProps := `app_name=TestApp
model_dir=../relative_model
app_dir=./relative_app`
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Assert that relative paths are resolved correctly
	expectedModelDir := filepath.Clean(filepath.Join(tempDir, "../relative_model"))
	expectedAppDir := filepath.Clean(filepath.Join(tempDir, "./relative_app"))
	assert.Equal(t, expectedModelDir, cfg.ModelDir)
	assert.Equal(t, expectedAppDir, cfg.AppDir)
}

func TestConfig_LoadProperties_AbsolutePaths(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Create another temp directory for absolute path testing
	absModelDir, err := ioutil.TempDir("", "absolute-model")
	require.NoError(t, err)
	defer os.RemoveAll(absModelDir)

	// Create a judo.properties file with absolute paths
	judoProps := fmt.Sprintf("app_name=TestApp\nmodel_dir=%s\napp_dir=%s", absModelDir, tempDir)
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Assert that absolute paths are preserved
	assert.Equal(t, absModelDir, cfg.ModelDir)
	assert.Equal(t, tempDir, cfg.AppDir)
}

func TestConfig_LoadProperties_DBTypeAliases(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Test that "postgres" is converted to "postgresql"
	judoProps := `dbtype=postgres`
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Assert that "postgres" is converted to "postgresql"
	assert.Equal(t, "postgresql", cfg.DBType)

	// Test that other dbtypes are preserved
	os.Remove(judoPropsPath)
	judoProps = `dbtype=hsqldb`
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	Reset()
	cfg = GetConfig()
	assert.Equal(t, "hsqldb", cfg.DBType)
}

func TestConfig_LoadProperties_BooleanValues(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// Test boolean parsing for karaf_enable_admin_user
	judoProps := `karaf_enable_admin_user=true`
	judoPropsPath := filepath.Join(tempDir, "judo.properties")
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Assert that boolean values are parsed correctly
	assert.True(t, cfg.KarafEnableAdminUser)

	// Test with "1"
	os.Remove(judoPropsPath)
	judoProps = `karaf_enable_admin_user=1`
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	Reset()
	cfg = GetConfig()
	assert.True(t, cfg.KarafEnableAdminUser)

	// Test with "false"
	os.Remove(judoPropsPath)
	judoProps = `karaf_enable_admin_user=false`
	err = ioutil.WriteFile(judoPropsPath, []byte(judoProps), 0644)
	require.NoError(t, err)

	Reset()
	cfg = GetConfig()
	assert.False(t, cfg.KarafEnableAdminUser)
}
