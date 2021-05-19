package executor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

func TestExecutor_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test doesn't work on Windows because dummydocker is written in bash")
	}

	addToPath(t, "testdata/dummydocker")

	srcCLIRepo := &graphql.Repository{
		ID:            "src-cli",
		Name:          "github.com/sourcegraph/src-cli",
		DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	}

	sourcegraphRepo := &graphql.Repository{
		ID:   "sourcegraph",
		Name: "github.com/sourcegraph/sourcegraph",
		DefaultBranch: &graphql.Branch{
			Name:   "main",
			Target: graphql.Target{OID: "f00b4r3r"},
		},
	}

	defaultBatchChangeAttributes := &BatchChangeAttributes{
		Name:        "integration-test-batch-change",
		Description: "this is an integration test",
	}

	const rootPath = ""
	type filesByPath map[string][]string
	type filesByRepository map[string]filesByPath

	tests := []struct {
		name string

		archives        []mock.RepoArchive
		additionalFiles []mock.MockRepoAdditionalFiles

		// We define the steps only once per test case so there's less duplication
		steps []batches.Step
		tasks []*Task

		executorTimeout time.Duration

		wantFilesChanged  filesByRepository
		wantTitle         string
		wantBody          string
		wantCommitMessage string
		wantAuthorName    string
		wantAuthorEmail   string

		wantErrInclude string
	}{
		{
			name: "success",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
				{Repo: sourcegraphRepo, Files: map[string]string{
					"README.md": "# Sourcegraph README\n",
				}},
			},
			steps: []batches.Step{
				{Run: `echo -e "foobar\n" >> README.md`},
				{Run: `[[ -f "main.go" ]] && go fmt main.go || exit 0`},
			},
			tasks: []*Task{
				{Repository: srcCLIRepo},
				{Repository: sourcegraphRepo},
			},
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{"README.md", "main.go"},
				},
				sourcegraphRepo.ID: {
					rootPath: []string{"README.md"},
				},
			},
		},
		{
			name: "empty",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
			},
			steps: []batches.Step{
				{Run: "true"},
			},

			tasks: []*Task{
				{Repository: srcCLIRepo},
			},
			// No diff should be generated.
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{},
				},
			},
		},
		{
			name: "timeout",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{"README.md": "line 1"}},
			},
			steps: []batches.Step{
				// This needs to be a loop, because when a process goes to sleep
				// it's not interruptible, meaning that while it will receive SIGKILL
				// it won't exit until it had its full night of sleep.
				// So.
				// Instead we take short powernaps.
				{Run: `while true; do echo "zZzzZ" && sleep 0.05; done`},
			},
			tasks: []*Task{
				{Repository: srcCLIRepo},
			},
			executorTimeout: 100 * time.Millisecond,
			wantErrInclude:  "execution in github.com/sourcegraph/src-cli failed: Timeout reached. Execution took longer than 100ms.",
		},
		{
			name: "templated steps",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
			},
			steps: []batches.Step{
				{Run: `go fmt main.go`},
				{Run: `touch modified-${{ join previous_step.modified_files " " }}.md`},
				{Run: `touch added-${{ join previous_step.added_files " " }}`},
				{
					Run: `echo "hello.txt"`,
					Outputs: batches.Outputs{
						"myOutput": batches.Output{
							Value: "${{ step.stdout }}",
						},
					},
				},
				{Run: `touch output-${{ outputs.myOutput }}`},
			},

			tasks: []*Task{
				{Repository: srcCLIRepo},
			},
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{
						"main.go",
						"modified-main.go.md",
						"added-modified-main.go.md",
						"output-hello.txt",
					},
				},
			},
		},
		{
			name: "workspaces",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Path: "", Files: map[string]string{
					".gitignore":      "node_modules",
					"message.txt":     "root-dir",
					"a/message.txt":   "a-dir",
					"a/.gitignore":    "node_modules-in-a",
					"a/b/message.txt": "b-dir",
				}},
				{Repo: srcCLIRepo, Path: "a", Files: map[string]string{
					"a/message.txt":   "a-dir",
					"a/.gitignore":    "node_modules-in-a",
					"a/b/message.txt": "b-dir",
				}},
				{Repo: srcCLIRepo, Path: "a/b", Files: map[string]string{
					"a/b/message.txt": "b-dir",
				}},
			},
			additionalFiles: []mock.MockRepoAdditionalFiles{
				{Repo: srcCLIRepo, AdditionalFiles: map[string]string{
					".gitignore":   "node_modules",
					"a/.gitignore": "node_modules-in-a",
				}},
			},
			steps: []batches.Step{
				{
					Run: "cat message.txt && echo 'Hello' > hello.txt",
					Outputs: batches.Outputs{
						"message": batches.Output{
							Value: "${{ step.stdout }}",
						},
					},
				},
				{Run: `if [[ -f ".gitignore" ]]; then echo "yes" >> gitignore-exists; fi`},
				{Run: `if [[ $(basename $(pwd)) == "a" && -f "../.gitignore" ]]; then echo "yes" >> gitignore-exists; fi`},
				// In `a/b` we want the `.gitignore` file in the root folder and in `a` to be fetched:
				{Run: `if [[ $(basename $(pwd)) == "b" && -f "../../.gitignore" ]]; then echo "yes" >> gitignore-exists; fi`},
				{Run: `if [[ $(basename $(pwd)) == "b" && -f "../.gitignore" ]]; then echo "yes" >> gitignore-exists-in-a; fi`},
			},
			tasks: []*Task{
				{Repository: srcCLIRepo, Path: ""},
				{Repository: srcCLIRepo, Path: "a"},
				{Repository: srcCLIRepo, Path: "a/b"},
			},

			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{"hello.txt", "gitignore-exists"},
					"a":      []string{"a/hello.txt", "a/gitignore-exists"},
					"a/b":    []string{"a/b/hello.txt", "a/b/gitignore-exists", "a/b/gitignore-exists-in-a"},
				},
			},
		},
		{
			name: "step condition",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{
					"README.md": "# Welcome to the README\n",
				}},
				{Repo: sourcegraphRepo, Files: map[string]string{
					"README.md": "# Sourcegraph README\n",
				}},
			},
			steps: []batches.Step{
				{Run: `echo -e "foobar\n" >> README.md`},
				{
					Run: `echo "foobar" >> hello.txt`,
					If:  `${{ matches repository.name "github.com/sourcegraph/sourcegra*" }}`,
				},
				{
					Run: `echo "foobar" >> in-path.txt`,
					If:  `${{ matches steps.path "sub/directory/of/repo" }}`,
				},
			},
			tasks: []*Task{
				{Repository: srcCLIRepo},
				{Repository: sourcegraphRepo, Path: "sub/directory/of/repo"},
			},
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{"README.md"},
				},
				sourcegraphRepo.ID: {
					"sub/directory/of/repo": []string{"README.md", "hello.txt", "in-path.txt"},
				},
			},
		},
		{
			name: "skips errors",
			archives: []mock.RepoArchive{
				{Repo: srcCLIRepo, Files: map[string]string{
					"README.md": "# Welcome to the README\n",
				}},
				{Repo: sourcegraphRepo, Files: map[string]string{
					"README.md": "# Sourcegraph README\n",
				}},
			},
			steps: []batches.Step{
				{Run: `echo -e "foobar\n" >> README.md`},
				{
					Run: `exit 1`,
					If:  fmt.Sprintf(`${{ eq repository.name %q }}`, sourcegraphRepo.Name),
				},
			},
			tasks: []*Task{
				{Repository: srcCLIRepo},
				{Repository: sourcegraphRepo},
			},
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByPath{
					rootPath: []string{"README.md"},
				},
				sourcegraphRepo.ID: {},
			},
			wantErrInclude: "execution in github.com/sourcegraph/sourcegraph failed: run: exit 1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Make sure that the steps and tasks are setup properly
			for i := range tc.steps {
				tc.steps[i].SetImage(&mock.Image{
					RawDigest: tc.steps[i].Container,
				})
			}

			for _, task := range tc.tasks {
				task.BatchChangeAttributes = defaultBatchChangeAttributes
				task.Steps = tc.steps
			}

			// Setup a mock test server so we also test the downloading of archives
			mux := mock.NewZipArchivesMux(t, nil, tc.archives...)

			middle := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				})
			}
			for _, additionalFiles := range tc.additionalFiles {
				mock.HandleAdditionalFiles(mux, additionalFiles, middle)
			}

			ts := httptest.NewServer(mux)
			defer ts.Close()

			// Setup an api.Client that points to this test server
			var clientBuffer bytes.Buffer
			client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

			// Temp dir for log files and downloaded archives
			testTempDir, err := ioutil.TempDir("", "executor-integration-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(testTempDir)

			// Setup executor
			opts := newExecutorOpts{
				Creator: workspace.NewCreator(context.Background(), "bind", testTempDir, testTempDir, []batches.Step{}),
				Fetcher: batches.NewRepoFetcher(client, testTempDir, false),
				Logger:  log.NewManager(testTempDir, false),

				TempDir:     testTempDir,
				Parallelism: runtime.GOMAXPROCS(0),
				Timeout:     tc.executorTimeout,
			}

			if opts.Timeout == 0 {
				opts.Timeout = 30 * time.Second
			}

			executor := newExecutor(opts)

			statusHandler := NewTaskStatusCollection([]*Task{})

			// Run executor
			executor.Start(context.Background(), tc.tasks, statusHandler)

			results, err := executor.Wait(context.Background())
			if tc.wantErrInclude == "" {
				if err != nil {
					t.Fatalf("execution failed: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error to include %q, but got no error", tc.wantErrInclude)
				} else {
					if err == nil {
						t.Fatalf("expected error to include %q, but got no error", tc.wantErrInclude)
					} else {
						if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.wantErrInclude)) {
							t.Errorf("wrong error. have=%q want included=%q", err, tc.wantErrInclude)
						}
					}
				}
			}

			wantResults := 0
			resultsFound := map[string]map[string]bool{}
			for repo, byPath := range tc.wantFilesChanged {
				wantResults += len(byPath)
				resultsFound[repo] = map[string]bool{}
				for path := range byPath {
					resultsFound[repo][path] = false
				}
			}

			if have, want := len(results), wantResults; have != want {
				t.Fatalf("wrong number of execution results. want=%d, have=%d", want, have)
			}

			for _, taskResult := range results {
				repoID := taskResult.task.Repository.ID
				path := taskResult.task.Path

				wantFiles, ok := tc.wantFilesChanged[repoID]
				if !ok {
					t.Fatalf("unexpected file changes in repo %s", repoID)
				}

				resultsFound[repoID][path] = true

				wantFilesInPath, ok := wantFiles[path]
				if !ok {
					t.Fatalf("spec for repo %q and path %q but no files expected in that branch", repoID, path)
				}

				fileDiffs, err := diff.ParseMultiFileDiff([]byte(taskResult.result.Diff))
				if err != nil {
					t.Fatalf("failed to parse diff: %s", err)
				}

				if have, want := len(fileDiffs), len(wantFilesInPath); have != want {
					t.Fatalf("repo %s: wrong number of fileDiffs. want=%d, have=%d", repoID, want, have)
				}

				diffsByName := map[string]*diff.FileDiff{}
				for _, fd := range fileDiffs {
					if fd.NewName == "/dev/null" {
						diffsByName[fd.OrigName] = fd
					} else {
						diffsByName[fd.NewName] = fd
					}
				}

				for _, file := range wantFilesInPath {
					if _, ok := diffsByName[file]; !ok {
						t.Errorf("%s was not changed (diffsByName=%#v)", file, diffsByName)
					}
				}
			}

			for repo, paths := range resultsFound {
				for path, found := range paths {
					for !found {
						t.Fatalf("expected spec to be created in path %s of repo %s, but was not", path, repo)
					}
				}
			}
		})
	}
}

func addToPath(t *testing.T, relPath string) {
	t.Helper()

	dummyDockerPath, err := filepath.Abs("testdata/dummydocker")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", dummyDockerPath, os.PathListSeparator, os.Getenv("PATH")))
}

func featuresAllEnabled() batches.FeatureFlags {
	return batches.FeatureFlags{
		AllowArrayEnvironments:   true,
		IncludeAutoAuthorDetails: true,
		UseGzipCompression:       true,
		AllowTransformChanges:    true,
		AllowWorkspaces:          true,
		AllowConditionalExec:     true,
	}
}
