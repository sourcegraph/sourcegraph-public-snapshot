package inference

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestEmptyHinters(t *testing.T) {
	testHinters(t,
		hinterTestCase{
			description:        "empty",
			repositoryContents: nil,
			expected:           []config.IndexJobHint{},
		},
	)
}

type hinterTestCase struct {
	description        string
	repositoryContents map[string]string
	expected           []config.IndexJobHint
}

func testHinters(t *testing.T, testCases ...hinterTestCase) {
	for _, testCase := range testCases {
		testHinter(t, testCase)
	}
}

func testHinter(t *testing.T, testCase hinterTestCase) {
	t.Run(testCase.description, func(t *testing.T) {
		service := testService(t, testCase.repositoryContents)

		jobHints, err := service.InferIndexJobHints(
			context.Background(),
			"github.com/test/test",
			"HEAD",
			"",
		)
		if err != nil {
			t.Fatalf("unexpected error inferring job hints: %s", err)
		}
		if diff := cmp.Diff(sortIndexJobHints(testCase.expected), sortIndexJobHints(jobHints)); diff != "" {
			t.Errorf("unexpected index job hints (-want +got):\n%s", diff)
		}
	})
}

func sortIndexJobHints(s []config.IndexJobHint) []config.IndexJobHint {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
