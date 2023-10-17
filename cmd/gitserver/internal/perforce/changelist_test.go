package perforce

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
)

func TestParseChangelistOutput(t *testing.T) {
	created := time.Unix(1629179137, 0)
	testCases := []struct {
		output             string
		expectedChangelist *p4types.Changelist
		shouldError        bool
	}{
		{
			output: `{"change":"10","changeType":"public","client":"tester","desc":"test-first-one\nAppend still another line to all SECOND.md files\nHere's a second line of message\n","path":"//go/src/...","status":"pending","time":"1629179137","user":"admin"}`,
			expectedChangelist: &p4types.Changelist{
				ID:           "10",
				CreationDate: created,
				Author:       "admin",
				Title:        "tester",
				State:        p4types.ChangelistStatePending,
				Message:      "test-first-one\nAppend still another line to all SECOND.md files\nHere's a second line of message",
			},
		},
		{
			output:      `{"change":123}`,
			shouldError: true,
		},
		{
			output:      `definitelynotjson`,
			shouldError: true,
		},
		{
			output:      `{"change":"1188","changeType":"public","client":"tester","desc":"test","path":"//go/src/...","status":"INVALID","time":"1629179137","user":"admin"}`,
			shouldError: true,
		},
	}

	for _, testCase := range testCases {
		changelist, err := parseChangelistOutput([]byte(testCase.output))
		if testCase.shouldError {
			if err == nil {
				t.Errorf("expected error but got nil")
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(testCase.expectedChangelist, changelist); diff != "" {
			t.Errorf("parsed changelist did not match expected (-want +got):\n%s", diff)
		}
	}
}
