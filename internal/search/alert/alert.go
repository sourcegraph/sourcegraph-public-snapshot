package alert

import (
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []ProposedQuery
	// The higher the priority the more important is the alert.
	Priority int
}

type ProposedQuery struct {
	Description string
	Query       string
	PatternType query.SearchType
}
