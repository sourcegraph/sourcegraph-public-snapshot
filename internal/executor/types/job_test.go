package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestJob_MarshalJSON(t *testing.T) {
	modAt, err := time.Parse(time.RFC3339, "2022-10-07T18:55:45.831031-06:00")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    Job
		expected string
	}{
		{
			name: "4.3",
			input: Job{
				Version:             2,
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]VirtualMachineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
					"script2.py": {
						Bucket:     "my-bucket",
						Key:        "file/key",
						ModifiedAt: modAt,
					},
				},
				DockerSteps: []DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
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
			expected: `{
		"version": 2,
		"id": 1,
		"token": "",
		"repositoryName": "my-repo",
		"repositoryDirectory": "foo/bar",
		"commit": "xyz",
		"fetchTags": true,
		"shallowClone": true,
		"sparseCheckout": ["a", "b", "c"],
		"files": {
			"script1.sh": {
				"content": "aGVsbG8=",
				"modifiedAt": "0001-01-01T00:00:00Z"
			},
			"script2.py": {
				"bucket": "my-bucket",
				"key": "file/key",
				"modifiedAt": "2022-10-07T18:55:45.831031-06:00"
			}
		},
		"dockerAuthConfig": {},
		"dockerSteps": [{
			"image": "my-image",
			"commands": ["run"],
			"dir": "faz/baz",
			"env": ["FOO=BAR"]
		}],
		"cliSteps": [{
			"command": ["x", "y", "z"],
			"dir": "raz/daz",
			"env": ["BAZ=FAZ"]
		}],
		"redactedValues": {
			"password": "foo"
		}
	}`,
		},
		{
			name: "4.2",
			input: Job{
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]VirtualMachineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
				},
				DockerSteps: []DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
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
			expected: `{
		"id": 1,
		"token": "",
		"repositoryName": "my-repo",
		"repositoryDirectory": "foo/bar",
		"commit": "xyz",
		"fetchTags": true,
		"shallowClone": true,
		"sparseCheckout": ["a", "b", "c"],
		"files": {
			"script1.sh": {
				"content": "hello",
				"modifiedAt": "0001-01-01T00:00:00Z"
			}
		},
		"dockerSteps": [{
			"image": "my-image",
			"commands": ["run"],
			"dir": "faz/baz",
			"env": ["FOO=BAR"]
		}],
		"cliSteps": [{
			"command": ["x", "y", "z"],
			"dir": "raz/daz",
			"env": ["BAZ=FAZ"]
		}],
		"redactedValues": {
			"password": "foo"
		}
	}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := json.Marshal(test.input)
			require.NoError(t, err)
			var actualMap, expectedMap map[string]any
			if err := json.Unmarshal(actual, &actualMap); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.expected), &expectedMap); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(expectedMap, actualMap); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestJob_UnmarshalJSON(t *testing.T) {
	modAt, err := time.Parse(time.RFC3339, "2022-10-07T18:55:45.831031-06:00")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected Job
	}{
		{
			name: "4.3",
			input: `{
	"version": 2,
	"id": 1,
	"repositoryName": "my-repo",
	"repositoryDirectory": "foo/bar",
	"commit": "xyz",
	"fetchTags": true,
	"shallowClone": true,
	"sparseCheckout": ["a", "b", "c"],
	"files": {
		"script1.sh": {
			"content": "aGVsbG8="
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
		"command": ["x", "y", "z"],
		"dir": "raz/daz",
		"env": ["BAZ=FAZ"]
	}],
	"redactedValues": {
		"password": "foo"
	}
}`,
			expected: Job{
				Version:             2,
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]VirtualMachineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
					"script2.py": {
						Bucket:     "my-bucket",
						Key:        "file/key",
						ModifiedAt: modAt,
					},
				},
				DockerSteps: []DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
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
			name: "4.2",
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
		}
	},
	"dockerSteps": [{
		"image": "my-image",
		"commands": ["run"],
		"dir": "faz/baz",
		"env": ["FOO=BAR"]
	}],
	"cliSteps": [{
		"command": ["x", "y", "z"],
		"dir": "raz/daz",
		"env": ["BAZ=FAZ"]
	}],
	"redactedValues": {
		"password": "foo"
	}
}`,
			expected: Job{
				ID:                  1,
				RepositoryName:      "my-repo",
				RepositoryDirectory: "foo/bar",
				Commit:              "xyz",
				FetchTags:           true,
				ShallowClone:        true,
				SparseCheckout:      []string{"a", "b", "c"},
				VirtualMachineFiles: map[string]VirtualMachineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
				},
				DockerSteps: []DockerStep{
					{
						Image:    "my-image",
						Commands: []string{"run"},
						Dir:      "faz/baz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
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
			var actual Job
			err := json.Unmarshal([]byte(test.input), &actual)
			require.NoError(t, err)
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
