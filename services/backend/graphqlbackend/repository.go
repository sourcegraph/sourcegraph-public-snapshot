package graphqlbackend

import (
	"context"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

type repositoryResolver struct {
	nodeBase
	repo *sourcegraph.Repo
}

func repositoryByID(ctx context.Context, id graphql.ID) (nodeResolver, error) {
	var repoID int32
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := localstore.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo.URI); err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func (r *repositoryResolver) ToRepository() (*repositoryResolver, bool) {
	return r, true
}

func (r *repositoryResolver) ID() graphql.ID {
	return relay.MarshalID("Repository", r.repo.ID)
}

func (r *repositoryResolver) URI() string {
	return r.repo.URI
}

func (r *repositoryResolver) Description() string {
	return r.repo.Description
}

func (r *repositoryResolver) Commit(ctx context.Context, args *struct{ Rev string }) (*commitStateResolver, error) {
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
		Rev:  args.Rev,
	})
	if err != nil {
		if err == vcs.ErrRevisionNotFound {
			return &commitStateResolver{}, nil
		}
		if err, ok := err.(vcs.RepoNotExistError); ok && err.CloneInProgress {
			return &commitStateResolver{cloneInProgress: true}, nil
		}
		return nil, err
	}
	return &commitStateResolver{commit: &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}}, nil
}

func (r *repositoryResolver) Latest(ctx context.Context) (*commitStateResolver, error) {
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
	})
	if err != nil {
		if err, ok := err.(vcs.RepoNotExistError); ok && err.CloneInProgress {
			return &commitStateResolver{cloneInProgress: true}, nil
		}
		return nil, err
	}
	return &commitStateResolver{commit: &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}}, nil
}

func (r *repositoryResolver) DefaultBranch() string {
	return r.repo.DefaultBranch
}

func (r *repositoryResolver) Branches(ctx context.Context) ([]string, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.repo.ID)
	if err != nil {
		return nil, err
	}

	branches, err := vcsrepo.Branches(ctx, vcs.BranchesOptions{})
	if err != nil {
		return nil, err
	}

	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	return names, nil
}

func (r *repositoryResolver) Tags(ctx context.Context) ([]string, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.repo.ID)
	if err != nil {
		return nil, err
	}

	tags, err := vcsrepo.Tags(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names, nil
}

func (r *repositoryResolver) Private() bool {
	return r.repo.Private
}

func (r *repositoryResolver) Language() string {
	return r.repo.Language
}

func (r *repositoryResolver) Name() string {
	return r.repo.Name
}

func (r *repositoryResolver) Fork() bool {
	return r.repo.Fork
}

func (r *repositoryResolver) Owner() string {
	return r.repo.Owner
}

func (r *repositoryResolver) PushedAt() string {
	if r.repo.PushedAt != nil {
		return r.repo.PushedAt.String()
	}
	return ""
}

func (r *repositoryResolver) CreatedAt() string {
	if r.repo.CreatedAt != nil {
		return r.repo.CreatedAt.String()
	}
	return ""
}
