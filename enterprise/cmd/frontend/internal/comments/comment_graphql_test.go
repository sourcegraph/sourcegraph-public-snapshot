package comments

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type mockComment struct {
	graphqlbackend.Comment
	body string
}

func (v mockComment) Body() string { return v.body }
