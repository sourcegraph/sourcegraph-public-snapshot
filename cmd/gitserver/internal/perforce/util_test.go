package perforce

import (
	"os/exec"
	"testing"

	"gotest.tools/assert"
)

func TestSpecifyCommandInErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		errorMsg    string
		command     *exec.Cmd
		expectedMsg string
	}{
		{
			name:     "empty error message",
			errorMsg: "",
			command: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "",
		},
		{
			name:     "error message without phrase to replace",
			errorMsg: "Some error",
			command: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "Some error",
		},
		{
			name:        "error message with phrase to replace, nil input Cmd",
			errorMsg:    "Some error",
			command:     nil,
			expectedMsg: "Some error",
		},
		{
			name:        "error message with phrase to replace, empty input Cmd",
			errorMsg:    "Some error",
			command:     &exec.Cmd{},
			expectedMsg: "Some error",
		},
		{
			name:     "error message with phrase to replace, valid input Cmd",
			errorMsg: "error cloning repo: repo perforce/path/to/depot not cloneable: exit status 1 (output follows)\n\nPerforce password (P4PASSWD) invalid or unset.",
			command: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "error cloning repo: repo perforce/path/to/depot not cloneable: exit status 1 (output follows)\n\nPerforce password (P4PASSWD) invalid or unset.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualMsg := specifyCommandInErrorMessage(test.errorMsg, test.command)
			assert.Equal(t, test.expectedMsg, actualMsg)
		})
	}
}
