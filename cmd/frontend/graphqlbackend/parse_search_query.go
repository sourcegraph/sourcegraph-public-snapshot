package graphqlbackend

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func toJSON(node query.Node) any {
	switch n := node.(type) {
	case query.Operator:
		var jsons []any
		for _, o := range n.Operands {
			jsons = append(jsons, toJSON(o))
		}

		switch n.Kind {
		case query.And:
			return struct {
				And []any `json:"and"`
			}{
				And: jsons,
			}
		case query.Or:
			return struct {
				Or []any `json:"or"`
			}{
				Or: jsons,
			}
		case query.Concat:
			// Concat should already be processed at this point, or
			// the original query expresses something that is not
			// supported. We just return the parse tree anyway.
			return struct {
				Concat []any `json:"concat"`
			}{
				Concat: jsons,
			}
		}
	case query.Parameter:
		return struct {
			Field   string      `json:"field"`
			Value   string      `json:"value"`
			Negated bool        `json:"negated"`
			Labels  []string    `json:"labels"`
			Range   query.Range `json:"range"`
		}{
			Field:   n.Field,
			Value:   n.Value,
			Negated: n.Negated,
			Labels:  n.Annotation.Labels.String(),
			Range:   n.Annotation.Range,
		}
	case query.Pattern:
		return struct {
			Value   string      `json:"value"`
			Negated bool        `json:"negated"`
			Labels  []string    `json:"labels"`
			Range   query.Range `json:"range"`
		}{
			Value:   n.Value,
			Negated: n.Negated,
			Labels:  n.Annotation.Labels.String(),
			Range:   n.Annotation.Range,
		}
	}
	// unreachable.
	return struct{}{}
}

func (r *schemaResolver) ParseSearchQuery(ctx context.Context, args *struct {
	Query       string
	PatternType string
}) (*JSONValue, error) {
	var searchType query.SearchType
	switch args.PatternType {
	case "literal":
		searchType = query.SearchTypeLiteralDefault
	case "structural":
		searchType = query.SearchTypeStructural
	case "regexp", "regex":
		searchType = query.SearchTypeRegex
	default:
		searchType = query.SearchTypeLiteralDefault
	}

	plan, err := query.Pipeline(query.Init(args.Query, searchType))
	if err != nil {
		return nil, err
	}

	var jsons []any
	for _, node := range plan.ToQ() {
		jsons = append(jsons, toJSON(node))
	}
	json, err := json.Marshal(jsons)
	if err != nil {
		return nil, err
	}
	return &JSONValue{Value: string(json)}, nil
}
