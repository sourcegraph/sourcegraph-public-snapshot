package main

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/machinebox/graphql"
)

func updateIssues(ctx context.Context, cli *graphql.Client, issues []*Issue) (err error) {
	ch := make(chan *Issue, len(issues))
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
				if err := updateIssue(ctx, cli, issue); err != nil {
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

func updateIssue(ctx context.Context, cli *graphql.Client, issue *Issue) (err error) {
	var q bytes.Buffer
	q.WriteString("mutation(")
	fmt.Fprintf(&q, "$issue%dInput: UpdateIssueInput!", issue.Number)
	q.WriteString(") {")
	fmt.Fprintf(&q, "issue%[1]d: updateIssue(input: $issue%[1]dInput) { issue { updatedAt } }\n", issue.Number)
	q.WriteString("}")

	r := graphql.NewRequest(q.String())

	type UpdateIssueInput struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	}

	r.Var(fmt.Sprintf("issue%dInput", issue.Number), &UpdateIssueInput{
		ID:   issue.ID,
		Body: issue.Body,
	})

	return cli.Run(ctx, r, nil)
}

func findMarker(s, marker string) (int, error) {
	location := strings.Index(s, marker)
	if location == -1 {
		return -1, fmt.Errorf("could not find marker %s in issue body", marker)
	}
	return location, nil
}

func patch(s, replacement string) (string, error) {
	start, err := findMarker(s, beginWorkMarker)
	if err != nil {
		return s, err
	}
	end, err := findMarker(s, endWorkMarker)
	if err != nil {
		return s, err
	}

	return s[:start+len(beginWorkMarker)] + replacement + s[end:], nil
}
