package gitserver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPushFalseToSliceAtIndexes(t *testing.T) {
	testCases := []struct {
		description    string
		exists         []bool
		missingIndexes []int
		expected       []bool
	}{
		{
			description:    "no missing indexes",
			exists:         []bool{true, false, true, false},
			missingIndexes: []int{},
			expected:       []bool{true, false, true, false},
		},
		{
			description:    "single missing index",
			exists:         []bool{true, true, true, true},
			missingIndexes: []int{1},
			expected:       []bool{true, false, true, true, true},
		},
		{
			description:    "multiple missing indexes",
			exists:         []bool{true, true, true, true},
			missingIndexes: []int{1, 3, 5},
			expected:       []bool{true, false, true, false, true, false, true},
		},
		{
			description:    "missing index run",
			exists:         []bool{true, true, true, true},
			missingIndexes: []int{1, 2, 3},
			expected:       []bool{true, false, false, false, true, true, true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			result, err := pushFalseToSliceAtIndexes(testCase.exists, testCase.missingIndexes)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(testCase.expected, result); diff != "" {
				t.Fatalf("unexpected results (-want +got):\n%s", diff)
			}
		})
	}
}
