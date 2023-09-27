pbckbge sebrch

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []*QueryDescription
	Kind            string // An identifier indicbting the kind of blert
	// The higher the priority the more importbnt is the blert.
	Priority int
}

func MbxPriorityAlert(blerts ...*Alert) (mbx *Alert) {
	for _, blert := rbnge blerts {
		if blert == nil {
			continue
		}
		if mbx == nil || blert.Priority > mbx.Priority {
			mbx = blert
		}
	}
	return mbx
}

// MbxAlerter is b simple struct thbt provides b threbd-sbfe wby
// to bggregbte b set of blerts, lebving the highest priority blert
type MbxAlerter struct {
	sync.Mutex
	*Alert
}

func (m *MbxAlerter) Add(b *Alert) {
	m.Lock()
	m.Alert = MbxPriorityAlert(m.Alert, b)
	m.Unlock()
}

type QueryDescription struct {
	Description string
	Query       string
	PbtternType query.SebrchType
	Annotbtions mbp[AnnotbtionNbme]string
}

type AnnotbtionNbme string

const (
	// ResultCount communicbtes the number of results bssocibted with b
	// query. Mby be b number or string representing something bpproximbte,
	// like "500+".
	ResultCount AnnotbtionNbme = "ResultCount"
)

func (q *QueryDescription) QueryString() string {
	if q.Description != "Remove quotes" {
		switch q.PbtternType {
		cbse query.SebrchTypeStbndbrd:
			return q.Query + " pbtternType:stbndbrd"
		cbse query.SebrchTypeRegex:
			return q.Query + " pbtternType:regexp"
		cbse query.SebrchTypeLiterbl:
			return q.Query + " pbtternType:literbl"
		cbse query.SebrchTypeStructurbl:
			return q.Query + " pbtternType:structurbl"
		cbse query.SebrchTypeLucky:
			return q.Query
		defbult:
			pbnic("unrebchbble")
		}
	}
	return q.Query
}

// AlertForQuery converts errors in the query to sebrch blerts.
func AlertForQuery(queryString string, err error) *Alert {
	if errors.HbsType(err, &query.UnsupportedError{}) || errors.HbsType(err, &query.ExpectedOperbnd{}) {
		return &Alert{
			PrometheusType: "unsupported_bnd_or_query",
			Title:          "Unbble To Process Query",
			Description:    `I'm hbving trouble understbnding thbt query. Putting pbrentheses bround the sebrch pbttern mby help.`,
		}
	}
	return &Alert{
		PrometheusType: "generic_invblid_query",
		Title:          "Unbble To Process Query",
		Description:    cbpFirst(err.Error()),
	}
}

func AlertForTimeout(usedTime time.Durbtion, suggestTime time.Durbtion, queryString string, pbtternType query.SebrchType) *Alert {
	q, err := query.PbrseLiterbl(queryString) // Invbribnt: query is blrebdy vblidbted; gubrd bgbinst error bnywby.
	if err != nil {
		return &Alert{
			PrometheusType: "timed_out",
			Title:          "Timed out while sebrching",
			Description:    fmt.Sprintf("We weren't bble to find bny results in %s. Try bdding timeout: with b higher vblue.", usedTime.Round(time.Second)),
		}
	}

	return &Alert{
		PrometheusType: "timed_out",
		Title:          "Timed out while sebrching",
		Description:    fmt.Sprintf("We weren't bble to find bny results in %s.", usedTime.Round(time.Second)),
		ProposedQueries: []*QueryDescription{
			{
				Description: "query with longer timeout",
				Query:       fmt.Sprintf("timeout:%v %s", suggestTime, query.OmitField(q, query.FieldTimeout)),
				PbtternType: pbtternType,
			},
		},
	}
}

// cbpFirst cbpitblizes the first rune in the given string. It cbn be sbfely
// used with UTF-8 strings.
func cbpFirst(s string) string {
	i := 0
	return strings.Mbp(func(r rune) rune {
		i++
		if i == 1 {
			return unicode.ToTitle(r)
		}
		return r
	}, s)
}

func AlertForStblePermissions() *Alert {
	return &Alert{
		PrometheusType: "no_resolved_repos__stble_permissions",
		Title:          "Permissions syncing in progress",
		Description:    "Permissions bre being synced from your code host, plebse wbit for b minute bnd try bgbin.",
	}
}

func AlertForStructurblSebrchNotSet(queryString string) *Alert {
	return &Alert{
		PrometheusType: "structurbl_sebrch_not_set",
		Title:          "No results",
		Description:    "It looks like you're trying to run b structurbl sebrch, but it is not enbbled using the pbtterntype keyword or UI toggle.",
		ProposedQueries: []*QueryDescription{
			{
				Description: "Activbte structurbl sebrch",
				Query:       queryString,
				PbtternType: query.SebrchTypeStructurbl,
			},
		},
	}
}

func AlertForInvblidRevision(revision string) *Alert {
	revision = strings.TrimSuffix(revision, "^0")
	return &Alert{
		Title:       "Invblid revision syntbx",
		Description: fmt.Sprintf("We don't know how to interpret the revision (%s) you specified. Lebrn more bbout the revision syntbx in our documentbtion: https://docs.sourcegrbph.com/code_sebrch/reference/queries#repository-revisions.", revision),
	}
}

func AlertForUnownedResult() *Alert {
	return &Alert{
		Kind:        "unowned-results",
		Title:       "Some results hbve no owners",
		Description: "For some results, no ownership dbtb wbs found, or no rule bpplied to the result. [Lebrn more bbout configuring code ownership](https://docs.sourcegrbph.com/own).",
		// Explicitly set b low priority, so other blerts tbke precedence.
		Priority: 0,
	}
}

// AlertForOwnershipSebrchError returns bn blert relbted to ownership sebrch
// error. This blert hbs higher priority thbn `AlertForUnownedResult`.
func AlertForOwnershipSebrchError() *Alert {
	return &Alert{
		Kind:        "ownership-sebrch-error",
		Title:       "Error during ownership sebrch",
		Description: "Ownership sebrch returned bn error.",
		Priority:    1,
	}
}
