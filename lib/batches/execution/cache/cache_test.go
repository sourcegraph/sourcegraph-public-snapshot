package cache

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var repo = batches.Repository{
	ID:          "my-repo",
	Name:        "github.com/sourcegraph/src-cli",
	BaseRef:     "refs/heads/f00b4r",
	BaseRev:     "c0mmit",
	FileMatches: []string{"baz.go"},
}

func TestKeyer_Key(t *testing.T) {
	var singleStepEnv env.Environment
	err := json.Unmarshal([]byte(`{"FOO": "BAR"}`), &singleStepEnv)
	require.NoError(t, err)

	var multipleStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "BAZ": "FAZ"}`), &multipleStepEnv)
	require.NoError(t, err)

	var nullStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "TEST_EXECUTION_CACHE_KEY_ENV": null}`), &nullStepEnv)
	require.NoError(t, err)

	var stepEnv env.Environment
	// use an array to get the key to have a nil value
	err = json.Unmarshal([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	tests := []struct {
		name          string
		keyer         Keyer
		expectedKey   string
		expectedError error
	}{
		{
			name: "ExecutionKey simple",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "ExecutionKey multiple steps",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps: []batches.Step{
					{Run: "foo"},
					{Run: "bar"},
				},
			},
			expectedKey: "nXrDA5Sv3jE2wGVTrixgJw",
		},
		{
			name: "ExecutionKey step env",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: singleStepEnv}},
			},
			expectedKey: "Ye3eFDmvvADzZuz-TWEA2g",
		},
		{
			name: "ExecutionKey multiple step envs",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: multipleStepEnv}},
			},
			expectedKey: "mZk8q7zjJioxI2nTwrt7XQ",
		},
		{
			name: "ExecutionKey null step env",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: nullStepEnv}},
			},
			expectedKey: "_txGuv3XrkWWVQz6hGsKhw",
		},
		{
			name: "ExecutionKeyWithGlobalEnv simple",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo"}},
				},
				GlobalEnv: []string{},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "ExecutionKeyWithGlobalEnv has global env",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og",
		},
		{
			name: "ExecutionKeyWithGlobalEnv env not updated",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"FAZ=BAZ"},
			},
			expectedKey: "tq9NsiMdvoKqMpgxE00XGQ",
		},
		{
			name: "ExecutionKeyWithGlobalEnv malformed global env",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV"},
			},
			expectedError: errors.New("resolving environment for step 0: unable to parse environment variable \"SOME_ENV\""),
		},
		{
			name: "StepsCacheKey simple",
			keyer: StepsCacheKey{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo"}},
				},
				StepIndex: 0,
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g-step-0",
		},
		{
			name: "StepsCacheKey multiple steps",
			keyer: StepsCacheKey{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps: []batches.Step{
						{Run: "foo"},
						{Run: "bar"},
					},
				},
				StepIndex: 0,
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g-step-0",
		},
		{
			name: "StepsCacheKeyWithGlobalEnv env set",
			keyer: &StepsCacheKeyWithGlobalEnv{
				StepsCacheKey: &StepsCacheKey{
					ExecutionKey: &ExecutionKey{
						Repository: repo,
						Steps: []batches.Step{
							{
								Run: "foo",
								Env: stepEnv,
							},
						},
					},
					StepIndex: 0,
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og-step-0",
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

func TestKeyer_Key_Mount(t *testing.T) {
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
		keyer Keyer
		// Instead of an expectedKey, use raw so the data that will be manually hashed can be controlled.
		expectedRaw   string
		expectedError error
	}{
		{
			name: "ExecutionKey single file",
			keyer: &ExecutionKey{
				Repository: repo,
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
			name: "ExecutionKey multiple files",
			keyer: &ExecutionKey{
				Repository: repo,
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
			name: "ExecutionKey directory",
			keyer: &ExecutionKey{
				Repository: repo,
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
		{
			name: "ExecutionKey file does not exist",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps: []batches.Step{
					{
						Run: "foo",
						Mount: []batches.Mount{{
							Path:       filepath.Join(tempDir, "some-file-does-not-exist.sh"),
							Mountpoint: "/tmp/file.sh",
						}},
					},
				},
			},
			expectedError: errors.Newf("path %s does not exist", filepath.Join(tempDir, "some-file-does-not-exist.sh")),
		},
		{
			name: "ExecutionKeyWithGlobalEnv single file",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
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
			name: "ExecutionKeyWithGlobalEnv multiple files",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
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
			name: "ExecutionKeyWithGlobalEnv directory",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
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
		{
			name: "StepsCacheKey single file",
			keyer: StepsCacheKey{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps: []batches.Step{
						{
							Run: "foo",
							Mount: []batches.Mount{
								{
									Path:       sampleScriptPath,
									Mountpoint: "/tmp/sample.sh",
								},
							},
						},
					},
				},
				StepIndex: 0,
			},
			expectedRaw: fmt.Sprintf(`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":{},"mount":[{"mountpoint":"/tmp/sample.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
				sampleScriptPath,
				string(modDateVal),
			),
		},
		{
			name: "StepsCacheKeyWithGlobalEnv single file",
			keyer: &StepsCacheKeyWithGlobalEnv{
				StepsCacheKey: &StepsCacheKey{
					ExecutionKey: &ExecutionKey{
						Repository: repo,
						Steps: []batches.Step{
							{
								Run: "foo",
								Mount: []batches.Mount{
									{
										Path:       sampleScriptPath,
										Mountpoint: "/tmp/sample.sh",
									},
								},
								Env: stepEnv,
							},
						},
					},
					StepIndex: 0,
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedRaw: fmt.Sprintf(`{"Repository":{"ID":"my-repo","Name":"github.com/sourcegraph/src-cli","BaseRef":"refs/heads/f00b4r","BaseRev":"c0mmit","FileMatches":["baz.go"]},"Path":"","OnlyFetchWorkspace":false,"Steps":[{"run":"foo","env":["SOME_ENV"],"mount":[{"mountpoint":"/tmp/sample.sh","path":"%s"}]}],"BatchChangeAttributes":null,"Environments":[{"SOME_ENV":"FOO"}],"MountsMetadata":[{"Path":"%s","Size":0,"Modified":%s}]}`,
				sampleScriptPath,
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
				// Since the expected key is being built manually, need to know if it is step and update the expected
				// key accordingly.
				switch test.keyer.(type) {
				case StepsCacheKey, *StepsCacheKeyWithGlobalEnv:
					assert.Equal(t, fmt.Sprintf("%s-step-0", expectedKey), key)
				case *ExecutionKey, *ExecutionKeyWithGlobalEnv:
					assert.Equal(t, expectedKey, key)
				default:
					assert.Fail(t, "unexpected Keyer implementation")
				}
			}
		})
	}
}
