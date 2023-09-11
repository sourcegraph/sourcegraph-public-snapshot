package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestTypeScriptGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("typescript")

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
							Image:    expectedIndexerImage,
							Commands: []string{"npm install --ignore-scripts"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index", "--infer-tsconfig"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
							Image:    expectedIndexerImage,
							Commands: []string{"yarn --ignore-scripts"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index", "--infer-tsconfig"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
					Steps:            nil,
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
					Steps:            nil,
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "a",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
				{
					Steps:            nil,
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "b",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
				{
					Steps:            nil,
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "c",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
							Image:    expectedIndexerImage,
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"npm install"},
						},
						{
							Root:     "foo/bar",
							Image:    expectedIndexerImage,
							Commands: []string{"yarn"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "foo/bar/baz",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"npm install"},
						},
						{
							Root:     "foo/bar",
							Image:    expectedIndexerImage,
							Commands: []string{"yarn"},
						},
						{
							Root:     "foo/bar/bonk",
							Image:    expectedIndexerImage,
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "foo/bar/bonk",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"npm install"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "foo/baz",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
							Image:    expectedIndexerImage,
							Commands: []string{"yarn"},
						},
					},
					LocalSteps:       []string{`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
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
							Image:    expectedIndexerImage,
							Commands: []string{"N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto", "npm install"},
						},
					},
					LocalSteps: []string{
						"N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto",
						`if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi`,
					},
					Root:             "",
					Indexer:          expectedIndexerImage,
					IndexerArgs:      []string{"scip-typescript", "index"},
					Outfile:          "index.scip",
					RequestedEnvVars: []string{"NPM_TOKEN"},
				},
			},
		},
	)
}
