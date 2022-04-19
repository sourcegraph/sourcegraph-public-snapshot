package inference

import (
	"context"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRecognizers(t *testing.T) {
	testCases := []struct {
		description        string
		repositoryContents map[string]string
		expected           []config.IndexJob
	}{
		{
			description:        "empty",
			repositoryContents: nil,
			expected:           nil,
		},

		// Rust recognizers

		{
			description: "rust-analyzer",
			repositoryContents: map[string]string{
				"foo/bar/Cargo.toml": "",
				"foo/baz/Cargo.toml": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-rust",
					IndexerArgs: []string{"lsif-rust", "index"},
					Outfile:     "dump.lsif",
				},
			},
		},

		// Go recognizers

		{
			description: "go modules",
			repositoryContents: map[string]string{
				"foo/bar/go.mod": "",
				"foo/baz/go.mod": "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "foo/bar",
							Image:    "sourcegraph/lsif-go:latest",
							Commands: []string{"go mod download"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/bar",
					Indexer:     "sourcegraph/lsif-go:latest",
					IndexerArgs: []string{"lsif-go", "--no-animation"},
					Outfile:     "",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "foo/baz",
							Image:    "sourcegraph/lsif-go:latest",
							Commands: []string{"go mod download"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/baz",
					Indexer:     "sourcegraph/lsif-go:latest",
					IndexerArgs: []string{"lsif-go", "--no-animation"},
					Outfile:     "",
				},
			},
		},
		{
			description: "go files in root",
			repositoryContents: map[string]string{
				"main.go":       "",
				"internal/a.go": "",
				"internal/b.go": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-go:latest",
					IndexerArgs: []string{"GO111MODULE=off", "lsif-go", "--no-animation"},
					Outfile:     "",
				},
			},
		},
		{
			description: "go files in non-root (no match)",
			repositoryContents: map[string]string{
				"cmd/src/main.go": "",
			},
			expected: nil,
		},

		// Java recognizers

		{
			description: "java project with lsif-java.json",
			repositoryContents: map[string]string{
				"lsif-java.json": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-java",
					IndexerArgs: []string{"lsif-java", "index", "--build-tool=lsif"},
					Outfile:     "dump.lsif",
				},
			},
		},
		{
			description: "java project without lsif-java.json (no match)",
			repositoryContents: map[string]string{
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
			expected: nil,
		},

		// Typescript recognizers

		{
			description: "javascript project with no tsconfig",
			repositoryContents: map[string]string{
				"package.json": "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install --ignore-scripts"},
						},
					},
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index", "--infer-tsconfig"},
					Outfile:     "",
				},
			},
		},
		{
			description: "javascript project with no tsconfig",
			repositoryContents: map[string]string{
				"package.json": "",
				"yarn.lock":    "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"yarn --ignore-engines --ignore-scripts"},
						},
					},
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index", "--infer-tsconfig"},
					Outfile:     "",
				},
			},
		},
		{
			description: "simple tsconfig",
			repositoryContents: map[string]string{
				"tsconfig.json": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
			},
		},
		{
			description: "tsconfig in subdirectories",
			repositoryContents: map[string]string{
				"a/tsconfig.json": "",
				"b/tsconfig.json": "",
				"c/tsconfig.json": "",
			},
			expected: []config.IndexJob{
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "a",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "b",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
				{
					Steps:       nil,
					LocalSteps:  nil,
					Root:        "c",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
			},
		},
		{
			description: "typescript installation steps",
			repositoryContents: map[string]string{
				"tsconfig.json":              "",
				"package.json":               "",
				"foo/bar/package.json":       "",
				"foo/bar/yarn.lock":          "",
				"foo/bar/baz/tsconfig.json":  "",
				"foo/bar/bonk/tsconfig.json": "",
				"foo/bar/bonk/package.json":  "",
				"foo/baz/tsconfig.json":      "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install"},
						},
						{
							Root:     "foo/bar",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"yarn --ignore-engines"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/bar/baz",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install"},
						},
						{
							Root:     "foo/bar",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"yarn --ignore-engines"},
						},
						{
							Root:     "foo/bar/bonk",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/bar/bonk",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:  nil,
					Root:        "foo/baz",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
			},
		},
		{
			description: "typescript with lerna configuration",
			repositoryContents: map[string]string{
				"package.json":  "",
				"lerna.json":    `{"npmClient": "yarn"}`,
				"tsconfig.json": "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"yarn --ignore-engines"},
						},
					},
					LocalSteps:  nil,
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
			},
		},
		{
			description: "typescript with node version",
			repositoryContents: map[string]string{
				"package.json":  `{"engines": {"node": "42"}}`,
				"tsconfig.json": "",
				".nvmrc":        "",
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    "sourcegraph/lsif-typescript:autoindex",
							Commands: []string{"N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto", "npm install"},
						},
					},
					LocalSteps:  []string{"N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"},
					Root:        "",
					Indexer:     "sourcegraph/lsif-typescript:autoindex",
					IndexerArgs: []string{"lsif-typescript-autoindex", "index"},
					Outfile:     "",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			// Real deal
			sandboxService := luasandbox.GetService()

			// Fake deal
			gitService := NewMockGitService()
			gitService.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (paths []string, _ error) {
				for path := range testCase.repositoryContents {
					if pattern.MatchString(path) {
						paths = append(paths, path)
					}
				}

				return
			})
			gitService.ArchiveFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
				files := map[string]io.Reader{}
				for _, spec := range opts.Pathspecs {
					if contents, ok := testCase.repositoryContents[strings.TrimPrefix(string(spec), ":(literal)")]; ok {
						files[string(spec)] = strings.NewReader(contents)
					}
				}

				return unpacktest.CreateZipArchive(t, files)
			})

			jobs, err := newService(sandboxService, gitService, &observation.TestContext).InferIndexJobs(
				context.Background(),
				api.RepoName("github.com/test/test"),
				"HEAD",
				"", // TODO
			)
			if err != nil {
				t.Fatalf("unexpected error inferring jobs: %s", err)
			}
			if diff := cmp.Diff(sortIndexJobs(testCase.expected), sortIndexJobs(jobs)); diff != "" {
				t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
			}
		})
	}
}

func sortIndexJobs(s []config.IndexJob) []config.IndexJob {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
