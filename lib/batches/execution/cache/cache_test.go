package cache

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/env"
)

func TestExecutionKey_Key(t *testing.T) {
	var singleStepEnv env.Environment
	err := json.Unmarshal([]byte(`{"FOO": "BAR"}`), &singleStepEnv)
	require.NoError(t, err)

	var multipleStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "BAZ": "FAZ"}`), &multipleStepEnv)
	require.NoError(t, err)

	var nullStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "TEST_EXECUTION_CACHE_KEY_ENV": null}`), &nullStepEnv)
	require.NoError(t, err)

	tests := []struct {
		name        string
		keyer       ExecutionKey
		expectedKey string
	}{
		{
			name: "simple key",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo"}},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "multiple steps",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{
					{Run: "foo"},
					{Run: "bar"},
				},
			},
			expectedKey: "nXrDA5Sv3jE2wGVTrixgJw",
		},
		{
			name: "step env",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: singleStepEnv}},
			},
			expectedKey: "Ye3eFDmvvADzZuz-TWEA2g",
		},
		{
			name: "multiple step envs",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: multipleStepEnv}},
			},
			expectedKey: "mZk8q7zjJioxI2nTwrt7XQ",
		},
		{
			name: "null step env",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: nullStepEnv}},
			},
			expectedKey: "_txGuv3XrkWWVQz6hGsKhw",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedKey, key)
		})
	}
}

func TestExecutionKey_Key_Mount(t *testing.T) {
	// Mounts are trickier because there are temp paths and modifications dates. Each run will generate new temp paths
	// and modification dates. Also, different Operating Systems will yield different results causing an expectedKey
	// assertion to fail.
	// So unfortunately, asserting the cache key is not as pretty for mounts.

	tempDir := t.TempDir()

	// Set a specific date on the temp files
	modDate := time.Date(2022, 1, 2, 3, 5, 6, 7, time.UTC)
	modDateVal, err := modDate.MarshalJSON()
	require.NoError(t, err)

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

	tests := []struct {
		name  string
		keyer ExecutionKey
		// Instead of an expectedKey, use raw so the data that will be manually hashed can be controlled.
		expectedRaw   string
		expectedError error
	}{
		{
			name: "single file",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{
					{
						Run: "foo",
						Mount: []batches.Mount{{
							Path:       sampleScriptPath,
							Mountpoint: "/tmp/foo.sh",
						}},
					},
				},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":{},"mount":[{"mountpoint":"/tmp/foo.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
				sampleScriptPath,
				string(modDateVal),
			),
		},
		{
			name: "multiple files",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{
					{
						Run: "foo",
						Mount: []batches.Mount{
							{
								Path:       sampleScriptPath,
								Mountpoint: "/tmp/foo.sh",
							},
							{
								Path:       anotherScriptPath,
								Mountpoint: "/tmp/bar.sh",
							},
						},
					},
				},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":{},"mount":[{"mountpoint":"/tmp/foo.sh","path":"%s"},{"mountpoint":"/tmp/bar.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s},{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
				anotherScriptPath,
				sampleScriptPath,
				string(modDateVal),
				anotherScriptPath,
				string(modDateVal),
			),
		},
		{
			name: "directory",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{
					{
						Run: "foo",
						Mount: []batches.Mount{{
							Path:       tempDir,
							Mountpoint: "/tmp/scripts",
						}},
					},
				},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":{},"mount":[{"mountpoint":"/tmp/scripts","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s},{"Path":"%s","Size":0,"Modified":%s}]}`,
				tempDir,
				anotherScriptPath,
				string(modDateVal),
				sampleScriptPath,
				string(modDateVal),
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				expectedHash := sha256.Sum256([]byte(test.expectedRaw))
				expectedKey := base64.RawURLEncoding.EncodeToString(expectedHash[:16])
				assert.Equal(t, expectedKey, key)
			}
		})
	}
}

func TestExecutionKeyWithGlobalEnv_Key(t *testing.T) {
	var stepEnv env.Environment
	// use an array to get the key to have a nil value
	err := json.Unmarshal([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	tests := []struct {
		name          string
		keyer         ExecutionKeyWithGlobalEnv
		expectedKey   string
		expectedError error
	}{
		{
			name: "simple key",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo"}},
				},
				GlobalEnv: []string{},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "has global env",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og",
		},
		{
			name: "env not updated",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"FAZ=BAZ"},
			},
			expectedKey: "tq9NsiMdvoKqMpgxE00XGQ",
		},
		{
			name: "malformed global env",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV"},
			},
			expectedError: errors.New("resolving environment for step 0: unable to parse environment variable \"SOME_ENV\""),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, test.expectedKey, key)
			}
		})
	}
}

func TestExecutionKeyWithGlobalEnv_Key_Mount(t *testing.T) {
	// Mounts are trickier because there are temp paths and modifications dates. Each run will generate new temp paths
	// and modification dates. Also, different Operating Systems will yield different results causing an expectedKey
	// assertion to fail.
	// So unfortunately, asserting the cache key is not as pretty for mounts.

	tempDir := t.TempDir()

	// Set a specific date on the temp files
	modDate := time.Date(2022, 1, 2, 3, 5, 6, 7, time.UTC)
	modDateVal, err := modDate.MarshalJSON()
	require.NoError(t, err)

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

	var stepEnv env.Environment
	// use an array to get the key to have a nil value
	err = json.Unmarshal([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	tests := []struct {
		name  string
		keyer ExecutionKeyWithGlobalEnv
		// Instead of an expectedKey, use raw so the data that will be manually hashed can be controlled.
		expectedRaw   string
		expectedError error
	}{
		{
			name: "single file",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{
						{
							Run: "foo",
							Mount: []batches.Mount{{
								Path:       sampleScriptPath,
								Mountpoint: "/tmp/foo.sh",
							}},
							Env: stepEnv,
						},
					},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":["SOME_ENV"],"mount":[{"mountpoint":"/tmp/foo.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{"SOME_ENV":"FOO"}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
				sampleScriptPath,
				string(modDateVal),
			),
		},
		{
			name: "multiple files",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{
						{
							Run: "foo",
							Mount: []batches.Mount{
								{
									Path:       sampleScriptPath,
									Mountpoint: "/tmp/foo.sh",
								},
								{
									Path:       anotherScriptPath,
									Mountpoint: "/tmp/bar.sh",
								},
							},
							Env: stepEnv,
						},
					},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":["SOME_ENV"],"mount":[{"mountpoint":"/tmp/foo.sh","path":"%s"},{"mountpoint":"/tmp/bar.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{"SOME_ENV":"FOO"}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s},{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
				anotherScriptPath,
				sampleScriptPath,
				string(modDateVal),
				anotherScriptPath,
				string(modDateVal),
			),
		},
		{
			name: "directory",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{
						{
							Run: "foo",
							Mount: []batches.Mount{{
								Path:       tempDir,
								Mountpoint: "/tmp/scripts",
							}},
							Env: stepEnv,
						},
					},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedRaw: fmt.Sprintf(
				`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":["SOME_ENV"],"mount":[{"mountpoint":"/tmp/scripts","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{"SOME_ENV":"FOO"}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s},{"Path":"%s","Size":0,"Modified":%s}]}`,
				tempDir,
				anotherScriptPath,
				string(modDateVal),
				sampleScriptPath,
				string(modDateVal),
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				expectedHash := sha256.Sum256([]byte(test.expectedRaw))
				expectedKey := base64.RawURLEncoding.EncodeToString(expectedHash[:16])
				assert.Equal(t, expectedKey, key)
			}
		})
	}
}
