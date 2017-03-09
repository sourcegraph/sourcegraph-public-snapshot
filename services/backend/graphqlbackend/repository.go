package graphqlbackend

import (
	"context"
	"strings"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/sourcegraph/zap"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

type repositoryResolver struct {
	repo *sourcegraph.Repo
}

func repositoryByID(ctx context.Context, id graphql.ID) (*repositoryResolver, error) {
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
	return createCommitState(*r.repo, rev), nil
}

func (r *repositoryResolver) RevState(ctx context.Context, args *struct{ Rev string }) (*commitStateResolver, error) {
	var zapRef *zapRefResolver

	if conf.AppURL.Host != "sourcegraph.dev.uberinternal.com:30000" && conf.AppURL.Host != "node.aws.sgdev.org:30000" {
		// If the revision is empty or if it ends in ^{git} then we do not need to query zap.
		if args.Rev != "" && !strings.HasSuffix(args.Rev, "^{git}") {
			cl, err := backend.NewZapClient(ctx)
			if err != nil {
				return nil, err
			}
			// TODO(matt,john): remove this hack, this zapRefInfo call was causing a front-end error
			// that looked like Server request for query `Workbench` failed for the following reasons: 1. repository does not exist
			zapRefInfo, _ := cl.RefInfo(ctx, zap.RefIdentifier{Repo: r.repo.URI, Ref: args.Rev})
			// TODO(john): add error-specific handling?
			if zapRefInfo != nil && zapRefInfo.State != nil {
				zapRef = &zapRefResolver{zapRef: zapRefSpec{Base: zapRefInfo.State.GitBase, Branch: zapRefInfo.State.GitBranch}}

				// We want to use the Git revision that the Zap branch was based on,
				// as all of the Zap operations were originating from that revision.
				// Using any other revision of the same branch would be a mistake
				// (e.g., the user may be on a revision of the master branch that is
				// just a few commits behind.)
				return &commitStateResolver{
					zapRef: zapRef,
					commit: &commitResolver{
						commit: commitSpec{
							RepoID:        r.repo.ID,
							CommitID:      zapRefInfo.State.GitBase,
							DefaultBranch: r.repo.DefaultBranch,
						},
					},
				}, nil
			}
		}
	}

	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
		Rev:  args.Rev,
	})
	if err != nil {
		if err == vcs.ErrRevisionNotFound {
			return &commitStateResolver{zapRef: zapRef}, nil
		}
		if err, ok := err.(vcs.RepoNotExistError); ok && err.CloneInProgress {
			return &commitStateResolver{cloneInProgress: true}, nil
		}
		return nil, err
	}

	return &commitStateResolver{zapRef: zapRef,
		commit: &commitResolver{
			commit: commitSpec{RepoID: r.repo.ID, CommitID: rev.CommitID, DefaultBranch: r.repo.DefaultBranch},
			repo:   *r.repo,
		},
	}, nil
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
	return createCommitState(*r.repo, rev), nil
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

func (r *repositoryResolver) Fork() bool {
	return r.repo.Fork
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

// TrialExpiration is the Unix timestamp that the repo trial will expire, or
// nil if this repo is not on a trial.
func (r *repositoryResolver) ExpirationDate(ctx context.Context) (*int32, error) {
	t, err := localstore.Payments.TrialExpirationDate(ctx, *r.repo)
	if err != nil {
		return nil, err
	}

	if t == nil {
		return nil, nil
	}

	n := int32(t.Unix())
	return &n, nil
}
