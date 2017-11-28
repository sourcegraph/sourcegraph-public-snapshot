package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"

	log15 "gopkg.in/inconshreveable/log15.v2"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/neelance/parallel"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type repositoryResolver struct {
	repo        *sourcegraph.Repo
	redirectURL *string
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

func repositoryByIDInt32(ctx context.Context, id int32) (*repositoryResolver, error) {
	repo, err := localstore.Repos.Get(ctx, id)
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

func (r *repositoryResolver) RedirectURL() *string {
	return r.redirectURL
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
	rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
		Rev:  args.Rev,
	})
	if err != nil {
		if err, ok := err.(vcs.RepoNotExistError); ok && err.CloneInProgress {
			return &commitStateResolver{cloneInProgress: true}, nil
		}
		return nil, err
	}

	return &commitStateResolver{
		commit: &commitResolver{
			commit: commitSpec{RepoID: r.repo.ID, CommitID: rev.CommitID},
			repo:   *r.repo,
		},
	}, nil
}

// GitCmdRaw executes whitelisted git cmds from the gitserver.
func (r *repositoryResolver) GitCmdRaw(ctx context.Context, args *struct {
	Params []string
}) (string, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.repo.ID)
	if err != nil {
		return "", err
	}

	return vcsrepo.GitCmdRaw(ctx, args.Params)
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

func (r *repositoryResolver) LastIndexedRevOrLatest(ctx context.Context) (*commitStateResolver, error) {
	// This method is a stopgap until we no longer require git:// URIs on the client which include rev data.
	// THIS RESOLVER WILL BE REMOVED SOON, DO NOT USE IT!!!
	if r.repo.IndexedRevision != nil && *r.repo.IndexedRevision != "" {
		return createCommitState(*r.repo, &sourcegraph.ResolvedRev{CommitID: *r.repo.IndexedRevision}), nil
	}
	return r.Latest(ctx)
}

func (r *repositoryResolver) DefaultBranch(ctx context.Context) (*string, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.repo.ID)
	if err != nil {
		return nil, err
	}
	defaultBranch, err := vcsrepo.GitCmdRaw(ctx, []string{"rev-parse", "--abbrev-ref", "HEAD"})
	// If we fail to get the default branch due to cloning, we return nothing.
	if err != nil {
		if err, ok := err.(vcs.RepoNotExistError); ok && err.CloneInProgress {
			return nil, nil
		}
		return nil, err
	}
	t := strings.TrimSpace(defaultBranch)
	return &t, nil
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

func (r *repositoryResolver) StarsCount() *int32 {
	return uintPtrToInt32Ptr(r.repo.StarsCount)
}

func (r *repositoryResolver) ForksCount() *int32 {
	return uintPtrToInt32Ptr(r.repo.ForksCount)
}

func uintPtrToInt32Ptr(v *uint) *int32 {
	if v == nil || *v > math.MaxInt32 {
		return nil
	}
	v32 := int32(*v)
	return &v32
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

func (r *repositoryResolver) ListTotalRefs(ctx context.Context) (*totalRefListResolver, error) {
	totalRefs, err := backend.Defs.ListTotalRefs(ctx, r.repo.URI)
	if err != nil {
		return nil, err
	}
	originalLength := len(totalRefs)

	// Limit total references to 250 to prevent the many localstore.Repos.Get
	// operations from taking too long.
	sort.Sort(sortByRepoSpecID(totalRefs))
	if limit := 250; len(totalRefs) > limit {
		totalRefs = totalRefs[:limit]
	}

	// Transform repo IDs into repository resolvers.
	var (
		par         = 8
		resolversMu sync.Mutex
		resolvers   = make([]*repositoryResolver, 0, len(totalRefs))
		run         = parallel.NewRun(par)
	)
	for _, repoSpec := range totalRefs {
		run.Acquire()
		go func(repoSpec sourcegraph.RepoSpec) {
			defer func() {
				if r := recover(); r != nil {
					run.Error(fmt.Errorf("recover: %v", r))
				}
				run.Release()
			}()
			resolver, err := repositoryByIDInt32(ctx, repoSpec.ID)
			if err != nil {
				run.Error(err)
				return
			}
			resolversMu.Lock()
			resolvers = append(resolvers, resolver)
			resolversMu.Unlock()
		}(repoSpec)
	}
	if err := run.Wait(); err != nil {
		// Log the error if we still have good results; otherwise return just
		// the error.
		if len(resolvers) > 5 {
			log.Println("ListTotalRefs:", r.repo.URI, err)
		} else {
			return nil, err
		}
	}
	return &totalRefListResolver{
		repositories: resolvers,
		total:        int32(originalLength),
	}, nil
}

type totalRefListResolver struct {
	repositories []*repositoryResolver
	total        int32
}

func (t *totalRefListResolver) Repositories() []*repositoryResolver {
	return t.repositories
}

func (t *totalRefListResolver) Total() int32 {
	return t.total
}

type sortByRepoSpecID []sourcegraph.RepoSpec

func (s sortByRepoSpecID) Len() int      { return len(s) }
func (s sortByRepoSpecID) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByRepoSpecID) Less(i, j int) bool {
	return s[i].ID < s[j].ID
}

func (*schemaResolver) AddPhabricatorRepo(ctx context.Context, args *struct {
	Callsign string
	URI      string
}) (*EmptyResponse, error) {
	if !envvar.DeploymentOnPrem() {
		return nil, errors.New("AddPhabricatorRepo: illegal operation on public Sourcegraph server")
	}
	_, err := localstore.Phabricator.CreateIfNotExists(ctx, args.Callsign, args.URI)
	if err != nil {
		log15.Error("adding phabricator repo", "callsign", args.Callsign, "uri", args.URI)
	}
	return nil, err
}
