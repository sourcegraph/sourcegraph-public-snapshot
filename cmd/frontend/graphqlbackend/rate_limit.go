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

// estimateQueryCost estimates the cost of the query before it is actually
// executed. It is a worst cast estimate of the number of fields expected to be
// returned by the query and handles nested queries a well as fragments.
func estimateQueryCost(query string, variables map[string]interface{}) (totalCost *queryCost, err error) {
	// NOTE: When we encounter errors in our visit funcs we return
	// visitor.ActionBreak to stop walking the tree and set the top level err
	// variable so that it is returned

	// TODO: Remove this. It's here as a safeguard until we've run over a large
	// number of real world queries.
	defer func() {
		if r := recover(); r != nil {
			totalCost = nil
			err = r.(error)
		}
	}()
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
	var fragments []*ast.FragmentDefinition

	for i, def := range doc.Definitions {
		switch def.GetKind() {
		case kinds.FragmentDefinition:
			frag, ok := doc.Definitions[i].(*ast.FragmentDefinition)
			if !ok {
				return nil, fmt.Errorf("expected FragmentDefinition, got %T", doc.Definitions[i])
			}
			fragments = append(fragments, frag)
		case kinds.OperationDefinition:
			operations = append(operations, doc.Definitions[i])
		}
	}

	// Costs of fragment definitions
	fragmentCosts := make(map[string]int)
	for _, frag := range fragments {
		name, cost, err := calcFragmentCost(frag)
		if err != nil {
			return nil, errors.Wrap(err, "calculating fragment cost")
		}
		fragmentCosts[name] = cost
	}

	totalCost = &queryCost{}
	for _, def := range operations {
		cost, err := calcOperationCost(def, fragmentCosts, variables)
		if err != nil {
			return nil, errors.Wrap(err, "calculating operation cost")
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

	return totalCost, nil
}

func calcOperationCost(def ast.Node, fragmentCosts map[string]int, variables map[string]interface{}) (*queryCost, error) {
	// NOTE: When we encounter errors in our visit funcs we return
	// visitor.ActionBreak to stop walking the tree and set the top level err
	// variable so that it is returned
	var err error

	if fragmentCosts == nil {
		fragmentCosts = make(map[string]int)
	}
	inlineFragmentCosts := make(map[string]int)
	inlineFragmentDepth := 0

	// limitStack keeps track of the limit as we increase and decrease our depth in
	// the tree and encounter limit values
	limitStack := make([]int, 0)
	currentLimit := 1

	fieldCount := 0
	depth := 0
	maxDepth := 0
	multiplier := 1

	pushLimit := func() {
		multiplier = multiplier * currentLimit
		limitStack = append(limitStack, currentLimit)
		// Set limit back to 1 as we've already used it to increase our multiplier
		currentLimit = 1
	}
	popLimit := func() {
		if len(limitStack) == 0 {
			return
		}
		currentLimit = limitStack[len(limitStack)-1]
		limitStack = limitStack[:len(limitStack)-1]
		multiplier = multiplier / currentLimit
	}

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch node := p.Node.(type) {
			case *ast.SelectionSet:
				depth++
				if depth > maxDepth {
					maxDepth = depth
				}
				pushLimit()
			case *ast.Field:
				if node.Name.Value == "nodes" {
					// Ignore the "nodes" field as it does not appear in the result
					return visitor.ActionNoChange, nil
				}
				if inlineFragmentDepth > 0 {
					// We don't count fields inside of inline fragments as we need to count all fragments
					// first to pick the largest one
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
				currentLimit = limit
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
				currentLimit = limit
			case *ast.FragmentSpread:
				fragmentCost, ok := fragmentCosts[node.Name.Value]
				if !ok {
					err = fmt.Errorf("unknown fragment %q", node.Name.Value)
					return visitor.ActionBreak, nil
				}
				fieldCount += fragmentCost * multiplier
			case *ast.InlineFragment:
				inlineFragmentDepth++
				// We calculate inline fragment costs and store them
				var fragCost *queryCost
				fragCost, err = calcOperationCost(node.SelectionSet, fragmentCosts, variables)
				if err != nil {
					err = errors.Wrap(err, "calculating inline fragment cost")
					return visitor.ActionBreak, nil
				}
				inlineFragmentCosts[node.TypeCondition.Name.Value] = fragCost.FieldCount * multiplier
			}
			return visitor.ActionNoChange, nil
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch p.Node.(type) {
			case *ast.SelectionSet:
				depth--
				popLimit()
			case *ast.InlineFragment:
				inlineFragmentDepth--
			}
			return visitor.ActionNoChange, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	// We also need to pick the largest inline fragment in our tree
	var maxInlineFragmentCost int
	for _, v := range inlineFragmentCosts {
		if v > maxInlineFragmentCost {
			maxInlineFragmentCost = v
		}
	}

	return &queryCost{
		FieldCount: fieldCount + maxInlineFragmentCost,
		MaxDepth:   maxDepth,
	}, err
}

func calcFragmentCost(frag *ast.FragmentDefinition) (string, int, error) {
	var cost int
	var currentFragment string

	fragmentCosts := make(map[string]int)

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch node := p.Node.(type) {
			case *ast.Field:
				cost++
			case *ast.Named:
				currentFragment = node.Name.Value
			case *ast.InlineFragment:
				cost = 0
			}
			return visitor.ActionNoChange, nil
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch p.Node.(type) {
			case *ast.SelectionSet:
				if currentFragment != "" {
					fragmentCosts[currentFragment] = cost
				}
			}
			return visitor.ActionNoChange, nil
		},
	}

	_ = visitor.Visit(frag, v, nil)

	// Find worst case cost
	cost = 0
	for _, v := range fragmentCosts {
		if v > cost {
			cost = v
		}
	}

	return frag.Name.Value, cost, nil
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
