package db

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mocking utils.ExecuteCommand and os.Create/os.Open is complex.
// These tests would require a more sophisticated mocking framework or
// integration tests that spin up actual Docker containers.

func TestDumpPostgresql(t *testing.T) {
	// This test would require mocking external commands (docker exec) and file system operations.
	// For now, it serves as a placeholder.
	t.Skip("Skipping TestDumpPostgresql: Requires mocking external dependencies.")
}

func TestImportPostgresql(t *testing.T) {
	// This test would require mocking external commands (docker exec) and file system operations.
	// For now, it serves as a placeholder.
	t.Skip("Skipping TestImportPostgresql: Requires mocking external dependencies.")
}

func TestFindLatestDump(t *testing.T) {
	// Create a temporary directory for test dumps
	tempDir, err := os.MkdirTemp("", "test_dumps")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir) // Clean up after test

	originalCwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(tempDir)) // Change to temp directory for glob to work

	defer func() {
		assert.NoError(t, os.Chdir(originalCwd)) // Change back to original directory
	}()

	schema := "testschema"
	
	// Create dummy dump files with different timestamps
	now := time.Now()
	
	// Oldest dump
	oldTime := now.Add(-24 * time.Hour)
	oldDumpName := fmt.Sprintf("%s_dump_%s.tar.gz", schema, oldTime.Format("20060102_150405"))
	_, err = os.Create(oldDumpName)
	assert.NoError(t, err)

	// Middle dump
	middleTime := now.Add(-12 * time.Hour)
	middleDumpName := fmt.Sprintf("%s_dump_%s.tar.gz", schema, middleTime.Format("20060102_150405"))
	_, err = os.Create(middleDumpName)
	assert.NoError(t, err)

	// Latest dump
	latestTime := now.Add(-1 * time.Hour) // Slightly before 'now' to ensure it's not exactly 'now'
	latestDumpName := fmt.Sprintf("%s_dump_%s.tar.gz", schema, latestTime.Format("20060102_150405"))
	_, err = os.Create(latestDumpName)
	assert.NoError(t, err)

	// Test case 1: Dumps exist, find latest
	foundDump, err := FindLatestDump(schema)
	assert.NoError(t, err)
	assert.Equal(t, latestDumpName, foundDump)

	// Test case 2: No dumps exist
	os.Remove(oldDumpName)
	os.Remove(middleDumpName)
	os.Remove(latestDumpName)
	
	_, err = FindLatestDump(schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no dump files found")

	// Test case 3: Only one dump exists
	_, err = os.Create(oldDumpName)
	assert.NoError(t, err)
	foundDump, err = FindLatestDump(schema)
	assert.NoError(t, err)
	assert.Equal(t, oldDumpName, foundDump)
}
