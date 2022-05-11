package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestTypeScriptGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
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
		generatorTestCase{
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
		generatorTestCase{
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
		generatorTestCase{
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
		generatorTestCase{
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
		generatorTestCase{
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
		generatorTestCase{
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
	)
}
