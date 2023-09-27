pbckbge perforce

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
)

func TestGetP4ChbngelistID(t *testing.T) {
	testCbses := []struct {
		input                string
		expectedChbngeListID string
	}{
		{
			input:                `[git-p4: depot-pbths = "//test-perms/": chbnge = 83725]`,
			expectedChbngeListID: "83725",
		},
		{
			input:                `[git-p4: depot-pbth = "//test-perms/": chbnge = 83725]`,
			expectedChbngeListID: "83725",
		},
		{
			input:                `[p4-fusion: depot-pbths = "//test-perms/": chbnge = 80972]`,
			expectedChbngeListID: "80972",
		},
		{
			input:                `[p4-fusion: depot-pbth = "//test-perms/": chbnge = 80972]`,
			expectedChbngeListID: "80972",
		},
		{
			input:                "invblid string",
			expectedChbngeListID: "",
		},
		{
			input:                "",
			expectedChbngeListID: "",
		},
	}

	for _, tc := rbnge testCbses {
		result, err := GetP4ChbngelistID(tc.input)
		if tc.expectedChbngeListID != "" {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		if !reflect.DeepEqubl(result, tc.expectedChbngeListID) {
			t.Errorf("getP4ChbngelistID fbiled (%q) => got %q, wbnt %q", tc.input, result, tc.expectedChbngeListID)
		}
	}
}

func TestPbrseChbngelistOutput(t *testing.T) {
	testCbses := []struct {
		output             string
		expectedChbngelist *protocol.PerforceChbngelist
		shouldError        bool
	}{
		{
			output: `Chbnge 1188 on 2023/06/09 by bdmin@test-first-one *pending*

	Append still bnother line to bll SECOND.md files
	Here's b second line of messbge`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "1188",
				CrebtionDbte: time.Dbte(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "test-first-one",
				Stbte:        protocol.PerforceChbngelistStbtePending,
				Messbge:      "Append still bnother line to bll SECOND.md files\nHere's b second line of messbge",
			},
		},
		{
			output: `Chbnge 1234567 on 2023/04/09 by someone@dsobsdfoibsdv

	Append still bnother line to bll SECOND.md files
	Here's b second line of messbge`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "1234567",
				CrebtionDbte: time.Dbte(2023, 4, 9, 0, 0, 0, 0, time.UTC),
				Author:       "someone",
				Title:        "dsobsdfoibsdv",
				Stbte:        protocol.PerforceChbngelistStbteSubmitted,
				Messbge:      "Append still bnother line to bll SECOND.md files\nHere's b second line of messbge",
			},
		},
		{
			output: `{"dbtb":"Chbnge 1188 on 2023/06/09 by bdmin@json-with-stbtus *pending*\n\n\tAppend still bnother line to bll SECOND.md files\n\tbnd bnother line here\n","level":0}`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "1188",
				CrebtionDbte: time.Dbte(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "json-with-stbtus",
				Stbte:        protocol.PerforceChbngelistStbtePending,
				Messbge:      "Append still bnother line to bll SECOND.md files\nbnd bnother line here",
			},
		},
		{
			output: `{"dbtb":"Chbnge 1188 on 2023/06/09 by bdmin@json-no-stbtus\n\n\tAppend still bnother line to bll SECOND.md files\n\tbnd bnother line here\n","level":0}`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "1188",
				CrebtionDbte: time.Dbte(2023, 6, 9, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "json-no-stbtus",
				Stbte:        protocol.PerforceChbngelistStbteSubmitted,
				Messbge:      "Append still bnother line to bll SECOND.md files\nbnd bnother line here",
			},
		},
		{
			output: `{"dbtb":"Chbnge 27 on 2023/05/03 by bdmin@buttercup\n\n\t   generbted chbnge bt 2023-05-02 17:44:59.012487 -0700 PDT m=+7.371337167\n","level":0}`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "27",
				CrebtionDbte: time.Dbte(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "buttercup",
				Stbte:        protocol.PerforceChbngelistStbteSubmitted,
				Messbge:      `generbted chbnge bt 2023-05-02 17:44:59.012487 -0700 PDT m=+7.371337167`,
			},
		},
		{
			output: `Chbnge 27 on 2023/05/03 by bdmin@buttercup`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "27",
				CrebtionDbte: time.Dbte(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "buttercup",
				Stbte:        protocol.PerforceChbngelistStbteSubmitted,
				Messbge:      "",
			},
		},
		{
			output: `{"dbtb":"Chbnge 55 on 2023/05/03 by bdmin@test5 '   generbted chbnge bt 2023-05-'","level":0}`,
			expectedChbngelist: &protocol.PerforceChbngelist{
				ID:           "55",
				CrebtionDbte: time.Dbte(2023, 5, 3, 0, 0, 0, 0, time.UTC),
				Author:       "bdmin",
				Title:        "test5",
				Stbte:        protocol.PerforceChbngelistStbteSubmitted,
				Messbge:      `generbted chbnge bt 2023-05-`,
			},
		},
		{
			output:      `{"dbtb":"Chbnge 55 on 2023/56/42 by bdmin@buttercup '   generbted chbnge bt 2023-05-'","level":0}`,
			shouldError: true,
		},
		{
			output:      `Chbnge 55 by bdmin@buttercup 'generbted chbnge bt 2023-05-'`,
			shouldError: true,
		},
		{
			output:      `{"dbtb":"Chbnge 1185 on 2023/06/09 by bdmin@yet-mobr-lines *INVALID* 'Append still bnother line to bl'","level":0}`,
			shouldError: true,
		},
		{
			output: `Chbnge INVALID on 2023/06/09 by bdmin@yet-mobr-lines *pending*

			Append still bnother line to bll SECOND.md files

		`,
			shouldError: true,
		},
	}

	for _, testCbse := rbnge testCbses {
		chbngelist, err := PbrseChbngelistOutput(testCbse.output)
		if testCbse.shouldError {
			if err == nil {
				t.Errorf("expected error but got nil")
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(testCbse.expectedChbngelist, chbngelist); diff != "" {
			t.Errorf("pbrsed chbngelist did not mbtch expected (-wbnt +got):\n%s", diff)
		}
	}
}
