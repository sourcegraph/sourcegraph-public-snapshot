package files_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetWorkspaceFiles(t *testing.T) {
	modifiedAt := time.Now()

	tests := []struct {
		name                   string
		job                    types.Job
		mockFunc               func(store *files.MockStore)
		assertFunc             func(t *testing.T, store *files.MockStore)
		expectedWorkspaceFiles []files.WorkspaceFile
		expectedErr            error
	}{
		{
			name: "No files or steps",
			job:  types.Job{},
			assertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspaceFiles: nil,
			expectedErr:            nil,
		},
		{
			name: "Docker Steps",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
				DockerSteps: []types.DockerStep{
					{
						Commands: []string{"echo hello"},
					},
					{
						Commands: []string{"echo world"},
					},
				},
			},
			assertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspaceFiles: []files.WorkspaceFile{
				{
					Path:         "/working/directory/.sourcegraph-executor/42.0_github.com_sourcegraph_sourcegraph@.sh",
					Content:      []byte(files.ScriptPreamble + "\n\necho hello\n"),
					IsStepScript: true,
				},
				{
					Path:         "/working/directory/.sourcegraph-executor/42.1_github.com_sourcegraph_sourcegraph@.sh",
					Content:      []byte(files.ScriptPreamble + "\n\necho world\n"),
					IsStepScript: true,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Virtual machine files",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
				VirtualMachineFiles: map[string]types.VirtualMachineFile{
					"foo.sh": {
						Content: []byte("echo hello"),
					},
					"bar.sh": {
						Content: []byte("echo world"),
					},
				},
			},
			assertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 0)
			},
			expectedWorkspaceFiles: []files.WorkspaceFile{
				{
					Path:         "/working/directory/foo.sh",
					Content:      []byte("echo hello"),
					IsStepScript: false,
				},
				{
					Path:         "/working/directory/bar.sh",
					Content:      []byte("echo world"),
					IsStepScript: false,
				},
			},
			expectedErr: nil,
		},
		{
			name: "Workspace files",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
				VirtualMachineFiles: map[string]types.VirtualMachineFile{
					"foo.sh": {
						Bucket:     "my-bucket",
						Key:        "foo.sh",
						ModifiedAt: modifiedAt,
					},
					"bar.sh": {
						Bucket:     "my-bucket",
						Key:        "bar.sh",
						ModifiedAt: modifiedAt,
					},
				},
			},
			mockFunc: func(store *files.MockStore) {
				store.GetFunc.SetDefaultHook(func(ctx context.Context, job types.Job, bucket string, key string) (io.ReadCloser, error) {
					if key == "foo.sh" {
						return io.NopCloser(bytes.NewBufferString("echo hello")), nil
					}
					if key == "bar.sh" {
						return io.NopCloser(bytes.NewBufferString("echo world")), nil
					}
					return nil, errors.New("unexpected key")
				})
			},
			assertFunc: func(t *testing.T, store *files.MockStore) {
				require.Len(t, store.GetFunc.History(), 2)
				assert.Equal(t, "my-bucket", store.GetFunc.History()[0].Arg2)
				assert.Contains(t, []string{"foo.sh", "bar.sh"}, store.GetFunc.History()[0].Arg3)
				assert.Contains(t, []string{"foo.sh", "bar.sh"}, store.GetFunc.History()[1].Arg3)
			},
			expectedWorkspaceFiles: []files.WorkspaceFile{
				{
					Path:         "/working/directory/foo.sh",
					Content:      []byte("echo hello"),
					IsStepScript: false,
					ModifiedAt:   modifiedAt,
				},
				{
					Path:         "/working/directory/bar.sh",
					Content:      []byte("echo world"),
					IsStepScript: false,
					ModifiedAt:   modifiedAt,
				},
			},
			expectedErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := files.NewMockStore()
			if test.mockFunc != nil {
				test.mockFunc(store)
			}

			workspaceFiles, err := files.GetWorkspaceFiles(context.Background(), store, test.job, "/working/directory")
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
				assert.Nil(t, workspaceFiles)
			} else {
				require.NoError(t, err)
				// To make comparisons easier, in case of failures, we will iterate over the expected and try and find a
				// match in the actual.
				// By doing this, we do not care about order and can more surgically test the expected values.
				for _, expected := range test.expectedWorkspaceFiles {
					found := false
					for _, actual := range workspaceFiles {
						if expected.Path == actual.Path {
							assert.Equal(t, string(expected.Content), string(actual.Content))
							assert.Equal(t, expected.IsStepScript, actual.IsStepScript)
							assert.Equal(t, expected.ModifiedAt, actual.ModifiedAt)
							found = true
							break
						}
					}
					if !found {
						// Get actual file paths
						var actualPaths []string
						for _, actual := range workspaceFiles {
							actualPaths = append(actualPaths, actual.Path)
						}
						assert.Fail(t, "Expected file not found", expected.Path, actualPaths)
					}
				}
			}

			if test.assertFunc != nil {
				test.assertFunc(t, store)
			}
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
		{
			name: "With commit",
			job: types.Job{
				ID:             42,
				RepositoryName: "github.com/sourcegraph/sourcegraph",
				Commit:         "deadbeef",
			},
			index:        1,
			expectedName: "42.1_github.com_sourcegraph_sourcegraph@deadbeef.sh",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scriptName := files.ScriptNameFromJobStep(test.job, test.index)
			assert.Equal(t, test.expectedName, scriptName)
		})
	}
}
