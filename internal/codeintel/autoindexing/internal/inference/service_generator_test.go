package inference

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestEmptyGenerators(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description:        "empty",
			repositoryContents: nil,
			expected:           []config.IndexJob{},
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
			expected: []config.IndexJob{
				// sg.test -> emits jobs with `test` indexer
				{Indexer: "test", Root: ""},
				{Indexer: "test", Root: "bar"},
				{Indexer: "test", Root: "baz"},
				{Indexer: "test", Root: "foo"},
				// mycompany.test -> emits jobs with `text-override` indexer
				{Indexer: "test-override", Root: ""},
				{Indexer: "test-override", Root: "bar"},
				{Indexer: "test-override", Root: "baz"},
				{Indexer: "test-override", Root: "foo"},
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
			expected: []config.IndexJob{
				// sg.test -> emits jobs with `test` indexer
				// No jobs should have been generated

				// acme.custom -> emits jobs with `acme/custom-indexer` indexer
				{Indexer: "acme/custom-indexer", Root: ""},
				{Indexer: "acme/custom-indexer", Root: "bar"},
				{Indexer: "acme/custom-indexer", Root: "baz"},
				{Indexer: "acme/custom-indexer", Root: "foo"},
			},
		},
	)
}

type generatorTestCase struct {
	description        string
	overrideScript     string
	repositoryContents map[string]string
	expected           []config.IndexJob
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
		if diff := cmp.Diff(sortIndexJobs(testCase.expected), sortIndexJobs(result.IndexJobs)); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	})
}

func sortIndexJobs(s []config.IndexJob) []config.IndexJob {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
