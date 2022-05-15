package inference

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

type generatorTestCase struct {
	description        string
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

		jobs, err := service.InferIndexJobs(
			context.Background(),
			api.RepoName("github.com/test/test"),
			"HEAD",
			// To be implemented in // See https://github.com/sourcegraph/sourcegraph/issues/33046
			"",
		)
		if err != nil {
			t.Fatalf("unexpected error inferring jobs: %s", err)
		}
		if diff := cmp.Diff(sortIndexJobs(testCase.expected), sortIndexJobs(jobs)); diff != "" {
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
