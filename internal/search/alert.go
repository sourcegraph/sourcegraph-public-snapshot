package search

import "github.com/sourcegraph/sourcegraph/internal/search/query"

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []*ProposedQuery
	// The higher the priority the more important is the alert.
	Priority int
}

type ProposedQuery struct {
	description string
	query       string
	patternType query.SearchType
}

func NewProposedQuery(description, query string, patternType query.SearchType) *ProposedQuery {
	return &ProposedQuery{description: description, query: query, patternType: patternType}
}

func (q *ProposedQuery) QueryString() string {
	if q.description != "Remove quotes" {
		switch q.patternType {
		case query.SearchTypeRegex:
			return q.query + " patternType:regexp"
		case query.SearchTypeLiteral:
			return q.query + " patternType:literal"
		case query.SearchTypeStructural:
			return q.query + " patternType:structural"
		default:
			panic("unreachable")
		}
	}
	return q.query
}

func (q *ProposedQuery) Description() *string {
	if q.description == "" {
		return nil
	}

	return &q.description
}
