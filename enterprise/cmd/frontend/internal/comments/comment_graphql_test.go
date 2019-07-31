package comments

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type mockComment struct {
	graphqlbackend.Comment
	body string
}

func (v mockComment) Body(context.Context) (string, error) { return v.body, nil }
