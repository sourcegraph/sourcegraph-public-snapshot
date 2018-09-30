package graphqlbackend

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	log15 "gopkg.in/inconshreveable/log15.v2"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/neelance/parallel"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/phabricator"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
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

func (r *repositoryResolver) Name() string {
	return string(r.repo.URI)
}

// TODO(chris): Remove URI in favor of Name.
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
	// Asking gitserver may trigger a clone of the repo, so ensure it is
	// enabled.
	if !r.repo.Enabled {
		return false, nil
	}

	return r.MirrorInfo().CloneInProgress(ctx)
}

type repositoryCommitArgs struct {
	Rev          string
	InputRevspec *string
}

func (r *repositoryResolver) Commit(ctx context.Context, args *repositoryCommitArgs) (*gitCommitResolver, error) {
	// Asking gitserver may trigger a clone of the repo, so ensure it is
	// enabled.
	if !r.repo.Enabled {
		return nil, nil
	}

	commitID, err := backend.Repos.ResolveRev(ctx, r.repo, args.Rev)
	if err != nil {
		if git.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	commit, err := backend.Repos.GetCommit(ctx, r.repo, commitID)
	if commit == nil || err != nil {
		return nil, err
	}

	resolver := toGitCommitResolver(r, commit)
	if args.InputRevspec != nil {
		resolver.inputRev = args.InputRevspec
	} else {
		resolver.inputRev = &args.Rev
	}
	return resolver, nil
}

func (r *repositoryResolver) LastIndexedRevOrLatest(ctx context.Context) (*gitCommitResolver, error) {
	// This method is a stopgap until we no longer require git:// URIs on the client which include rev data.
	// THIS RESOLVER WILL BE REMOVED SOON, DO NOT USE IT!!!
	if r.repo.IndexedRevision != nil && *r.repo.IndexedRevision != "" {
		return r.Commit(ctx, &repositoryCommitArgs{Rev: string(*r.repo.IndexedRevision)})
	}
	return r.Commit(ctx, &repositoryCommitArgs{Rev: "HEAD"})
}

func (r *repositoryResolver) DefaultBranch(ctx context.Context) (*gitRefResolver, error) {
	// Asking gitserver may trigger a clone of the repo, so ensure it is
	// enabled.
	if !r.repo.Enabled {
		return nil, nil
	}

	refBytes, _, exitCode, err := git.ExecSafe(ctx, backend.CachedGitRepo(r.repo), []string{"symbolic-ref", "HEAD"})
	refName := string(bytes.TrimSpace(refBytes))

	if err == nil && exitCode == 0 {
		// Check that our repo is not empty
		_, err = git.ResolveRevision(ctx, backend.CachedGitRepo(r.repo), nil, "HEAD", &git.ResolveRevisionOptions{NoEnsureRevision: true})
	}

	// If we fail to get the default branch due to cloning or being empty, we return nothing.
	if err != nil {
		if vcs.IsCloneInProgress(err) || git.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &gitRefResolver{repo: r, name: refName}, nil
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
	Name     *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
	URL string
}) (*EmptyResponse, error) {
	if args.Name != nil {
		args.URI = args.Name
	}

	_, err := db.Phabricator.CreateIfNotExists(ctx, args.Callsign, api.RepoURI(*args.URI), args.URL)
	if err != nil {
		log15.Error("adding phabricator repo", "callsign", args.Callsign, "uri", args.URI, "url", args.URL)
	}
	return nil, err
}

func (*schemaResolver) ResolvePhabricatorDiff(ctx context.Context, args *struct {
	RepoName    string
	DiffID      int32
	BaseRev     string
	Patch       *string
	AuthorName  *string
	AuthorEmail *string
	Description *string
	Date        *string
}) (*gitCommitResolver, error) {
	repo, err := db.Repos.GetByURI(ctx, api.RepoURI(args.RepoName))
	if err != nil {
		return nil, err
	}
	targetRef := fmt.Sprintf("phabricator/diff/%d", args.DiffID)
	getCommit := func() (*gitCommitResolver, error) {
		// We first check via the vcsrepo api so that we can toggle
		// NoEnsureRevision. We do this, otherwise repositoryResolver.Commit
		// will try and fetch it from the remote host. However, this is not on
		// the remote host since we created it.
		_, err := git.ResolveRevision(ctx, backend.CachedGitRepo(repo), nil, targetRef, &git.ResolveRevisionOptions{
			NoEnsureRevision: true,
		})
		if err != nil {
			return nil, err
		}
		r := &repositoryResolver{repo: repo}
		return r.Commit(ctx, &repositoryCommitArgs{Rev: targetRef})
	}

	// If we already created the commit
	if commit, err := getCommit(); commit != nil || (err != nil && !git.IsRevisionNotFound(err)) {
		return commit, err
	}

	origin := ""
	if phabRepo, err := db.Phabricator.GetByURI(ctx, api.RepoURI(args.RepoName)); err == nil {
		origin = phabRepo.URL
	}

	if origin == "" {
		return nil, errors.New("unable to resolve the origin of the phabricator instance")
	}

	client, clientErr := makePhabClientForOrigin(origin)

	patch := ""
	if args.Patch != nil {
		patch = *args.Patch
	} else if client == nil {
		return nil, clientErr
	} else {
		diff, err := client.GetRawDiff(int(args.DiffID))
		// No diff contents were given and we couldn't fetch them
		if err != nil {
			return nil, err
		}

		patch = diff
	}

	var info *phabricator.DiffInfo
	if client != nil && (args.AuthorEmail == nil || args.AuthorName == nil || args.Date == nil) {
		info, err = client.GetDiffInfo(int(args.DiffID))
		// Not all the information was given and we couldn't fetch it
		if err != nil {
			return nil, err
		}
	} else {
		var description, authorName, authorEmail string
		if args.Description != nil {
			description = *args.Description
		}
		if args.AuthorName != nil {
			authorName = *args.AuthorName
		}
		if args.AuthorEmail != nil {
			authorEmail = *args.AuthorEmail
		}
		date, err := phabricator.ParseDate(*args.Date)
		if err != nil {
			return nil, err
		}

		info = &phabricator.DiffInfo{
			AuthorName:  authorName,
			AuthorEmail: authorEmail,
			Message:     description,
			Date:        *date,
		}
	}

	_, err = gitserver.DefaultClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoURI(args.RepoName),
		BaseCommit: api.CommitID(args.BaseRev),
		TargetRef:  targetRef,
		Patch:      patch,
		CommitInfo: protocol.PatchCommitInfo{
			AuthorName:  info.AuthorName,
			AuthorEmail: info.AuthorEmail,
			Message:     info.Message,
			Date:        info.Date,
		},
	})
	if err != nil {
		return nil, err
	}

	return getCommit()
}

func makePhabClientForOrigin(origin string) (*phabricator.Client, error) {
	phabs := conf.Get().Phabricator
	for _, phab := range phabs {
		if phab.Url != origin {
			continue
		}

		if phab.Token == "" {
			return nil, errors.Errorf("no phabricator token was given for: %s", origin)
		}

		return phabricator.NewClient(phab.Url, phab.Token), nil
	}

	return nil, errors.Errorf("no phabricator was configured for: %s", origin)
}
