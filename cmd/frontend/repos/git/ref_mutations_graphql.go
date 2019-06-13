package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

	if !strings.HasPrefix(repo.Name(), "github.com/sd9/") && !strings.HasPrefix(repo.Name(), "github.com/sd9org/") {
		return nil, errors.New("refusing to modify non-sd9 test repo") // TODO!(sqs)
	}

	commitID, err := gitserver.DefaultClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name()),
		BaseCommit: api.CommitID(arg.Input.BaseCommit),
		TargetRef:  arg.Input.Name,
		Patch:      arg.Input.Patch,
		CommitInfo: protocol.PatchCommitInfo{
			AuthorName:  "Quinn Slack",         // TODO!(sqs): un-hardcode
			AuthorEmail: "sqs@sourcegraph.com", // TODO!(sqs): un-hardcode
			Message:     arg.Input.CommitMessage,
			Date:        time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}

	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return nil, err
	}

	// Push up the update to GitHub. TODO!(sqs) this only makes sense for the demo
	cmd := gitserver.DefaultClient.Command("git", "push", "-f", "--", "origin", fmt.Sprintf("refs/heads/%s:refs/heads/%s", defaultBranch.AbbrevName(), defaultBranch.AbbrevName()), arg.Input.Name+":"+arg.Input.Name)
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repo.Name())}
	if out, err := cmd.CombinedOutput(ctx); err != nil {
		return nil, fmt.Errorf("%s\n\n%s", err, out)
	}

	return &gitCreateRefFromPatchPayload{ref: graphqlbackend.NewGitRefResolver(repo, arg.Input.Name, graphqlbackend.GitObjectID(commitID))}, nil
}

type gitCreateRefFromPatchPayload struct {
	ref *graphqlbackend.GitRefResolver
}

func (p *gitCreateRefFromPatchPayload) Ref() *graphqlbackend.GitRefResolver { return p.ref }
