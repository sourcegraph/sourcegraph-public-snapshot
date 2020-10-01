package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/machinebox/graphql"
)

const costPerSearch = 30
const maxCostPerRequest = 1000

// IssueLoader efficiently fetches issues and pull request that match a given set
// of queries.
type IssueLoader struct {
	queries   []string
	fragments []string
	args      [][]string
	cursors   []string
	done      []bool
}

// NewIssueLoader creates a new IssueLoader with the given queries.
func NewIssueLoader(queries []string) *IssueLoader {
	fragments, args := makeFragmentArgs(len(queries))

	return &IssueLoader{
		queries:   queries,
		fragments: fragments,
		args:      args,
		cursors:   make([]string, len(queries)),
		done:      make([]bool, len(queries)),
	}
}

// Load will load all issues and pull requests matching the configured queries.
// Returned objects are deduplicated and tracking issues are filtered out of the
// issues list.
func (l *IssueLoader) Load(ctx context.Context, cli *graphql.Client) (issues []*Issue, pullRequests []*PullRequest, _ error) {
	for {
		r, ok := l.makeNextRequest()
		if !ok {
			break
		}

		pageIssues, pagePullRequests, err := l.performRequest(ctx, cli, r)
		if err != nil {
			return nil, nil, err
		}

		issues = append(issues, pageIssues...)
		pullRequests = append(pullRequests, pagePullRequests...)
	}

	return deduplicateIssues(issues), deduplicatePullRequests(pullRequests), nil
}

// makeNextRequest will construct a new request based on the given cursor values.
// If no request should be performed, this method will return a false-valued flag.
func (l *IssueLoader) makeNextRequest() (*graphql.Request, bool) {
	var args []string
	var fragments []string
	vars := map[string]interface{}{}

	cost := 0
	for i := range l.queries {
		cost += costPerSearch

		if l.done[i] || cost > maxCostPerRequest {
			continue
		}

		args = append(args, l.args[i]...)
		fragments = append(fragments, l.fragments[i])
		vars[fmt.Sprintf("query%d", i)] = l.queries[i]
		vars[fmt.Sprintf("count%d", i)] = costPerSearch
		if l.cursors[i] != "" {
			vars[fmt.Sprintf("cursor%d", i)] = l.cursors[i]
		}
	}

	if len(fragments) == 0 {
		return nil, false
	}

	r := graphql.NewRequest(fmt.Sprintf(`query(%s) { %s }`, strings.Join(args, ", "), strings.Join(fragments, "\n")))
	for k, v := range vars {
		r.Var(k, v)
	}

	return r, true
}

// performRequest will perform the given request and return the deserialized
// list of issues and pull requests.
func (l *IssueLoader) performRequest(ctx context.Context, cli *graphql.Client, r *graphql.Request) (issues []*Issue, pullRequests []*PullRequest, _ error) {
	var payload map[string]SearchResult
	if err := cli.Run(ctx, r, &payload); err != nil {
		return nil, nil, err
	}

	for name, result := range payload {
		// Note: the search fragment aliases have the form `search123`
		index, err := strconv.Atoi(name[6:])
		if err != nil {
			return nil, nil, err
		}

		searchIssues, searchPullRequests := unmarshalSearchNodes(result.Nodes)
		issues = append(issues, searchIssues...)
		pullRequests = append(pullRequests, searchPullRequests...)

		if len(result.Nodes) > 0 && result.PageInfo.HasNextPage {
			l.cursors[index] = result.PageInfo.EndCursor
		} else {
			l.done[index] = true
		}
	}

	return issues, pullRequests, nil
}

// makeFragmentArgs makes `n` named GraphQL fragment and an associated set of variables.
// This is used to later construct a GraphQL request with a subset of these queries.
func makeFragmentArgs(n int) (fragments []string, args [][]string) {
	for i := 0; i < n; i++ {
		fragments = append(fragments, makeSearchQuery(fmt.Sprintf("%d", i)))

		args = append(args, []string{
			fmt.Sprintf("$query%d: String!", i),
			fmt.Sprintf("$count%d: Int!", i),
			fmt.Sprintf("$cursor%d: String", i),
		})
	}

	return fragments, args
}
