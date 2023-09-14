package search

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []*QueryDescription
	Kind            string // An identifier indicating the kind of alert
	// The higher the priority the more important is the alert.
	Priority int
}

func MaxPriorityAlert(alerts ...*Alert) (max *Alert) {
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		if max == nil || alert.Priority > max.Priority {
			max = alert
		}
	}
	return max
}

// MaxAlerter is a simple struct that provides a thread-safe way
// to aggregate a set of alerts, leaving the highest priority alert
type MaxAlerter struct {
	sync.Mutex
	*Alert
}

func (m *MaxAlerter) Add(a *Alert) {
	m.Lock()
	m.Alert = MaxPriorityAlert(m.Alert, a)
	m.Unlock()
}

type QueryDescription struct {
	Description string
	Query       string
	PatternType query.SearchType
	Annotations map[AnnotationName]string
}

type AnnotationName string

const (
	// ResultCount communicates the number of results associated with a
	// query. May be a number or string representing something approximate,
	// like "500+".
	ResultCount AnnotationName = "ResultCount"
)

func (q *QueryDescription) QueryString() string {
	if q.Description != "Remove quotes" {
		switch q.PatternType {
		case query.SearchTypeStandard:
			return q.Query + " patternType:standard"
		case query.SearchTypeRegex:
			return q.Query + " patternType:regexp"
		case query.SearchTypeLiteral:
			return q.Query + " patternType:literal"
		case query.SearchTypeStructural:
			return q.Query + " patternType:structural"
		case query.SearchTypeLucky:
			return q.Query
		default:
			panic("unreachable")
		}
	}
	return q.Query
}

// AlertForQuery converts errors in the query to search alerts.
func AlertForQuery(queryString string, err error) *Alert {
	if errors.HasType(err, &query.UnsupportedError{}) || errors.HasType(err, &query.ExpectedOperand{}) {
		return &Alert{
			PrometheusType: "unsupported_and_or_query",
			Title:          "Unable To Process Query",
			Description:    `I'm having trouble understanding that query. Putting parentheses around the search pattern may help.`,
		}
	}
	return &Alert{
		PrometheusType: "generic_invalid_query",
		Title:          "Unable To Process Query",
		Description:    capFirst(err.Error()),
	}
}

func AlertForTimeout(usedTime time.Duration, suggestTime time.Duration, queryString string, patternType query.SearchType) *Alert {
	q, err := query.ParseLiteral(queryString) // Invariant: query is already validated; guard against error anyway.
	if err != nil {
		return &Alert{
			PrometheusType: "timed_out",
			Title:          "Timed out while searching",
			Description:    fmt.Sprintf("We weren't able to find any results in %s. Try adding timeout: with a higher value.", usedTime.Round(time.Second)),
		}
	}

	return &Alert{
		PrometheusType: "timed_out",
		Title:          "Timed out while searching",
		Description:    fmt.Sprintf("We weren't able to find any results in %s.", usedTime.Round(time.Second)),
		ProposedQueries: []*QueryDescription{
			{
				Description: "query with longer timeout",
				Query:       fmt.Sprintf("timeout:%v %s", suggestTime, query.OmitField(q, query.FieldTimeout)),
				PatternType: patternType,
			},
		},
	}
}

// capFirst capitalizes the first rune in the given string. It can be safely
// used with UTF-8 strings.
func capFirst(s string) string {
	i := 0
	return strings.Map(func(r rune) rune {
		i++
		if i == 1 {
			return unicode.ToTitle(r)
		}
		return r
	}, s)
}

func AlertForStalePermissions() *Alert {
	return &Alert{
		PrometheusType: "no_resolved_repos__stale_permissions",
		Title:          "Permissions syncing in progress",
		Description:    "Permissions are being synced from your code host, please wait for a minute and try again.",
	}
}

func AlertForStructuralSearchNotSet(queryString string) *Alert {
	return &Alert{
		PrometheusType: "structural_search_not_set",
		Title:          "No results",
		Description:    "It looks like you're trying to run a structural search, but it is not enabled using the patterntype keyword or UI toggle.",
		ProposedQueries: []*QueryDescription{
			{
				Description: "Activate structural search",
				Query:       queryString,
				PatternType: query.SearchTypeStructural,
			},
		},
	}
}

func AlertForInvalidRevision(revision string) *Alert {
	revision = strings.TrimSuffix(revision, "^0")
	return &Alert{
		Title:       "Invalid revision syntax",
		Description: fmt.Sprintf("We don't know how to interpret the revision (%s) you specified. Learn more about the revision syntax in our documentation: https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions.", revision),
	}
}

func AlertForUnownedResult() *Alert {
	return &Alert{
		Kind:        "unowned-results",
		Title:       "Some results have no owners",
		Description: "For some results, no ownership data was found, or no rule applied to the result. [Learn more about configuring code ownership](https://docs.sourcegraph.com/own).",
		// Explicitly set a low priority, so other alerts take precedence.
		Priority: 0,
	}
}

// AlertForOwnershipSearchError returns an alert related to ownership search
// error. This alert has higher priority than `AlertForUnownedResult`.
func AlertForOwnershipSearchError() *Alert {
	return &Alert{
		Kind:        "ownership-search-error",
		Title:       "Error during ownership search",
		Description: "Ownership search returned an error.",
		Priority:    1,
	}
}
