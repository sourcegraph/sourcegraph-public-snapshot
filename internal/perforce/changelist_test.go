package perforce

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func TestGetP4ChangelistID(t *testing.T) {
	testCases := []struct {
		input                string
		expectedChangeListID string
	}{
		{
			input:                `[git-p4: depot-paths = "//test-perms/": change = 83725]`,
			expectedChangeListID: "83725",
		},
		{
			input:                `[git-p4: depot-path = "//test-perms/": change = 83725]`,
			expectedChangeListID: "83725",
		},
		{
			input:                `[p4-fusion: depot-paths = "//test-perms/": change = 80972]`,
			expectedChangeListID: "80972",
		},
		{
			input:                `[p4-fusion: depot-path = "//test-perms/": change = 80972]`,
			expectedChangeListID: "80972",
		},
		{
			input:                "invalid string",
			expectedChangeListID: "",
		},
		{
			input:                "",
			expectedChangeListID: "",
		},
	}

	for _, tc := range testCases {
		result, err := GetP4ChangelistID(tc.input)
		if tc.expectedChangeListID != "" {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		if !reflect.DeepEqual(result, tc.expectedChangeListID) {
			t.Errorf("getP4ChangelistID failed (%q) => got %q, want %q", tc.input, result, tc.expectedChangeListID)
		}
	}
}

func TestParseChangelistOutput(t *testing.T) {
	testCases := []struct {
		output             string
		expectedChangelist *protocol.PerforceChangelist
		shouldError        bool
	}{
		{
			output: `Change 1188 on 2023/06/09 by admin@test-first-one *pending*

	Append still another line to all SECOND.md files
	Here's a second line of message`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "1188",
				CreationDate: time.Date(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "test-first-one",
				State:        protocol.PerforceChangelistStatePending,
				Message:      "Append still another line to all SECOND.md files\nHere's a second line of message",
			},
		},
		{
			output: `Change 1234567 on 2023/04/09 by someone@dsobsdfoibsdv

	Append still another line to all SECOND.md files
	Here's a second line of message`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "1234567",
				CreationDate: time.Date(2023, 4, 9, 0, 0, 0, 0, time.UTC),
				Author:       "someone",
				Title:        "dsobsdfoibsdv",
				State:        protocol.PerforceChangelistStateSubmitted,
				Message:      "Append still another line to all SECOND.md files\nHere's a second line of message",
			},
		},
		{
			output: `{"data":"Change 1188 on 2023/06/09 by admin@json-with-status *pending*\n\n\tAppend still another line to all SECOND.md files\n\tand another line here\n","level":0}`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "1188",
				CreationDate: time.Date(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "json-with-status",
				State:        protocol.PerforceChangelistStatePending,
				Message:      "Append still another line to all SECOND.md files\nand another line here",
			},
		},
		{
			output: `{"data":"Change 1188 on 2023/06/09 by admin@json-no-status\n\n\tAppend still another line to all SECOND.md files\n\tand another line here\n","level":0}`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "1188",
				CreationDate: time.Date(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "json-no-status",
				State:        protocol.PerforceChangelistStateSubmitted,
				Message:      "Append still another line to all SECOND.md files\nand another line here",
			},
		},
		{
			output: `{"data":"Change 27 on 2023/05/03 by admin@buttercup\n\n\t   generated change at 2023-05-02 17:44:59.012487 -0700 PDT m=+7.371337167\n","level":0}`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "27",
				CreationDate: time.Date(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "buttercup",
				State:        protocol.PerforceChangelistStateSubmitted,
				Message:      `generated change at 2023-05-02 17:44:59.012487 -0700 PDT m=+7.371337167`,
			},
		},
		{
			output: `Change 27 on 2023/05/03 by admin@buttercup`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "27",
				CreationDate: time.Date(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "buttercup",
				State:        protocol.PerforceChangelistStateSubmitted,
				Message:      "",
			},
		},
		{
			output: `{"data":"Change 55 on 2023/05/03 by admin@test5 '   generated change at 2023-05-'","level":0}`,
			expectedChangelist: &protocol.PerforceChangelist{
				ID:           "55",
				CreationDate: time.Date(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "admin",
				Title:        "test5",
				State:        protocol.PerforceChangelistStateSubmitted,
				Message:      `generated change at 2023-05-`,
			},
		},
		{
			output:      `{"data":"Change 55 on 2023/56/42 by admin@buttercup '   generated change at 2023-05-'","level":0}`,
			shouldError: true,
		},
		{
			output:      `Change 55 by admin@buttercup 'generated change at 2023-05-'`,
			shouldError: true,
		},
		{
			output:      `{"data":"Change 1185 on 2023/06/09 by admin@yet-moar-lines *INVALID* 'Append still another line to al'","level":0}`,
			shouldError: true,
		},
		{
			output: `Change INVALID on 2023/06/09 by admin@yet-moar-lines *pending*

			Append still another line to all SECOND.md files

		`,
			shouldError: true,
		},
	}

	for _, testCase := range testCases {
		changelist, err := ParseChangelistOutput(testCase.output)
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
