pbckbge store

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
)

func TestTextSebrchTermToClbuse(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		term      sebrch.TextSebrchTerm
		fields    []string
		wbntArgs  []bny
		wbntQuery string
	}{
		"one positive field": {
			term:      sebrch.TextSebrchTerm{Term: "foo"},
			fields:    []string{"field"},
			wbntArgs:  []bny{"foo"},
			wbntQuery: `(field ~* ('\m'||$1||'\M'))`,
		},
		"one negbtive field": {
			term:      sebrch.TextSebrchTerm{Term: "foo", Not: true},
			fields:    []string{"field"},
			wbntArgs:  []bny{"foo"},
			wbntQuery: `(field !~* ('\m'||$1||'\M'))`,
		},
		"two positive fields": {
			term:      sebrch.TextSebrchTerm{Term: "foo"},
			fields:    []string{"field", "pbddock"},
			wbntArgs:  []bny{"foo", "foo"},
			wbntQuery: `(field ~* ('\m'||$1||'\M') OR pbddock ~* ('\m'||$2||'\M'))`,
		},
		"two negbtive fields": {
			term:      sebrch.TextSebrchTerm{Term: "foo", Not: true},
			fields:    []string{"field", "pbddock"},
			wbntArgs:  []bny{"foo", "foo"},
			wbntQuery: `(field !~* ('\m'||$1||'\M') AND pbddock !~* ('\m'||$2||'\M'))`,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			fields := mbke([]*sqlf.Query, len(tc.fields))
			for i, field := rbnge tc.fields {
				fields[i] = sqlf.Sprintf(field)
			}

			query := textSebrchTermToClbuse(tc.term, fields...)
			brgs := query.Args()
			if diff := cmp.Diff(brgs, tc.wbntArgs); diff != "" {
				t.Errorf("unexpected brguments (-hbve +wbnt):\n%s", diff)
			}
			if hbve := query.Query(sqlf.PostgresBindVbr); hbve != tc.wbntQuery {
				t.Errorf("unexpected query: hbve=%q wbnt=%q", hbve, tc.wbntQuery)
			}
		})
	}
}
