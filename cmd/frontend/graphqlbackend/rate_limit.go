package graphqlbackend

import (
	"fmt"
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
	"github.com/pkg/errors"
)

// Included in tracing so that we can differentiate different costs as we tweak
// the algorithm
const costEstimateVersion = 2

type queryCost struct {
	FieldCount int
	MaxDepth   int
}

// estimateQueryCost estimates the cost of the query .
// TODO: Update description
func estimateQueryCost(query string, variables map[string]interface{}) (*queryCost, error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	doc, err := parser.Parse(parser.ParseParams{
		Source: query,
	})
	if err != nil {
		return nil, errors.Wrap(err, "parsing query")
	}

	// We need to separate operations from fragments
	var operations []ast.Node
	var fragments []ast.Node

	for i, def := range doc.Definitions {
		switch def.GetKind() {
		case kinds.FragmentDefinition:
			fragments = append(fragments, doc.Definitions[i])
		case kinds.OperationDefinition:
			operations = append(operations, doc.Definitions[i])
		}
	}

	var totalCost queryCost
	for _, def := range operations {
		cost, err := calcOperationCost(def, variables)
		if err != nil {
			return nil, err
		}
		totalCost.FieldCount += cost.FieldCount
		if totalCost.MaxDepth < cost.MaxDepth {
			totalCost.MaxDepth = cost.MaxDepth
		}
	}

	if totalCost.FieldCount < 1 {
		totalCost.FieldCount = 1
	}
	if totalCost.MaxDepth < 1 {
		totalCost.MaxDepth = 1
	}

	return &totalCost, nil
}

func calcOperationCost(def ast.Node, variables map[string]interface{}) (*queryCost, error) {
	var err error
	fieldCount := 0
	multiplier := 1
	depth := 0

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch node := p.Node.(type) {
			case *ast.SelectionSet:
				depth++
			case *ast.Field:
				// Ignore the "nodes" field as it does not appear in the result
				if node.Name.Value == "nodes" {
					return visitor.ActionNoChange, nil
				}
				fieldCount += multiplier
			case *ast.Variable:
				// We may have a limit variable
				if !shouldCheckParam(p) {
					return visitor.ActionNoChange, nil
				}
				limitVar, ok := variables[node.Name.Value]
				if !ok {
					err = fmt.Errorf("missing variable: %q", node.Name.Value)
					return visitor.ActionBreak, nil
				}
				limit, err := extractInt(limitVar)
				if err != nil {
					err = errors.Wrap(err, "extracting limit")
					return visitor.ActionBreak, nil
				}
				if limit <= 0 {
					return visitor.ActionNoChange, nil
				}
				multiplier *= limit
			case *ast.IntValue:
				// We may have a limit
				if !shouldCheckParam(p) {
					return visitor.ActionNoChange, nil
				}
				limit, err := strconv.Atoi(node.Value)
				if err != nil {
					err = errors.Wrap(err, "parsing limit")
					return visitor.ActionBreak, nil
				}
				if limit <= 0 {
					return visitor.ActionNoChange, nil
				}
				multiplier *= limit
			}
			return visitor.ActionNoChange, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	return &queryCost{
		FieldCount: fieldCount,
		MaxDepth:   depth,
	}, err
}

var quantityParams = map[string]struct{}{
	"first": {},
	"last":  {},
}

func extractInt(i interface{}) (int, error) {
	switch v := i.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unkown limit type: %T", i)
	}
}

func shouldCheckParam(p visitor.VisitFuncParams) bool {
	parent, ok := p.Parent.(*ast.Argument)
	if !ok {
		return false
	}
	if parent.Name == nil {
		return false
	}
	if _, ok := quantityParams[parent.Name.Value]; !ok {
		return false
	}
	return true
}

func getLimit(x interface{}) (int, bool) {
	switch v := x.(type) {
	case int:
		return v, true
	default:
		return 0, false
	}
}
