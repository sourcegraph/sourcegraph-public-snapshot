package inference

import (
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var update = flag.Bool("update", false, "update testdata")

func TestEmptyGenerators(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description:        "empty",
			repositoryContents: nil,
		},
	)
}

func TestOverrideGenerators(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "override",
			overrideScript: `
				local path = require("path")
				local pattern = require("sg.autoindex.patterns")
				local recognizer = require("sg.autoindex.recognizer")

				local custom_recognizer = recognizer.new_path_recognizer {
					patterns = { pattern.new_path_basename("sg-test") },

					-- Invoked when go.mod files exist
					generate = function(_, paths)
						local jobs = {}
						for i = 1, #paths do
							table.insert(jobs, {
								steps = {},
								root = path.dirname(paths[i]),
								indexer = "test-override",
								indexer_args = {},
								outfile = "",
							})
						end

						return jobs
					end,
				}

				return require("sg.autoindex.config").new({
					["custom.test"] = custom_recognizer,
				})
			`,
			repositoryContents: map[string]string{
				"sg-test":     "",
				"foo/sg-test": "",
				"bar/sg-test": "",
				"baz/sg-test": "",
			},
		},
		generatorTestCase{
			description: "disable default",
			overrideScript: `
				local path = require("path")
				local pattern = require("sg.autoindex.patterns")
				local recognizer = require("sg.autoindex.recognizer")

				local custom_recognizer = recognizer.new_path_recognizer {
					patterns = {
						pattern.new_path_basename("acme-custom.yaml")
					},

					-- Invoked with paths matching acme-custom.yaml anywhere in repo
					generate = function(_, paths)
						local jobs = {}
						for i = 1, #paths do
							table.insert(jobs, {
								steps = {},
								root = path.dirname(paths[i]),
								indexer = "acme/custom-indexer",
								indexer_args = {},
								outfile = "",
							})
						end

						return jobs
					end,
				}

				return require("sg.autoindex.config").new({
					["sg.test"] = false,
					["acme.custom"] = custom_recognizer,
				})
			`,
			repositoryContents: map[string]string{
				"acme-custom.yaml":     "",
				"foo/acme-custom.yaml": "",
				"bar/acme-custom.yaml": "",
				"baz/acme-custom.yaml": "",
			},
			// sg.test -> emits jobs with `test` indexer
			// No jobs should have been generated
			// acme.custom -> emits jobs with `acme/custom-indexer` indexer
		},
	)
}

// Run 'go test ./... -update' in this subdirectory to update snapshot outputs
type generatorTestCase struct {
	description        string
	overrideScript     string
	repositoryContents map[string]string
}

func testGenerators(t *testing.T, testCases ...generatorTestCase) {
	for _, testCase := range testCases {
		testGenerator(t, testCase)
	}
}

func testGenerator(t *testing.T, testCase generatorTestCase) {
	t.Run(testCase.description, func(t *testing.T) {
		service := testService(t, testCase.repositoryContents)

		result, err := service.InferIndexJobs(
			context.Background(),
			"github.com/test/test",
			"HEAD",
			testCase.overrideScript,
		)
		if err != nil {
			t.Fatalf("unexpected error inferring jobs: %s", err)
		}
		snapshotPath := filepath.Join("testdata", strings.Replace(testCase.description, " ", "_", -1)+".yaml")
		sortIndexJobs(result.IndexJobs)
		if update != nil && *update == true {
			bytes, err := yaml.Marshal(result.IndexJobs)
			require.NoError(t, err)
			file, err := os.Create(snapshotPath)
			require.NoError(t, err)
			_, err = file.Write(bytes)
			require.NoError(t, err)
			return
		}
		file, err := os.Open(snapshotPath)
		require.NoError(t, err)
		bytes, err := io.ReadAll(file)
		require.NoError(t, err)
		var expected []config.AutoIndexJobSpec
		require.NoError(t, yaml.Unmarshal(bytes, &expected))
		if diff := cmp.Diff(expected, result.IndexJobs, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	})
}

func sortIndexJobs(s []config.AutoIndexJobSpec) []config.AutoIndexJobSpec {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
