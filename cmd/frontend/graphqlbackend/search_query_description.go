package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// searchQueryDescriptionResolver is a type for the SearchQueryDescription resolver used
// by SearchAlert.
type searchQueryDescriptionResolver struct {
	pq alert.ProposedQuery
}

func (q searchQueryDescriptionResolver) Query() string {
	if q.pq.Description != "Remove quotes" {
		switch q.pq.PatternType {
		case query.SearchTypeRegex:
			return q.pq.Query + " patternType:regexp"
		case query.SearchTypeLiteral:
			return q.pq.Query + " patternType:literal"
		case query.SearchTypeStructural:
			return q.pq.Query + " patternType:structural"
		default:
			panic("unreachable")
		}
	}
	return q.pq.Query
}

func (q searchQueryDescriptionResolver) Description() *string {
	if q.pq.Description == "" {
		return nil
	}

	return &q.pq.Description
}
