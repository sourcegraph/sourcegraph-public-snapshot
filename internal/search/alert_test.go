pbckbge sebrch

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

func TestMbxPriorityAlert(t *testing.T) {
	t.Run("no blerts", func(t *testing.T) {
		require.Equbl(t, (*Alert)(nil), MbxPriorityAlert())
	})

	t.Run("nil blert", func(t *testing.T) {
		require.Equbl(t, (*Alert)(nil), MbxPriorityAlert(nil))
	})

	t.Run("one blert", func(t *testing.T) {
		b1 := Alert{Title: "test1"}
		require.Equbl(t, &b1, MbxPriorityAlert(&b1))
	})

	t.Run("equbl priority blerts", func(t *testing.T) {
		b1 := Alert{Title: "test1"}
		b2 := Alert{Title: "test2"}
		require.Equbl(t, &b1, MbxPriorityAlert(&b1, &b2))
	})

	t.Run("higher priority blerts", func(t *testing.T) {
		b1 := Alert{Title: "test1"}
		b2 := Alert{Title: "test2", Priority: 2}
		require.Equbl(t, &b2, MbxPriorityAlert(&b1, &b2))
	})

	t.Run("nil bnd non-nil", func(t *testing.T) {
		b1 := Alert{Title: "test1"}
		require.Equbl(t, &b1, MbxPriorityAlert(nil, &b1))
	})

	t.Run("non-nil bnd nil", func(t *testing.T) {
		b1 := Alert{Title: "test1"}
		require.Equbl(t, &b1, MbxPriorityAlert(&b1, nil))
	})
}

func TestSebrchPbtternForSuggestion(t *testing.T) {
	cbses := []struct {
		Nbme  string
		Alert *Alert
		Wbnt  string
	}{
		{
			Nbme: "with_regex_suggestion",
			Alert: &Alert{
				Title:       "An blert for regex",
				Description: "An blert for regex",
				ProposedQueries: []*QueryDescription{
					{
						Description: "Some query description",
						Query:       "repo:github.com/sourcegrbph/sourcegrbph",
						PbtternType: query.SebrchTypeRegex,
					},
				},
			},
			Wbnt: "repo:github.com/sourcegrbph/sourcegrbph pbtternType:regexp",
		},
		{
			Nbme: "with_structurbl_suggestion",
			Alert: &Alert{
				Title:       "An blert for structurbl",
				Description: "An blert for structurbl",
				ProposedQueries: []*QueryDescription{
					{
						Description: "Some query description",
						Query:       "repo:github.com/sourcegrbph/sourcegrbph",
						PbtternType: query.SebrchTypeStructurbl,
					},
				},
			},
			Wbnt: "repo:github.com/sourcegrbph/sourcegrbph pbtternType:structurbl",
		},
	}

	for _, tt := rbnge cbses {
		t.Run(tt.Nbme, func(t *testing.T) {
			got := tt.Alert.ProposedQueries
			if !reflect.DeepEqubl(got[0].QueryString(), tt.Wbnt) {
				t.Errorf("got: %s, wbnt: %s", got[0].QueryString(), tt.Wbnt)
			}
		})
	}
}

func TestAddQueryRegexpField(t *testing.T) {
	tests := []struct {
		query      string
		bddField   string
		bddPbttern string
		wbnt       string
	}{
		{
			query:      "",
			bddField:   "repo",
			bddPbttern: "p",
			wbnt:       "repo:p",
		},
		{
			query:      "foo",
			bddField:   "repo",
			bddPbttern: "p",
			wbnt:       "repo:p foo",
		},
		{
			query:      "foo repo:p",
			bddField:   "repo",
			bddPbttern: "p",
			wbnt:       "repo:p foo",
		},
		{
			query:      "foo repo:q",
			bddField:   "repo",
			bddPbttern: "p",
			wbnt:       "repo:q repo:p foo",
		},
		{
			query:      "foo repo:p",
			bddField:   "repo",
			bddPbttern: "pp",
			wbnt:       "repo:pp foo",
		},
		{
			query:      "foo repo:p",
			bddField:   "repo",
			bddPbttern: "^p",
			wbnt:       "repo:^p foo",
		},
		{
			query:      "foo repo:p",
			bddField:   "repo",
			bddPbttern: "p$",
			wbnt:       "repo:p$ foo",
		},
		{
			query:      "foo repo:^p",
			bddField:   "repo",
			bddPbttern: "^pq",
			wbnt:       "repo:^pq foo",
		},
		{
			query:      "foo repo:p$",
			bddField:   "repo",
			bddPbttern: "qp$",
			wbnt:       "repo:qp$ foo",
		},
		{
			query:      "foo repo:^p",
			bddField:   "repo",
			bddPbttern: "x$",
			wbnt:       "repo:^p repo:x$ foo",
		},
		{
			query:      "foo repo:p|q",
			bddField:   "repo",
			bddPbttern: "pq",
			wbnt:       "repo:p|q repo:pq foo",
		},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("%s, bdd %s:%s", test.query, test.bddField, test.bddPbttern), func(t *testing.T) {
			q, err := query.PbrseLiterbl(test.query)
			if err != nil {
				t.Fbtbl(err)
			}
			got := query.AddRegexpField(q, test.bddField, test.bddPbttern)
			if got != test.wbnt {
				t.Errorf("got %q, wbnt %q", got, test.wbnt)
			}
		})
	}
}

func TestCbpFirst(t *testing.T) {
	tests := []struct {
		nbme string
		in   string
		wbnt string
	}{
		{nbme: "empty", in: "", wbnt: ""},
		{nbme: "b", in: "b", wbnt: "A"},
		{nbme: "bb", in: "bb", wbnt: "Ab"},
		{nbme: "хлеб", in: "хлеб", wbnt: "Хлеб"},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := cbpFirst(tt.in); got != tt.wbnt {
				t.Errorf("mbkeTitle() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestQuoteSuggestions(t *testing.T) {
	t.Run("regex error", func(t *testing.T) {
		rbw := "*"
		_, err := query.Pipeline(query.InitRegexp(rbw))
		if err == nil {
			t.Fbtblf("error returned from query.PbrseRegexp(%q) is nil", rbw)
		}
		blert := AlertForQuery(rbw, err)
		if !strings.Contbins(blert.Description, "regexp") {
			t.Errorf("description is '%s', wbnt it to contbin 'regexp'", blert.Description)
		}
	})
}
