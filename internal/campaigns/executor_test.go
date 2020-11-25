package campaigns

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

type mockRepoArchive struct {
	repo  *graphql.Repository
	files map[string]string
}

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

	tests := []struct {
		name string

		repos    []*graphql.Repository
		archives []mockRepoArchive
		steps    []Step

		executorTimeout time.Duration

		wantFilesChanged map[string][]string
		wantErrInclude   string
	}{
		{
			name:  "success",
			repos: []*graphql.Repository{srcCLIRepo, sourcegraphRepo},
			archives: []mockRepoArchive{
				{repo: srcCLIRepo, files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
				{repo: sourcegraphRepo, files: map[string]string{
					"README.md": "# Sourcegraph README\n",
				}},
			},
			steps: []Step{
				{Run: `echo -e "foobar\n" >> README.md`, Container: "alpine:13"},
				{Run: `[[ -f "main.go" ]] && go fmt main.go || exit 0`, Container: "doesntmatter:13"},
			},
			wantFilesChanged: map[string][]string{
				srcCLIRepo.ID:      []string{"README.md", "main.go"},
				sourcegraphRepo.ID: []string{"README.md"},
			},
		},
		{
			name:  "timeout",
			repos: []*graphql.Repository{srcCLIRepo},
			archives: []mockRepoArchive{
				{repo: srcCLIRepo, files: map[string]string{"README.md": "line 1"}},
			},
			steps: []Step{
				// This needs to be a loop, because when a process goes to sleep
				// it's not interruptible, meaning that while it will receive SIGKILL
				// it won't exit until it had its full night of sleep.
				// So.
				// Instead we take short powernaps.
				{Run: `while true; do echo "zZzzZ" && sleep 0.05; done`, Container: "alpine:13"},
			},
			executorTimeout: 100 * time.Millisecond,
			wantErrInclude:  "execution in github.com/sourcegraph/src-cli failed: Timeout reached. Execution took longer than 100ms.",
		},
		{
			name:  "templated",
			repos: []*graphql.Repository{srcCLIRepo},
			archives: []mockRepoArchive{
				{repo: srcCLIRepo, files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
			},
			steps: []Step{
				{Run: `go fmt main.go`, Container: "doesntmatter:13"},
				{Run: `touch modified-${{ join previous_step.modified_files " " }}.md`, Container: "alpine:13"},
				{Run: `touch added-${{ join previous_step.added_files " " }}`, Container: "alpine:13"},
			},
			wantFilesChanged: map[string][]string{
				srcCLIRepo.ID: []string{"main.go", "modified-main.go.md", "added-modified-main.go.md"},
			},
		},
		{
			name:  "empty",
			repos: []*graphql.Repository{srcCLIRepo},
			archives: []mockRepoArchive{
				{repo: srcCLIRepo, files: map[string]string{
					"README.md": "# Welcome to the README\n",
					"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
				}},
			},
			steps: []Step{
				{Run: `true`, Container: "doesntmatter:13"},
			},
			// No changesets should be generated.
			wantFilesChanged: map[string][]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(newZipArchivesMux(t, nil, tc.archives...))
			defer ts.Close()

			var clientBuffer bytes.Buffer
			client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

			testTempDir, err := ioutil.TempDir("", "executor-integration-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(testTempDir)

			cache := newInMemoryExecutionCache()
			creator := &WorkspaceCreator{dir: testTempDir, client: client}
			opts := ExecutorOpts{
				Cache:       cache,
				Creator:     creator,
				TempDir:     testTempDir,
				Parallelism: runtime.GOMAXPROCS(0),
				Timeout:     tc.executorTimeout,
			}
			if opts.Timeout == 0 {
				opts.Timeout = 30 * time.Second
			}

			// execute contains the actual logic running the tasks on an
			// executor. We'll run this multiple times to cover both the cache
			// and non-cache code paths.
			execute := func() {
				executor := newExecutor(opts, client, featuresAllEnabled())

				template := &ChangesetTemplate{}
				for _, r := range tc.repos {
					executor.AddTask(r, tc.steps, template)
				}

				executor.Start(context.Background())
				specs, err := executor.Wait()
				if tc.wantErrInclude == "" && err != nil {
					t.Fatalf("execution failed: %s", err)
				}
				if err != nil && !strings.Contains(err.Error(), tc.wantErrInclude) {
					t.Errorf("wrong error. have=%q want included=%q", err, tc.wantErrInclude)
				}
				if tc.wantErrInclude != "" {
					return
				}

				if have, want := len(specs), len(tc.wantFilesChanged); have != want {
					t.Fatalf("wrong number of changeset specs. want=%d, have=%d", want, have)
				}

				for _, spec := range specs {
					if have, want := len(spec.Commits), 1; have != want {
						t.Fatalf("wrong number of commits. want=%d, have=%d", want, have)
					}

					fileDiffs, err := diff.ParseMultiFileDiff([]byte(spec.Commits[0].Diff))
					if err != nil {
						t.Fatalf("failed to parse diff: %s", err)
					}

					wantFiles, ok := tc.wantFilesChanged[spec.BaseRepository]
					if !ok {
						t.Fatalf("unexpected file changes in repo %s", spec.BaseRepository)
					}

					if have, want := len(fileDiffs), len(wantFiles); have != want {
						t.Fatalf("repo %s: wrong number of fileDiffs. want=%d, have=%d", spec.BaseRepository, want, have)
					}

					diffsByName := map[string]*diff.FileDiff{}
					for _, fd := range fileDiffs {
						if fd.NewName == "/dev/null" {
							diffsByName[fd.OrigName] = fd
						} else {
							diffsByName[fd.NewName] = fd
						}
					}

					for _, file := range wantFiles {
						if _, ok := diffsByName[file]; !ok {
							t.Errorf("%s was not changed (diffsByName=%#v)", file, diffsByName)
						}
					}
				}
			}

			verifyCache := func() {
				want := len(tc.repos)
				if tc.wantErrInclude != "" {
					want = 0
				}

				// Verify that there is a cache entry for each repo.
				if have := cache.size(); have != want {
					t.Errorf("unexpected number of cache entries: have=%d want=%d cache=%+v", have, want, cache)
				}
			}

			// Sanity check, since we're going to be looking at the side effects
			// on the cache.
			if cache.size() != 0 {
				t.Fatalf("unexpectedly hot cache: %+v", cache)
			}

			// Run with a cold cache.
			t.Run("cold cache", func(t *testing.T) {
				execute()
				verifyCache()
			})

			// Run with a warm cache.
			t.Run("warm cache", func(t *testing.T) {
				execute()
				verifyCache()
			})
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

func newZipArchivesMux(t *testing.T, callback http.HandlerFunc, archives ...mockRepoArchive) *http.ServeMux {
	mux := http.NewServeMux()

	for _, archive := range archives {
		files := archive.files
		path := fmt.Sprintf("/%s@%s/-/raw", archive.repo.Name, archive.repo.BaseRef())

		downloadName := filepath.Base(archive.repo.Name)
		mediaType := mime.FormatMediaType("Attachment", map[string]string{
			"filename": downloadName,
		})

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", mediaType)

			zipWriter := zip.NewWriter(w)
			for name, body := range files {
				f, err := zipWriter.Create(name)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := f.Write([]byte(body)); err != nil {
					t.Errorf("failed to write body for %s to zip: %s", name, err)
				}

				if callback != nil {
					callback(w, r)
				}
			}
			if err := zipWriter.Close(); err != nil {
				t.Fatalf("closing zipWriter failed: %s", err)
			}
		})
	}

	return mux
}

// inMemoryExecutionCache provides an in-memory cache for testing purposes.
type inMemoryExecutionCache struct {
	cache map[string][]byte
	mu    sync.RWMutex
}

func newInMemoryExecutionCache() *inMemoryExecutionCache {
	return &inMemoryExecutionCache{
		cache: make(map[string][]byte),
	}
}

func (c *inMemoryExecutionCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *inMemoryExecutionCache) Get(ctx context.Context, key ExecutionCacheKey) (*ChangesetSpec, error) {
	k, err := key.Key()
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if raw, ok := c.cache[k]; ok {
		var spec ChangesetSpec
		if err := json.Unmarshal(raw, &spec); err != nil {
			return nil, err
		}

		return &spec, nil
	}
	return nil, nil
}

func (c *inMemoryExecutionCache) Set(ctx context.Context, key ExecutionCacheKey, spec *ChangesetSpec) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	v, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[k] = v
	return nil
}

func (c *inMemoryExecutionCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	k, err := key.Key()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, k)
	return nil
}
