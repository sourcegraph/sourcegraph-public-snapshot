package main

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/machinebox/graphql"
)

type searchNode struct {
	Typename   string `json:"__typename"`
	ID         string
	Title      string
	Body       string
	State      string
	Number     int
	URL        string
	Repository struct {
		NameWithOwner string
		IsPrivate     bool
	}
	Author    struct{ Login string }
	Assignees struct{ Nodes []struct{ Login string } }
	Labels    struct{ Nodes []struct{ Name string } }
	Milestone struct{ Title string }
	Commits   struct {
		Nodes []struct {
			Commit struct{ AuthoredDate time.Time }
		}
	}
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  time.Time
}

type search struct {
	PageInfo struct {
		EndCursor   string
		HasNextPage bool
	}
	Nodes []searchNode
}

func loadTrackingIssues(ctx context.Context, cli *graphql.Client, org string, issues []*TrackingIssue) (err error) {
	ch := make(chan *TrackingIssue, len(issues))
	for _, issue := range issues {
		ch <- issue
	}
	close(ch)

	var wg sync.WaitGroup
	errs := make(chan error, len(issues))

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for issue := range ch {
				if err := loadTrackingIssue(ctx, cli, org, issue); err != nil {
					errs <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		if err == nil {
			err = e
		} else {
			err = multierror.Append(err, e)
		}
	}

	return err
}

func loadTrackingIssue(ctx context.Context, cli *graphql.Client, org string, issue *TrackingIssue) error {
	var q bytes.Buffer
	q.WriteString("query(\n")

	type query struct {
		cursor string
		query  string
	}

	queries := map[string]*query{}
	if issue.Milestone == "" {
		name := "tracking" + strconv.Itoa(issue.Number)
		fmt.Fprintf(&q, "$%[1]sCount: Int!, $%[1]sCursor: String, $%[1]sQuery: String!\n", name)
		queries[name] = &query{query: listIssuesSearchQuery(org, "", issue.Labels, false)}
	} else {
		milestoned := "tracking" + strconv.Itoa(issue.Number) + "Milestoned"
		fmt.Fprintf(&q, "$%[1]sCount: Int!, $%[1]sCursor: String, $%[1]sQuery: String!,\n", milestoned)
		queries[milestoned] = &query{query: listIssuesSearchQuery(org, issue.Milestone, issue.Labels, false)}

		demilestoned := "tracking" + strconv.Itoa(issue.Number) + "Demilestoned"
		fmt.Fprintf(&q, "$%[1]sCount: Int!, $%[1]sCursor: String, $%[1]sQuery: String!\n", demilestoned)
		queries[demilestoned] = &query{query: listIssuesSearchQuery(org, issue.Milestone, issue.Labels, true)}
	}

	q.WriteString(") {")

	for query := range queries {
		q.WriteString(searchGraphQLQuery(query))
	}

	q.WriteString("}")

	for {
		r := graphql.NewRequest(q.String())

		for query, args := range queries {
			r.Var(query+"Count", 100)
			r.Var(query+"Query", args.query)
			if args.cursor != "" {
				r.Var(query+"Cursor", args.cursor)
			}
		}

		var data map[string]search

		err := cli.Run(ctx, r, &data)
		if err != nil {
			return err
		}

		var hasNextPage bool
		for query, s := range data {
			q := queries[query]

			if s.PageInfo.HasNextPage && len(s.Nodes) > 0 {
				hasNextPage = true
				q.cursor = s.PageInfo.EndCursor
			}

			issues, prs := unmarshalSearchNodes(s.Nodes)
			issue.Issues = append(issue.Issues, issues...)
			issue.PRs = append(issue.PRs, prs...)
		}

		if !hasNextPage {
			break
		}
	}

	return nil
}

func listTrackingIssues(ctx context.Context, cli *graphql.Client, issuesQuery string) (all []*TrackingIssue, _ error) {
	var q strings.Builder
	q.WriteString("query($trackingCount: Int!, $trackingCursor: String, $trackingQuery: String!) {\n")
	q.WriteString(searchGraphQLQuery("tracking"))
	q.WriteString("}")

	r := graphql.NewRequest(q.String())

	r.Var("trackingCount", 100)
	r.Var("trackingQuery", issuesQuery)

	for {
		var data struct{ Tracking search }

		err := cli.Run(ctx, r, &data)
		if err != nil {
			return nil, err
		}

		issues, _ := unmarshalSearchNodes(data.Tracking.Nodes)

		for _, issue := range issues {
			if len(issue.Labels) > 1 { // Skip tracking issues that have only the "tracking" label
				all = append(all, NewTrackingIssue(issue))
			}
		}

		if data.Tracking.PageInfo.HasNextPage {
			r.Var("trackingCursor", data.Tracking.PageInfo.EndCursor)
		} else {
			break
		}
	}

	return all, nil
}

func unmarshalSearchNodes(nodes []searchNode) (issues []*Issue, prs []*PullRequest) {
	for _, n := range nodes {
		switch n.Typename {
		case "PullRequest":
			pr := &PullRequest{
				ID:         n.ID,
				Title:      n.Title,
				Body:       n.Body,
				State:      n.State,
				Number:     n.Number,
				URL:        n.URL,
				Repository: n.Repository.NameWithOwner,
				Private:    n.Repository.IsPrivate,
				Assignees:  make([]string, 0, len(n.Assignees.Nodes)),
				Labels:     make([]string, 0, len(n.Labels.Nodes)),
				Milestone:  n.Milestone.Title,
				Author:     n.Author.Login,
				CreatedAt:  n.CreatedAt,
				UpdatedAt:  n.UpdatedAt,
				ClosedAt:   n.ClosedAt,
				BeganAt:    n.Commits.Nodes[0].Commit.AuthoredDate,
			}

			for _, assignee := range n.Assignees.Nodes {
				pr.Assignees = append(pr.Assignees, assignee.Login)
			}

			for _, label := range n.Labels.Nodes {
				pr.Labels = append(pr.Labels, label.Name)
			}

			prs = append(prs, pr)

		case "Issue":
			issue := &Issue{
				ID:         n.ID,
				Title:      n.Title,
				Body:       n.Body,
				State:      n.State,
				Number:     n.Number,
				URL:        n.URL,
				Repository: n.Repository.NameWithOwner,
				Private:    n.Repository.IsPrivate,
				Assignees:  make([]string, 0, len(n.Assignees.Nodes)),
				Labels:     make([]string, 0, len(n.Labels.Nodes)),
				Milestone:  n.Milestone.Title,
				Author:     n.Author.Login,
				CreatedAt:  n.CreatedAt,
				UpdatedAt:  n.UpdatedAt,
				ClosedAt:   n.ClosedAt,
			}

			for _, assignee := range n.Assignees.Nodes {
				issue.Assignees = append(issue.Assignees, assignee.Login)
			}

			for _, label := range n.Labels.Nodes {
				issue.Labels = append(issue.Labels, label.Name)
			}

			issues = append(issues, issue)
		}
	}

	return issues, prs
}

func searchGraphQLQuery(alias string) string {
	const searchQuery = `%[1]s: search(first: $%[1]sCount, type: ISSUE, after: $%[1]sCursor query: $%[1]sQuery) {
		pageInfo {
			endCursor
			hasNextPage
		}
		nodes {
			... on Issue {
				%s
			}
			... on PullRequest {
				%s
			}
		}
	}`

	return fmt.Sprintf(searchQuery,
		alias,
		searchNodeFields(false),
		searchNodeFields(true),
	)
}

func searchNodeFields(isPR bool) string {
	fields := `
		__typename
		id, title, body, state, number, url
		createdAt, closedAt
		repository { nameWithOwner, isPrivate }
		author { login }
		assignees(first: 25) { nodes { login } }
		labels(first: 25) { nodes { name } }
		milestone { title }
	`

	if isPR {
		fields += `
			commits(first: 1) { nodes { commit { authoredDate } } }
		`
	}

	return fields
}

func listIssuesSearchQuery(org, milestone string, labels []string, demilestoned bool) string {
	var q strings.Builder

	fmt.Fprintf(&q, "org:%q", org)

	if milestone != "" {
		if demilestoned {
			fmt.Fprintf(&q, ` -milestone:%q label:"planned/%s"`, milestone, milestone)
		} else {
			fmt.Fprintf(&q, " milestone:%q", milestone)
		}
	}

	for _, label := range labels {
		if label != "" && label != "tracking" {
			fmt.Fprintf(&q, " label:%q", label)
		}
	}

	return q.String()
}
