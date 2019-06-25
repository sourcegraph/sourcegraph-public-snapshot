package git

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func (GraphQLResolver) CreateRefFromPatch(ctx context.Context, arg *struct {
	Input graphqlbackend.GitCreateRefFromPatchInput
}) (graphqlbackend.GitCreateRefFromPatchPayload, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	commitID, err := gitserver.DefaultClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name()),
		BaseCommit: api.CommitID(arg.Input.BaseCommit),
		TargetRef:  arg.Input.Name,
		Patch:      arg.Input.Patch,
		CommitInfo: protocol.PatchCommitInfo{
			AuthorName:  "Sourcegraph",
			AuthorEmail: "bot@sourcegraph.com",
			Message:     "bot commit",
			Date:        time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}

	return &gitCreateRefFromPatchPayload{ref: graphqlbackend.NewGitRefResolver(repo, arg.Input.Name, commitID)}, nil
}

type gitCreateRefFromPatchPayload struct {
	ref *graphqlbackend.GitRefResolver
}

func (p gitCreateRefFromPatchPayload) Ref(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {

}
