package graphqlbackend

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"github.com/neelance/parallel"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/externallink"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type repositoryResolver struct {
	repo        *types.Repo
	redirectURL *string
}

func repositoryByID(ctx context.Context, id graphql.ID) (*repositoryResolver, error) {
	var repoID api.RepoID
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := db.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo); err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func repositoryByIDInt32(ctx context.Context, repoID api.RepoID) (*repositoryResolver, error) {
	repo, err := db.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo); err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func (r *repositoryResolver) ID() graphql.ID {
	return marshalRepositoryID(r.repo.ID)
}

func marshalRepositoryID(repo api.RepoID) graphql.ID { return relay.MarshalID("Repository", repo) }

func unmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

func (r *repositoryResolver) URI() string {
	return string(r.repo.URI)
}

func (r *repositoryResolver) Description() string {
	return r.repo.Description
}

func (r *repositoryResolver) RedirectURL() *string {
	return r.redirectURL
}

func (r *repositoryResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err == backend.ErrMustBeSiteAdmin || err == backend.ErrNotAuthenticated {
			return false, nil // not an error
		}
		return false, err
	}
	return true, nil
}

func (r *repositoryResolver) CloneInProgress(ctx context.Context) (bool, error) {
	return r.MirrorInfo().CloneInProgress(ctx)
}

func (r *repositoryResolver) Commit(ctx context.Context, args *struct{ Rev string }) (*gitCommitResolver, error) {
	commitID, err := backend.Repos.ResolveRev(ctx, r.repo, args.Rev)
	if err != nil {
		if vcs.IsRevisionNotFound(err) {
			return nil, nil
		}
		if vcs.IsCloneInProgress(err) {
			return nil, err
		}
		return nil, err
	}
	commit, err := backend.Repos.GetCommit(ctx, r.repo, commitID)
	if err != nil {
		return nil, err
	}
	resolver := toGitCommitResolver(r, commit)
	resolver.inputRev = &args.Rev
	return resolver, nil
}

func (r *repositoryResolver) LastIndexedRevOrLatest(ctx context.Context) (*gitCommitResolver, error) {
	// This method is a stopgap until we no longer require git:// URIs on the client which include rev data.
	// THIS RESOLVER WILL BE REMOVED SOON, DO NOT USE IT!!!
	if r.repo.IndexedRevision != nil && *r.repo.IndexedRevision != "" {
		return r.Commit(ctx, &struct{ Rev string }{Rev: string(*r.repo.IndexedRevision)})
	}
	return r.Commit(ctx, &struct{ Rev string }{Rev: "HEAD"})
}

func (r *repositoryResolver) DefaultBranch(ctx context.Context) (*gitRefResolver, error) {
	vcsrepo := backend.Repos.CachedVCS(r.repo)
	ref, err := vcsrepo.GitCmdRaw(ctx, []string{"symbolic-ref", "HEAD"})

	// If we fail to get the default branch due to cloning, we return nothing.
	if err != nil {
		if vcs.IsCloneInProgress(err) {
			return nil, nil
		}
		return nil, err
	}
	return &gitRefResolver{repo: r, name: strings.TrimSpace(ref)}, nil
}

func (r *repositoryResolver) Language() string {
	return r.repo.Language
}

func (r *repositoryResolver) Enabled() bool { return r.repo.Enabled }

func (r *repositoryResolver) CreatedAt() string {
	return r.repo.CreatedAt.Format(time.RFC3339)
}

func (r *repositoryResolver) UpdatedAt() *string {
	if r.repo.UpdatedAt != nil {
		t := r.repo.UpdatedAt.Format(time.RFC3339)
		return &t
	}
	return nil
}

func (r *repositoryResolver) URL() string { return "/" + string(r.repo.URI) }

func (r *repositoryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return externallink.Repository(ctx, r.repo)
}

func (r *repositoryResolver) ListTotalRefs(ctx context.Context) (*totalRefListResolver, error) {
	totalRefs, err := backend.Defs.ListTotalRefs(ctx, r.repo.URI)
	if err != nil {
		return nil, err
	}
	originalLength := len(totalRefs)

	// Limit total references to 250 to prevent the many db.Repos.Get
	// operations from taking too long.
	sort.Slice(totalRefs, func(i, j int) bool {
		return totalRefs[i] < totalRefs[j]
	})
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
	for _, refRepo := range totalRefs {
		run.Acquire()
		go func(refRepo api.RepoID) {
			defer func() {
				if r := recover(); r != nil {
					run.Error(fmt.Errorf("recover: %v", r))
				}
				run.Release()
			}()
			resolver, err := repositoryByIDInt32(ctx, refRepo)
			if err != nil {
				run.Error(err)
				return
			}
			resolversMu.Lock()
			resolvers = append(resolvers, resolver)
			resolversMu.Unlock()
		}(refRepo)
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

func (*schemaResolver) AddPhabricatorRepo(ctx context.Context, args *struct {
	Callsign string
	URI      string
	URL      string
}) (*EmptyResponse, error) {
	_, err := db.Phabricator.CreateIfNotExists(ctx, args.Callsign, api.RepoURI(args.URI), args.URL)
	if err != nil {
		log15.Error("adding phabricator repo", "callsign", args.Callsign, "uri", args.URI, "url", args.URL)
	}
	return nil, err
}
