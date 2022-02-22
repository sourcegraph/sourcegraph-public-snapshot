package query

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func Test(t *testing.T) {
	var plan query.Plan
	q := "repo:sourcegraph/.* -repo:asdf/.*"
	plan, err := query.Pipeline(
		query.Init(q, query.SearchTypeRegex),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(plan.ToParseTree().Repositories())
}
