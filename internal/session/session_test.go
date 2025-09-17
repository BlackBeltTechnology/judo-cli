package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetCurrentDir(t *testing.T) {
	result := GetCurrentDir()
	if result == "unknown" {
		t.Error("GetCurrentDir should not return 'unknown'")
	}
	if !strings.Contains(result, "/") {
		t.Errorf("GetCurrentDir should return a valid path, got: %s", result)
	}
}

func TestGetHistoryFilePath(t *testing.T) {
	result := getHistoryFilePath()
	if result == "" {
		t.Error("getHistoryFilePath should not return empty string")
	}
	if !strings.Contains(result, ".judo") || !strings.Contains(result, "session_history.json") {
		t.Errorf("getHistoryFilePath should contain .judo and session_history.json, got: %s", result)
	}
}

func TestLoadSessionHistory(t *testing.T) {
	// Test with non-existent file (by using a non-existent temp dir)
	tempDir := t.TempDir()

	// Test loading by temporarily setting environment variable to non-existent path
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Join(tempDir, "nonexistent"))
	defer os.Setenv("HOME", originalHome)

	history := loadSessionHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history for non-existent file, got: %v", history)
	}

	// Test with valid file
	tempFile := filepath.Join(tempDir, ".judo", "session_history.json")
	os.MkdirAll(filepath.Dir(tempFile), 0755)

	// Create test data
	testHistory := []string{"command1", "command2", "command3"}
	data, _ := json.Marshal(testHistory)
	os.WriteFile(tempFile, data, 0644)

	// Test loading by temporarily setting environment variable
	os.Setenv("HOME", tempDir)

	history = loadSessionHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 commands in history, got: %v", history)
	}
	if len(history) >= 3 && (history[0] != "command1" || history[1] != "command2" || history[2] != "command3") {
		t.Errorf("History content mismatch, got: %v", history)
	}

	// Test with corrupted file
	os.WriteFile(tempFile, []byte("invalid json"), 0644)
	history = loadSessionHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history for corrupted file, got: %v", history)
	}
}

func TestSaveSessionHistory(t *testing.T) {
	tempDir := t.TempDir()

	// Test saving by temporarily setting environment variable
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test saving history
	testHistory := []string{"test1", "test2", "test3"}
	saveSessionHistory(testHistory)

	// Verify file was created and contains correct data
	historyFile := filepath.Join(tempDir, ".judo", "session_history.json")
	data, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("Failed to read saved history file: %v", err)
	}

	var loadedHistory []string
	err = json.Unmarshal(data, &loadedHistory)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved history: %v", err)
	}

	if len(loadedHistory) != 3 {
		t.Errorf("Expected 3 commands in saved history, got: %v", loadedHistory)
	}

	// Test history truncation (limit to 100)
	largeHistory := make([]string, 150)
	for i := 0; i < 150; i++ {
		largeHistory[i] = "command" + string(rune('0'+i%10))
	}
	saveSessionHistory(largeHistory)

	data, _ = os.ReadFile(historyFile)
	json.Unmarshal(data, &loadedHistory)
	if len(loadedHistory) != 100 {
		t.Errorf("Expected history to be truncated to 100, got: %d", len(loadedHistory))
	}
}

func TestGetCommandSuggestions(t *testing.T) {
	tests := []struct {
		input      string
		expected   []string
		shouldFind bool
	}{
		{"he", []string{"help"}, true},
		{"ex", []string{"exit"}, true},
		{"st", []string{"start", "stop", "status"}, true},
		{"xyz", []string{}, false},
		{"", []string{"help", "exit", "quit", "clear", "history", "status", "doctor", "init", "build", "start", "stop", "clean", "prune", "update", "generate", "generate-root", "dump", "import", "schema-upgrade", "reckless", "self-update"}, true},
	}

	for _, test := range tests {
		result := getCommandSuggestions(test.input)

		if test.shouldFind && len(result) == 0 {
			t.Errorf("Expected suggestions for input '%s', got none", test.input)
		}

		if !test.shouldFind && len(result) > 0 {
			t.Errorf("Expected no suggestions for input '%s', got: %v", test.input, result)
		}

		// Check that all returned suggestions start with the input
		for _, suggestion := range result {
			if !strings.HasPrefix(strings.ToLower(suggestion), strings.ToLower(test.input)) {
				t.Errorf("Suggestion '%s' does not start with input '%s'", suggestion, test.input)
			}
		}
	}
}

func TestGetArgumentSuggestions(t *testing.T) {
	tests := []struct {
		commandLine string
		expected    []string
	}{
		{"build", []string{"--build-parallel", "-p", "--build-app-module", "-a", "--build-frontend-module", "-f", "--docker", "--skip-model", "--skip-backend", "--skip-frontend", "--skip-karaf", "--skip-schema", "--build-schema-cli", "--version", "-v", "--maven-argument", "-m", "--quick", "-q", "--ignore-checksum", "-i"}},
		{"start", []string{"--skip-keycloak", "--skip-watch-bundles", "--options"}},
		{"doctor", []string{"--verbose", "-v"}},
		{"prune", []string{"--frontend", "-f", "--yes", "-y"}},
		{"unknown", []string{}},
		{"", []string{}},
	}

	for _, test := range tests {
		result := getArgumentSuggestions(test.commandLine)

		if len(result) != len(test.expected) {
			t.Errorf("For command '%s', expected %d suggestions, got %d: %v", test.commandLine, len(test.expected), len(result), result)
		}

		for i, expectedArg := range test.expected {
			if i >= len(result) || result[i] != expectedArg {
				t.Errorf("For command '%s', expected suggestion %d to be '%s', got '%s'", test.commandLine, i, expectedArg, result[i])
			}
		}
	}
}

func TestGetHistoryBasedSuggestions(t *testing.T) {
	history := []string{
		"build",
		"start",
		"build --docker",
		"status",
		"build --quick",
	}

	tests := []struct {
		input    string
		expected []string
	}{
		{"b", []string{"build --quick", "build --docker", "build"}},
		{"bu", []string{"build --quick", "build --docker", "build"}},
		{"s", []string{"status", "start"}},
		{"st", []string{"status", "start"}},
		{"xyz", []string{}},
	}

	for _, test := range tests {
		result := getHistoryBasedSuggestions(test.input, history)

		if len(result) != len(test.expected) {
			t.Errorf("For input '%s', expected %d suggestions, got %d: %v", test.input, len(test.expected), len(result), result)
		}

		for i, expected := range test.expected {
			if i >= len(result) || result[i] != expected {
				t.Errorf("For input '%s', expected suggestion %d to be '%s', got '%s'", test.input, i, expected, result[i])
			}
		}
	}
}

func TestGetStatusColor(t *testing.T) {
	if getStatusColor(true) != "\x1b[32m✓\x1b[0m" {
		t.Error("Expected green check for running status")
	}
	if getStatusColor(false) != "\x1b[31m✗\x1b[0m" {
		t.Error("Expected red x for not running status")
	}
}

func TestGetServiceEmoji(t *testing.T) {
	tests := []struct {
		service  string
		expected string
	}{
		{"karaf", "⚙️"},
		{"keycloak", "🔐"},
		{"postgres", "🐘"},
		{"unknown", "⚙️"},
	}

	for _, test := range tests {
		result := getServiceEmoji(test.service)
		if result != test.expected {
			t.Errorf("For service '%s', expected '%s', got '%s'", test.service, test.expected, result)
		}
	}
}

func TestExecuteCommandInSession(t *testing.T) {
	// Test with a simple mock command
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing
		},
	}

	err := executeCommandInSession(cmd, []string{})
	if err != nil {
		t.Errorf("executeCommandInSession should not fail for valid command, got: %v", err)
	}

	// Test with command that has RunE
	cmdWithRunE := &cobra.Command{
		Use: "test-error",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	err = executeCommandInSession(cmdWithRunE, []string{})
	if err != nil {
		t.Errorf("executeCommandInSession should not fail for RunE command, got: %v", err)
	}

	// Test with command that has no run function
	cmdNoRun := &cobra.Command{
		Use: "no-run",
	}

	err = executeCommandInSession(cmdNoRun, []string{})
	if err == nil || !strings.Contains(err.Error(), "no run function") {
		t.Errorf("Expected error about no run function, got: %v", err)
	}
}

func TestUpdateSessionStatus(t *testing.T) {
	// This test is tricky because it depends on external functions
	// We'll just test that it doesn't panic
	state := &SessionState{
		CurrentDir:         "/old/dir",
		ProjectInitialized: false,
	}

	// Should not panic
	updateSessionStatus(state)

	// The state should be updated (though we can't easily mock the external calls)
	if state.CurrentDir == "/old/dir" {
		t.Log("Note: CurrentDir not updated (may be expected depending on mocking)")
	}
}

func TestPrintCommandHistory(t *testing.T) {
	// This function prints to stdout, so we can only test it doesn't panic
	history := []string{}
	printCommandHistory(history) // Should not panic

	history = []string{"command1", "command2"}
	printCommandHistory(history) // Should not panic
}

func TestShowCommandSuggestions(t *testing.T) {
	// This function prints to stdout, so we can only test it doesn't panic
	showCommandSuggestions("he")  // Should not panic
	showCommandSuggestions("xyz") // Should not panic
}

func TestShowEnhancedSuggestions(t *testing.T) {
	// This function prints to stdout, so we can only test it doesn't panic
	history := []string{"build", "start"}
	showEnhancedSuggestions("b", history)   // Should not panic
	showEnhancedSuggestions("xyz", history) // Should not panic
}
