package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func SetupTmpDir(t *testing.T) string {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	// Change the current working directory to the temporary directory
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Return the path of the temporary directory for cleanup purposes
	return tempDir
}
