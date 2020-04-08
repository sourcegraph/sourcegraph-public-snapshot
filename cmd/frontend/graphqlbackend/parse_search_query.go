package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

/*
type parseSearchQueryResolver struct {
	Value *JSONValue
}

func (r *parseSearchQueryResolver)
*/

func (*schemaResolver) ParseSearchQuery(ctx context.Context, args *struct {
	Query       string
	PatternType string
}) (*JSONValue, error) {
	queryInfo, err := query.ParseAndOr(args.Query)
	if err != nil {
		return nil, err
	}
	json, err := queryInfo.JSON()
	if err != nil {
		return nil, err
	}
	return &JSONValue{Value: string(json)}, nil
}
