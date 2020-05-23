package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/search/query"

// searchQueryDescription is a type for the SearchQueryDescription resolver used
// by SearchAlert.
type searchQueryDescription struct {
	description string
	query       string
	patternType query.SearchType
}

func (q searchQueryDescription) Query() string {
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

func (q searchQueryDescription) Description() *string {
	if q.description == "" {
		return nil
	}

	return &q.description
}
