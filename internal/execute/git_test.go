package execute

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleGitCommandExec(t *testing.T) {
	// Store the current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "git-ops-test-temp-dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Change to the temporary directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Defer changing back to the original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to change back to original directory: %v", err)
		}
	}()

	tests := []struct {
		name           string
		cmdSetup       func() *exec.Cmd
		expectedOutput string
		expectedError  string
	}{
		{
			name: "Successful command",
			cmdSetup: func() *exec.Cmd {
				return exec.Command("echo", "file1.txt\nfile2.txt")
			},
			expectedOutput: "file1.txt\nfile2.txt\n",
			expectedError:  "",
		},
		{
			name: "Git fatal error",
			cmdSetup: func() *exec.Cmd {
				return exec.Command("git", "rev-parse", "--is-inside-work-tree")
			},
			expectedOutput: "",
			expectedError:  "fatal: not a git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdSetup()
			output, err := handleGitCommandExec(cmd)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedOutput, string(output))
		})
	}
}
