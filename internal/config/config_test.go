package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := ioutil.TempDir("", "judo-cli-test")
	assert.NoError(t, err)
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
	assert.NoError(t, err)

	// Reset the config instance
	Reset()

	// Get the config
	cfg := GetConfig()

	// Get the actual resolved temp directory path (macOS may symlink /var to /private/var)
	resolvedTempDir, err := filepath.EvalSymlinks(tempDir)
	assert.NoError(t, err)

	// Assert that the config is loaded correctly
	assert.Equal(t, "TestApp", cfg.AppName)
	assert.Equal(t, filepath.Join(resolvedTempDir, "model"), cfg.ModelDir)
	assert.Equal(t, filepath.Join(resolvedTempDir, "application"), cfg.AppDir)
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
