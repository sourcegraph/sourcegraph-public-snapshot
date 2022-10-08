package executor_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
)

func TestJob_UnmarshalJSON(t *testing.T) {
	modAt, err := time.Parse(time.RFC3339, "2022-10-07T18:55:45.831031-06:00")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected executor.Job
	}{
		{
			name: "4.1",
			input: `{
	"id": 1,
	"repositoryName": "my-repo",
	"repositoryDirectory": "foo/bar",
	"commit": "xyz",
	"fetchTags": true,
	"shallowClone": true,
	"sparseCheckout": ["a", "b", "c"],
	"files": {
		"script1.sh": {
			"content": "hello"
		},
		"script2.py": {
			"bucket": "my-bucket",
			"key": "file/key",
			"modifiedAt": "2022-10-07T18:55:45.831031-06:00"
		}
	},
	"dockerSteps": [{
		"image": "my-image",
		"commands": ["run"],
		"dir": "faz/baz",
		"env": ["FOO=BAR"]
	}],
	"cliSteps": [{
		"commands": ["x", "y", "z"],
		"dir": "raz/daz",
		"env": ["BAZ=FAZ"]
	}],
	"redactedValues": {
		"password": "foo"
	}
}`,
			expected: executor.Job{
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]executor.VirtualMachineFile{
					"script1.sh": {
						Content: "hello",
					},
					"script2.py": {
						Bucket:     "my-bucket",
						Key:        "file/key",
						ModifiedAt: modAt,
					},
				},
				DockerSteps: []executor.DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []executor.CliStep{
					{
						Commands: []string{"x", "y", "z"},
						Dir:      "raz/daz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedactedValues: map[string]string{
					"password": "foo",
				},
			},
		},
		{
			name: "4.0",
			input: `{
	"id": 1,
	"repositoryName": "my-repo",
	"repositoryDirectory": "foo/bar",
	"commit": "xyz",
	"fetchTags": true,
	"shallowClone": true,
	"sparseCheckout": ["a", "b", "c"],
	"files": {
		"script1.sh": "hello"
	},
	"dockerSteps": [{
		"image": "my-image",
		"commands": ["run"],
		"dir": "faz/baz",
		"env": ["FOO=BAR"]
	}],
	"cliSteps": [{
		"commands": ["x", "y", "z"],
		"dir": "raz/daz",
		"env": ["BAZ=FAZ"]
	}],
	"redactedValues": {
		"password": "foo"
	}
}`,
			expected: executor.Job{
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]executor.VirtualMachineFile{
					"script1.sh": {
						Content: "hello",
					},
				},
				DockerSteps: []executor.DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []executor.CliStep{
					{
						Commands: []string{"x", "y", "z"},
						Dir:      "raz/daz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedactedValues: map[string]string{
					"password": "foo",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actual executor.Job
			err := json.Unmarshal([]byte(test.input), &actual)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
