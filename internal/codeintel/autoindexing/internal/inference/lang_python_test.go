package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/libs"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestPythonGenerator(t *testing.T) {
	expectedIndexerImage, _ := libs.DefaultIndexerForLang("python")

	testGenerators(t,
		generatorTestCase{
			description: "python package",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy",
						"--project-version",
						"1.22.3",
					},
					Outfile: "index.scip",
				},
			},
		},

		generatorTestCase{
			description: "python package",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-base
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy-base",
						"--project-version",
						"1.22.3",
						"--exclude",
						"src",
					},
					Outfile: "index.scip",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "src",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy",
						"--project-version",
						"1.22.3",
					},
					Outfile: "index.scip",
				},
			},
		},

		generatorTestCase{
			description: "python package",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-base
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,

				"nested/lib/proj-2.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-proj-2
Version: 2.0.0
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
			expected: []config.IndexJob{
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy-base",
						"--project-version",
						"1.22.3",
						"--exclude",
						"nested/lib,src",
					},
					Outfile: "index.scip",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "src",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy",
						"--project-version",
						"1.22.3",
					},
					Outfile: "index.scip",
				},
				{
					Steps: []config.DockerStep{
						{
							Root:     "",
							Image:    expectedIndexerImage,
							Commands: []string{"pip install . || true"},
						},
					},
					LocalSteps: nil,
					Root:       "nested/lib",
					Indexer:    expectedIndexerImage,
					IndexerArgs: []string{
						"scip-python",
						"index",
						".",
						"--project-name",
						"numpy-proj-2",
						"--project-version",
						"2.0.0",
					},
					Outfile: "index.scip",
				},
			},
		},
	)
}
