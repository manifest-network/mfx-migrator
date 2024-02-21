package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// SetupTempDir creates a temporary directory for testing and returns the path to the directory and a cleanup function.
func SetupTempDir(t *testing.T) (string, func()) {
	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	// Change the current working directory to the temporary directory so that the state files are written there
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	return tempDir, func() {
		// Cleanup function to remove the temporary directory
		os.RemoveAll(tempDir)
	}
}
