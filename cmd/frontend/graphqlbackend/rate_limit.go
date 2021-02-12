package graphqlbackend

import (
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
	"github.com/pkg/errors"
)

// Included in tracing so that we can differentiate different costs as we tweak
// the algorithm
const costEstimateVersion = 1

// estimateQueryCost estimates the cost of the query based on the method used by GitHub as described here:
// https://developer.github.com/v4/guides/resource-limitations/#calculating-a-rate-limit-score-before-running-the-call
func estimateQueryCost(query string) (int, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: query,
	})
	if err != nil {
		return 0, errors.Wrap(err, "parsing query")
	}

	var totalCost int
	for _, def := range doc.Definitions {
		cost := calcDefinitionCost(def)
		totalCost += cost
	}

	// As per the calculation spec, cost should be divided by 100
	totalCost /= 100
	if totalCost < 1 {
		return 1, nil
	}
	return totalCost, nil
}

type limitDepth struct {
	// The 'first' or 'last' limit
	limit int
	// The depth at which it was added
	depth int
}

func calcDefinitionCost(def ast.Node) int {
	var cost int
	limitStack := make([]limitDepth, 0)

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch node := p.Node.(type) {
			case *ast.IntValue:
				// We're looking for a 'first' or 'last' param indicating a limit
				parent, ok := p.Parent.(*ast.Argument)
				if !ok {
					return visitor.ActionNoChange, nil
				}
				if parent.Name == nil {
					return visitor.ActionNoChange, nil
				}
				if parent.Name.Value != "first" && parent.Name.Value != "last" {
					return visitor.ActionNoChange, nil
				}

				// Prune anything above our current depth as we may have started walking
				// back down the tree
				currentDepth := len(p.Ancestors)
				limitStack = filterInPlace(limitStack, currentDepth)

				limit, err := strconv.Atoi(node.Value)
				if err != nil {
					return "", errors.Wrap(err, "parsing limit")
				}
				limitStack = append(limitStack, limitDepth{limit: limit, depth: currentDepth})
				// The first item in the tree is always worth 1
				if len(limitStack) == 1 {
					cost++
					return visitor.ActionNoChange, nil
				}
				// The cost of the current item is calculated using the limits of
				// its children
				children := limitStack[:len(limitStack)-1]
				product := 1
				// Multiply them all together
				for _, n := range children {
					product = n.limit * product
				}
				cost += product
			}
			return visitor.ActionNoChange, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	return cost
}

func filterInPlace(limitStack []limitDepth, depth int) []limitDepth {
	n := 0
	for _, x := range limitStack {
		if depth > x.depth {
			limitStack[n] = x
			n++
		}
	}
	limitStack = limitStack[:n]
	return limitStack
}
