package files_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
)

func TestGetWorkspaceFiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

func TestScriptNameFromJobStep(t *testing.T) {
	tests := []struct {
		name         string
		job          types.Job
		index        int
		expectedName string
	}{
		{
			name: "Simple",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
			},
			index:        0,
			expectedName: "42.0_github.com_sourcegraph_sourcegraph@.sh",
		},
		{
			name: "Step one",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
			},
			index:        1,
			expectedName: "42.1_github.com_sourcegraph_sourcegraph@.sh",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scriptName := files.ScriptNameFromJobStep(test.job, test.index)
			assert.Equal(t, test.expectedName, scriptName)
		})
	}
}
