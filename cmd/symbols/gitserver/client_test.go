package gitserver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseGitDiffOutput(t *testing.T) {
	testCases := []struct {
		output          []byte
		expectedChanges Changes
		shouldError     bool
	}{
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("added1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL, []byte("deleted1.json"), NUL,
				[]byte("A"), NUL, []byte("added2.json"), NUL,
				[]byte("M"), NUL, []byte("modified2.json"), NUL,
				[]byte("D"), NUL, []byte("deleted2.json"), NUL,
				[]byte("A"), NUL, []byte("added3.json"), NUL,
				[]byte("M"), NUL, []byte("modified3.json"), NUL,
				[]byte("D"), NUL, []byte("deleted3.json"), NUL,
			),
			expectedChanges: Changes{
				Added:    []string{"added1.json", "added2.json", "added3.json"},
				Modified: []string{"modified1.json", "modified2.json", "modified3.json"},
				Deleted:  []string{"deleted1.json", "deleted2.json", "deleted3.json"},
			},
		},
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("added1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL,
			),
			shouldError: true,
		},
		{
			output: []byte{},
		},
	}

	for _, testCase := range testCases {
		changes, err := parseGitDiffOutput(testCase.output)
		if err != nil {
			if !testCase.shouldError {
				t.Fatalf("unexpected error parsing git diff output: %s", err)
			}
		} else if testCase.shouldError {
			t.Fatalf("expected error, got none")
		}

		if diff := cmp.Diff(testCase.expectedChanges, changes); diff != "" {
			t.Errorf("unexpected changes (-want +got):\n%s", diff)
		}
	}
}

func combineBytes(bss ...[]byte) (combined []byte) {
	for _, bs := range bss {
		combined = append(combined, bs...)
	}

	return combined
}
