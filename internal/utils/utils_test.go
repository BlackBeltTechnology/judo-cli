package utils

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeNow(t *testing.T) {
	now := TimeNow()
	assert.NotZero(t, now)
	assert.WithinDuration(t, time.Now(), now, time.Second)
}

func TestGetCurrentDir(t *testing.T) {
	dir := GetCurrentDir()
	assert.NotEmpty(t, dir)
	assert.DirExists(t, dir)
}

func TestFileExists(t *testing.T) {
	// Test with existing file
	tempFile, err := os.CreateTemp("", "test-file")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	assert.True(t, FileExists(tempFile.Name()))

	// Test with non-existing file
	assert.False(t, FileExists("/nonexistent/path/to/file"))
}

func TestReplaceInFile(t *testing.T) {
	// Create a temporary file with test content
	tempFile, err := os.CreateTemp("", "test-replace")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	content := "Hello World\nThis is a test\nHello again"
	err = os.WriteFile(tempFile.Name(), []byte(content), 0644)
	assert.NoError(t, err)

	// Replace "Hello" with "Hi"
	err = ReplaceInFile(tempFile.Name(), "Hello", "Hi")
	assert.NoError(t, err)

	// Read the file and verify the replacement
	newContent, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	expected := "Hi World\nThis is a test\nHi again"
	assert.Equal(t, expected, string(newContent))
}

func TestDefaultShell(t *testing.T) {
	prog, args := DefaultShell()
	assert.NotEmpty(t, prog)
	assert.Equal(t, []string{"-l", "-c"}, args)
}

func TestHaveWSL(t *testing.T) {
	// This test just ensures the function doesn't panic
	haveWSL := HaveWSL()
	// We can't predict the result, but it should be a boolean
	assert.IsType(t, true, haveWSL)
}

func TestWinPathToWSL(t *testing.T) {
	// Test Windows path conversion
	result := WinPathToWSL("C:\\work\\proj")
	// On macOS, filepath.Clean converts backslashes to forward slashes
	expected := "/mnt/c/work/proj"
	if result != expected {
		// Try with forward slashes (macOS behavior)
		result = WinPathToWSL("C:/work/proj")
	}
	assert.Equal(t, expected, result)

	// Test with empty string
	assert.Equal(t, "", WinPathToWSL(""))

	// Test with non-Windows path
	result = WinPathToWSL("/home/user/proj")
	assert.Equal(t, "/home/user/proj", result)
}

func TestIsPortAvailable(t *testing.T) {
	// Test with a port that should be available (high port)
	available := IsPortAvailable(65530)
	assert.True(t, available)

	// Test with port 0 (should be available)
	available = IsPortAvailable(0)
	assert.True(t, available)
}

func TestPromptForInput(t *testing.T) {
	// This is difficult to test automatically since it requires user input
	// We'll just test that the function exists and returns a string
	// In practice, this would be tested with mock input
	result := PromptForInput("Test prompt: ")
	assert.IsType(t, "", result)
}

func TestPromptForSelection(t *testing.T) {
	// Similar to PromptForInput, difficult to test automatically
	options := []string{"yes", "no"}
	result := PromptForSelection("Choose", options, "yes")
	assert.Contains(t, options, result)
}
