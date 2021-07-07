package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func (r *Resolver) UserCodeGraph(ctx context.Context, user *graphqlbackend.UserResolver) (graphqlbackend.CodeGraphPersonNodeResolver, error) {
	return &CodeGraphPersonNodeResolver{
		user:     user,
		resolver: r,
	}, nil
}

type CodeGraphPersonNodeResolver struct {
	user *graphqlbackend.UserResolver

	resolver *Resolver
}

// TODO(sqs): un-hardcode
var repoNames = []api.RepoName{
	"github.com/sourcegraph/go-jsonschema",
	"github.com/sourcegraph/go-diff",
	"github.com/sourcegraph/jsonx",
	"github.com/sourcegraph/docsite",
	// "github.com/sourcegraph/sourcegraph",
}

const myEmail = "quinn@slack.org"
