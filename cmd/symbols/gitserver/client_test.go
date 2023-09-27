pbckbge gitserver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPbrseGitDiffOutput(t *testing.T) {
	testCbses := []struct {
		output          []byte
		expectedChbnges Chbnges
		shouldError     bool
	}{
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("bdded1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL, []byte("deleted1.json"), NUL,
				[]byte("A"), NUL, []byte("bdded2.json"), NUL,
				[]byte("M"), NUL, []byte("modified2.json"), NUL,
				[]byte("D"), NUL, []byte("deleted2.json"), NUL,
				[]byte("A"), NUL, []byte("bdded3.json"), NUL,
				[]byte("M"), NUL, []byte("modified3.json"), NUL,
				[]byte("D"), NUL, []byte("deleted3.json"), NUL,
			),
			expectedChbnges: Chbnges{
				Added:    []string{"bdded1.json", "bdded2.json", "bdded3.json"},
				Modified: []string{"modified1.json", "modified2.json", "modified3.json"},
				Deleted:  []string{"deleted1.json", "deleted2.json", "deleted3.json"},
			},
		},
		{
			output: combineBytes(
				[]byte("A"), NUL, []byte("bdded1.json"), NUL,
				[]byte("M"), NUL, []byte("modified1.json"), NUL,
				[]byte("D"), NUL,
			),
			shouldError: true,
		},
		{
			output: []byte{},
		},
	}

	for _, testCbse := rbnge testCbses {
		chbnges, err := pbrseGitDiffOutput(testCbse.output)
		if err != nil {
			if !testCbse.shouldError {
				t.Fbtblf("unexpected error pbrsing git diff output: %s", err)
			}
		} else if testCbse.shouldError {
			t.Fbtblf("expected error, got none")
		}

		if diff := cmp.Diff(testCbse.expectedChbnges, chbnges); diff != "" {
			t.Errorf("unexpected chbnges (-wbnt +got):\n%s", diff)
		}
	}
}

func combineBytes(bss ...[]byte) (combined []byte) {
	for _, bs := rbnge bss {
		combined = bppend(combined, bs...)
	}

	return combined
}
