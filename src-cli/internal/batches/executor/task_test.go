package executor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFileMetadataRetriever_Get(t *testing.T) {
	// TODO: use TempDir when https://github.com/golang/go/issues/51442 is cherry-picked into 1.18 or upgrade to 1.19+
	//tempDir := t.TempDir()
	tempDir, err := os.MkdirTemp("", "metadata")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Set a specific date on the temp files
	modDate := time.Date(2022, 1, 2, 3, 5, 6, 0, time.UTC)

	// create temp files/dirs that can be used by the tests
	sampleScriptPath := filepath.Join(tempDir, "sample.sh")
	_, err = os.Create(sampleScriptPath)
	require.NoError(t, err)
	err = os.Chtimes(sampleScriptPath, modDate, modDate)
	require.NoError(t, err)

	anotherScriptPath := filepath.Join(tempDir, "another.sh")
	_, err = os.Create(anotherScriptPath)
	require.NoError(t, err)
	err = os.Chtimes(anotherScriptPath, modDate, modDate)
	require.NoError(t, err)

	retriever := fileMetadataRetriever{
		workingDirectory: tempDir,
	}

	tests := []struct {
		name             string
		steps            []batches.Step
		expectedMetadata []cache.MountMetadata
		expectedError    error
	}{
		{
			name: "simple file",
			steps: []batches.Step{
				{
					Run: "foo",
					Mount: []batches.Mount{{
						Path:       "./sample.sh",
						Mountpoint: "/tmp/foo.sh",
					}},
				},
			},
			expectedMetadata: []cache.MountMetadata{
				{Path: "sample.sh", Size: 0, Modified: modDate},
			},
		},
		{
			name: "multiple files",
			steps: []batches.Step{
				{
					Run: "foo",
					Mount: []batches.Mount{
						{
							Path:       "./sample.sh",
							Mountpoint: "/tmp/foo.sh",
						},
						{
							Path:       "./another.sh",
							Mountpoint: "/tmp/bar.sh",
						},
					},
				},
			},
			expectedMetadata: []cache.MountMetadata{
				{Path: "sample.sh", Size: 0, Modified: modDate},
				{Path: "another.sh", Size: 0, Modified: modDate},
			},
		},
		{
			name: "directory",
			steps: []batches.Step{
				{
					Run: "foo",
					Mount: []batches.Mount{{
						Path:       "./",
						Mountpoint: "/tmp/scripts",
					}},
				},
			},
			expectedMetadata: []cache.MountMetadata{
				{Path: "another.sh", Size: 0, Modified: modDate},
				{Path: "sample.sh", Size: 0, Modified: modDate},
			},
		},
		{
			name: "file does not exist",
			steps: []batches.Step{
				{
					Run: "foo",
					Mount: []batches.Mount{{
						Path:       filepath.Join(tempDir, "some-file-does-not-exist.sh"),
						Mountpoint: "/tmp/file.sh",
					}},
				},
			},
			expectedError: errors.Newf("path %s does not exist", filepath.Join(tempDir, "some-file-does-not-exist.sh")),
		},
		{
			name: "multiple steps",
			steps: []batches.Step{
				{
					Run: "foo",
					Mount: []batches.Mount{{
						Path:       "./sample.sh",
						Mountpoint: "/tmp/foo.sh",
					}},
				},
				{
					Run: "foo",
					Mount: []batches.Mount{{
						Path:       "./sample.sh",
						Mountpoint: "/tmp/foo.sh",
					}},
				},
			},
			expectedMetadata: []cache.MountMetadata{
				{Path: "sample.sh", Size: 0, Modified: modDate},
				{Path: "sample.sh", Size: 0, Modified: modDate},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata, err := retriever.Get(test.steps)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedMetadata, metadata)
			}
		})
	}
}
