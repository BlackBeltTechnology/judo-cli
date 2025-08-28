package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGenerateCommand(t *testing.T) {
	cmd := CreateGenerateCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "generate", cmd.Use)
	assert.Equal(t, "Generate application based on model in JUDO project.", cmd.Short)

	// Test flags
	ignoreFlag := cmd.Flags().Lookup("ignore-checksum")
	assert.NotNil(t, ignoreFlag)
	assert.Equal(t, "i", ignoreFlag.Shorthand)
	assert.Equal(t, "Ignore checksum errors and update checksums", ignoreFlag.Usage)
	assert.Equal(t, "false", ignoreFlag.DefValue)

	// Test RunE (basic check, full test would require mocking os.Getwd and utils.Run)
	// For now, we just ensure it's not nil
	assert.NotNil(t, cmd.RunE)
}
