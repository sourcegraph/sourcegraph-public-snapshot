package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func (r *schemaResolver) ParseSearchQuery(ctx context.Context, args *struct {
	Query       string
	PatternType string
}) (*JSONValue, error) {
	var searchType query.SearchType
	switch args.PatternType {
	case "literal":
		searchType = query.SearchTypeLiteral
	case "structural":
		searchType = query.SearchTypeStructural
	case "regexp", "regex":
		searchType = query.SearchTypeRegex
	default:
		searchType = query.SearchTypeLiteral
	}

	plan, err := query.Pipeline(query.Init(args.Query, searchType))
	if err != nil {
		return nil, err
	}

	jsonString, err := query.ToJSON(plan.ToQ())
	if err != nil {
		return nil, err
	}
	return &JSONValue{Value: jsonString}, nil
}
