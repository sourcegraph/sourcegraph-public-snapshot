package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
	"github.com/pkg/errors"
)

func (c *Client) requestGraphQL(ctx context.Context, query string, vars map[string]interface{}, result interface{}) (err error) {
	reqBody, err := json.Marshal(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return err
	}

	// GitHub.com GraphQL endpoint is api.github.com/graphql. GitHub Enterprise is /api/graphql (the
	// REST endpoint is /api/v3, necessitating the "..").
	graphqlEndpoint := "/graphql"
	if !c.githubDotCom {
		graphqlEndpoint = "../graphql"
	}
	req, err := http.NewRequest("POST", graphqlEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	// Enable Checks API
	// https://developer.github.com/v4/previews/#checks
	req.Header.Add("Accept", "application/vnd.github.antiope-preview+json")
	var respBody struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}

	cost, err := estimateGraphQLCost(query)
	if err != nil {
		return errors.Wrap(err, "estimating graphql cost")
	}

	if err := c.rateLimit.WaitN(ctx, cost); err != nil {
		return errors.Wrap(err, "rate limit")
	}

	if err := c.do(ctx, req, &respBody); err != nil {
		return err
	}

	// If the GraphQL response has errors, still attempt to unmarshal the data portion, as some
	// requests may expect errors but have useful responses (e.g., querying a list of repositories,
	// some of which you expect to 404).
	if len(respBody.Errors) > 0 {
		err = respBody.Errors
	}
	if result != nil && respBody.Data != nil {
		if err0 := unmarshal(respBody.Data, result); err0 != nil && err == nil {
			return err0
		}
	}
	return err
}

// estimateGraphQLCost estimates the cost of the query as described here:
// https://developer.github.com/v4/guides/resource-limitations/#calculating-a-rate-limit-score-before-running-the-call
func estimateGraphQLCost(query string) (int, error) {
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
	totalCost = totalCost / 100
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

// graphqlErrors describes the errors in a GraphQL response. It contains at least 1 element when returned by
// requestGraphQL. See https://graphql.github.io/graphql-spec/June2018/#sec-Errors.
type graphqlErrors []struct {
	Message   string        `json:"message"`
	Type      string        `json:"type"`
	Path      []interface{} `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

const graphqlErrTypeNotFound = "NOT_FOUND"

func (e graphqlErrors) Error() string {
	return fmt.Sprintf("error in GraphQL response: %s", e[0].Message)
}
