package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type RepositoryResolver struct {
	logger    log.Logger
	hydration sync.Once
	err       error

	// Invariant: Name and ID of RepoMatch are always set and safe to use. They are
	// used to hydrate the inner repo, and should always be the same as the name and
	// id of the inner repo, but referring to the inner repo directly is unsafe
	// because it may cause a race during hydration.
	result.RepoMatch

	db              database.DB
	gitserverClient gitserver.Client

	// innerRepo may only contain ID and Name information.
	// To access any other repo information, use repo() instead.
	innerRepo *types.Repo

	defaultBranchOnce sync.Once
	defaultBranch     *GitRefResolver
	defaultBranchErr  error
}

func NewRepositoryResolver(db database.DB, client gitserver.Client, repo *types.Repo) *RepositoryResolver {
	// Protect against a nil repo
	var name api.RepoName
	var id api.RepoID
	if repo != nil {
		name = repo.Name
		id = repo.ID
	}

	return &RepositoryResolver{
		db:              db,
		innerRepo:       repo,
		gitserverClient: client,
		RepoMatch: result.RepoMatch{
			Name: name,
			ID:   id,
		},
		logger: log.Scoped("repositoryResolver", "resolve a specific repository").
			With(log.Object("repo",
				log.String("name", string(name)),
				log.Int32("id", int32(id)))),
	}
}

func (r *RepositoryResolver) ID() graphql.ID {
	return MarshalRepositoryID(r.IDInt32())
}

func (r *RepositoryResolver) IDInt32() api.RepoID {
	return r.RepoMatch.ID
}

func (r *RepositoryResolver) EmbeddingExists(ctx context.Context) (bool, error) {
	if !conf.EmbeddingsEnabled() {
		return false, nil
	}

	return r.db.Repos().RepoEmbeddingExists(ctx, r.IDInt32())
}

func MarshalRepositoryID(repo api.RepoID) graphql.ID { return relay.MarshalID("Repository", repo) }

func UnmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

// repo makes sure the repo is hydrated before returning it.
func (r *RepositoryResolver) repo(ctx context.Context) (*types.Repo, error) {
	err := r.hydrate(ctx)
	return r.innerRepo, err
}

func (r *RepositoryResolver) RepoName() api.RepoName {
	return r.RepoMatch.Name
}

func (r *RepositoryResolver) Name() string {
	return string(r.RepoMatch.Name)
}

func (r *RepositoryResolver) ExternalRepo(ctx context.Context) (*api.ExternalRepoSpec, error) {
	repo, err := r.repo(ctx)
	return &repo.ExternalRepo, err
}

func (r *RepositoryResolver) IsFork(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Fork, err
}

func (r *RepositoryResolver) IsArchived(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Archived, err
}

func (r *RepositoryResolver) IsPrivate(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Private, err
}

func (r *RepositoryResolver) URI(ctx context.Context) (string, error) {
	repo, err := r.repo(ctx)
	return repo.URI, err
}

func (r *RepositoryResolver) Description(ctx context.Context) (string, error) {
	repo, err := r.repo(ctx)
	return repo.Description, err
}

func (r *RepositoryResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if err == auth.ErrMustBeSiteAdmin || err == auth.ErrNotAuthenticated {
			return false, nil // not an error
		}
		return false, err
	}
	return true, nil
}

func (r *RepositoryResolver) CloneInProgress(ctx context.Context) (bool, error) {
	return r.MirrorInfo().CloneInProgress(ctx)
}

func (r *RepositoryResolver) DiskSizeBytes(ctx context.Context) (*BigInt, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if err == auth.ErrMustBeSiteAdmin || err == auth.ErrNotAuthenticated {
			return nil, nil // not an error
		}
		return nil, err
	}
	repo, err := r.db.GitserverRepos().GetByID(ctx, r.IDInt32())
	if err != nil {
		return nil, err
	}
	size := BigInt(repo.RepoSizeBytes)
	return &size, nil
}

func (r *RepositoryResolver) BatchChanges(ctx context.Context, args *ListBatchChangesArgs) (BatchChangesConnectionResolver, error) {
	id := r.ID()
	args.Repo = &id
	return EnterpriseResolvers.batchChangesResolver.BatchChanges(ctx, args)
}

func (r *RepositoryResolver) ChangesetsStats(ctx context.Context) (RepoChangesetsStatsResolver, error) {
	id := r.ID()
	return EnterpriseResolvers.batchChangesResolver.RepoChangesetsStats(ctx, &id)
}

func (r *RepositoryResolver) BatchChangesDiffStat(ctx context.Context) (*DiffStat, error) {
	id := r.ID()
	return EnterpriseResolvers.batchChangesResolver.RepoDiffStat(ctx, &id)
}

type RepositoryCommitArgs struct {
	Rev          string
	InputRevspec *string
}

func (r *RepositoryResolver) Commit(ctx context.Context, args *RepositoryCommitArgs) (*GitCommitResolver, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "repository.commit") //nolint:staticcheck // OT is deprecated
	defer span.Finish()
	span.SetTag("commit", args.Rev)

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	commitID, err := backend.NewRepos(r.logger, r.db, r.gitserverClient).ResolveRev(ctx, repo, args.Rev)
	if err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}

	return r.CommitFromID(ctx, args, commitID)
}

func (r *RepositoryResolver) FirstEverCommit(ctx context.Context) (*GitCommitResolver, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "repository.firstEverCommit") //nolint:staticcheck // OT is deprecated
	defer span.Finish()

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	commit, err := r.gitserverClient.FirstEverCommit(ctx, authz.DefaultSubRepoPermsChecker, repo.Name)
	if err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}

	return r.CommitFromID(ctx, &RepositoryCommitArgs{}, commit.ID)
}

func (r *RepositoryResolver) CommitFromID(ctx context.Context, args *RepositoryCommitArgs, commitID api.CommitID) (*GitCommitResolver, error) {
	resolver := NewGitCommitResolver(r.db, r.gitserverClient, r, commitID, nil)
	if args.InputRevspec != nil {
		resolver.inputRev = args.InputRevspec
	} else {
		resolver.inputRev = &args.Rev
	}
	return resolver, nil
}

func (r *RepositoryResolver) DefaultBranch(ctx context.Context) (*GitRefResolver, error) {
	do := func() (*GitRefResolver, error) {
		refName, _, err := r.gitserverClient.GetDefaultBranch(ctx, r.RepoName(), false)
		if err != nil {
			return nil, err
		}
		if refName == "" {
			return nil, nil
		}
		return &GitRefResolver{repo: r, name: refName}, nil
	}
	r.defaultBranchOnce.Do(func() {
		r.defaultBranch, r.defaultBranchErr = do()
	})
	return r.defaultBranch, r.defaultBranchErr
}

func (r *RepositoryResolver) Language(ctx context.Context) (string, error) {
	// The repository language is the most common language at the HEAD commit of the repository.
	// Note: the repository database field is no longer updated as of
	// https://github.com/sourcegraph/sourcegraph/issues/2586, so we do not use it anymore and
	// instead compute the language on the fly.
	repo, err := r.repo(ctx)
	if err != nil {
		return "", err
	}

	commitID, err := backend.NewRepos(r.logger, r.db, r.gitserverClient).ResolveRev(ctx, repo, "")
	if err != nil {
		// Comment: Should we return a nil error?
		return "", err
	}

	inventory, err := backend.NewRepos(r.logger, r.db, r.gitserverClient).GetInventory(ctx, repo, commitID, false)
	if err != nil {
		return "", err
	}
	if len(inventory.Languages) == 0 {
		return "", err
	}
	return inventory.Languages[0].Name, nil
}

func (r *RepositoryResolver) Enabled() bool { return true }

// CreatedAt is deprecated and will be removed in a future release.
// No clients that we know of read this field. Additionally on performance profiles
// the marshalling of timestamps is significant in our postgres client. So we
// deprecate the fields and return fake data for created_at.
// https://github.com/sourcegraph/sourcegraph/pull/4668
func (r *RepositoryResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: time.Now()}
}

func (r *RepositoryResolver) RawCreatedAt() string {
	if r.innerRepo == nil {
		return ""
	}

	return r.innerRepo.CreatedAt.Format(time.RFC3339)
}

func (r *RepositoryResolver) UpdatedAt() *gqlutil.DateTime {
	return nil
}

func (r *RepositoryResolver) URL() string {
	return r.url().String()
}

func (r *RepositoryResolver) url() *url.URL {
	return r.RepoMatch.URL()
}

func (r *RepositoryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}
	return externallink.Repository(ctx, r.db, repo)
}

func (r *RepositoryResolver) Rev() string {
	return r.RepoMatch.Rev
}

func (r *RepositoryResolver) Label() (Markdown, error) {
	var label string
	if r.Rev() != "" {
		label = r.Name() + "@" + r.Rev()
	} else {
		label = r.Name()
	}
	text := "[" + label + "](" + r.URL() + ")"
	return Markdown(text), nil
}

func (r *RepositoryResolver) Detail() Markdown {
	return "Repository match"
}

func (r *RepositoryResolver) Matches() []*searchResultMatchResolver {
	return nil
}

func (r *RepositoryResolver) ToRepository() (*RepositoryResolver, bool) { return r, true }
func (r *RepositoryResolver) ToFileMatch() (*FileMatchResolver, bool)   { return nil, false }
func (r *RepositoryResolver) ToCommitSearchResult() (*CommitSearchResultResolver, bool) {
	return nil, false
}

func (r *RepositoryResolver) Type(ctx context.Context) (*types.Repo, error) {
	return r.repo(ctx)
}

func (r *RepositoryResolver) Stars(ctx context.Context) (int32, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return 0, err
	}
	return int32(repo.Stars), nil
}

// Deprecated: Use RepositoryResolver.Metadata instead.
func (r *RepositoryResolver) KeyValuePairs(ctx context.Context) ([]KeyValuePair, error) {
	return r.Metadata(ctx)
}

func (r *RepositoryResolver) Metadata(ctx context.Context) ([]KeyValuePair, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	kvps := make([]KeyValuePair, 0, len(repo.KeyValuePairs))
	for k, v := range repo.KeyValuePairs {
		kvps = append(kvps, KeyValuePair{key: k, value: v})
	}
	return kvps, nil
}

func (r *RepositoryResolver) hydrate(ctx context.Context) error {
	r.hydration.Do(func() {
		// Repositories with an empty creation date were created using RepoName.ToRepo(),
		// they only contain ID and name information.
		if r.innerRepo != nil && !r.innerRepo.CreatedAt.IsZero() {
			return
		}

		r.logger.Debug("RepositoryResolver.hydrate", log.String("repo.ID", string(r.IDInt32())))

		var repo *types.Repo
		repo, r.err = r.db.Repos().Get(ctx, r.IDInt32())
		if r.err == nil {
			r.innerRepo = repo
		}
	})

	return r.err
}

func (r *RepositoryResolver) IndexConfiguration(ctx context.Context) (resolverstubs.IndexConfigurationResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.IndexConfiguration(ctx, r.ID())
}

func (r *RepositoryResolver) CodeIntelligenceCommitGraph(ctx context.Context) (resolverstubs.CodeIntelligenceCommitGraphResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.CommitGraph(ctx, r.ID())
}

func (r *RepositoryResolver) CodeIntelSummary(ctx context.Context) (resolverstubs.CodeIntelRepositorySummaryResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.RepositorySummary(ctx, r.ID())
}

func (r *RepositoryResolver) PreviewGitObjectFilter(ctx context.Context, args *resolverstubs.PreviewGitObjectFilterArgs) (resolverstubs.GitObjectFilterPreviewResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.PreviewGitObjectFilter(ctx, r.ID(), args)
}

type AuthorizedUserArgs struct {
	RepositoryID graphql.ID
	Permission   string
	First        int32
	After        *string
}

type RepoAuthorizedUserArgs struct {
	RepositoryID graphql.ID
	*AuthorizedUserArgs
}

func (r *RepositoryResolver) AuthorizedUsers(ctx context.Context, args *AuthorizedUserArgs) (UserConnectionResolver, error) {
	return EnterpriseResolvers.authzResolver.AuthorizedUsers(ctx, &RepoAuthorizedUserArgs{
		RepositoryID:       r.ID(),
		AuthorizedUserArgs: args,
	})
}

func (r *RepositoryResolver) PermissionsInfo(ctx context.Context) (PermissionsInfoResolver, error) {
	return EnterpriseResolvers.authzResolver.RepositoryPermissionsInfo(ctx, r.ID())
}

func (r *schemaResolver) AddPhabricatorRepo(ctx context.Context, args *struct {
	Callsign string
	Name     *string
	// TODO(chris): Remove URI in favor of Name.
	URI *string
	URL string
},
) (*EmptyResponse, error) {
	if args.Name != nil {
		args.URI = args.Name
	}

	_, err := r.db.Phabricator().CreateIfNotExists(ctx, args.Callsign, api.RepoName(*args.URI), args.URL)
	if err != nil {
		r.logger.Error("adding phabricator repo", log.String("callsign", args.Callsign), log.Stringp("name", args.URI), log.String("url", args.URL))
	}
	return nil, err
}

func (r *schemaResolver) ResolvePhabricatorDiff(ctx context.Context, args *struct {
	RepoName    string
	DiffID      int32
	BaseRev     string
	Patch       *string
	AuthorName  *string
	AuthorEmail *string
	Description *string
	Date        *string
},
) (*GitCommitResolver, error) {
	db := r.db
	repo, err := db.Repos().GetByName(ctx, api.RepoName(args.RepoName))
	if err != nil {
		return nil, err
	}
	targetRef := fmt.Sprintf("phabricator/diff/%d", args.DiffID)
	getCommit := func() (*GitCommitResolver, error) {
		// We first check via the vcsrepo api so that we can toggle
		// NoEnsureRevision. We do this, otherwise RepositoryResolver.Commit
		// will try and fetch it from the remote host. However, this is not on
		// the remote host since we created it.
		_, err = r.gitserverClient.ResolveRevision(ctx, repo.Name, targetRef, gitserver.ResolveRevisionOptions{
			NoEnsureRevision: true,
		})
		if err != nil {
			return nil, err
		}
		r := NewRepositoryResolver(db, r.gitserverClient, repo)
		return r.Commit(ctx, &RepositoryCommitArgs{Rev: targetRef})
	}

	// If we already created the commit
	if commit, err := getCommit(); commit != nil || (err != nil && !errors.HasType(err, &gitdomain.RevisionNotFoundError{})) {
		return commit, err
	}

	origin := ""
	if phabRepo, err := db.Phabricator().GetByName(ctx, api.RepoName(args.RepoName)); err == nil {
		origin = phabRepo.URL
	}

	if origin == "" {
		return nil, errors.New("unable to resolve the origin of the phabricator instance")
	}

	client, clientErr := makePhabClientForOrigin(ctx, r.logger, db, origin)

	patch := ""
	if args.Patch != nil {
		patch = *args.Patch
	} else if client == nil {
		return nil, clientErr
	} else {
		diff, err := client.GetRawDiff(ctx, int(args.DiffID))
		// No diff contents were given and we couldn't fetch them
		if err != nil {
			return nil, err
		}

		patch = diff
	}

	var info *phabricator.DiffInfo
	if client != nil && (args.AuthorEmail == nil || args.AuthorName == nil || args.Date == nil) {
		info, err = client.GetDiffInfo(ctx, int(args.DiffID))
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

	_, err = r.gitserverClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(args.RepoName),
		BaseCommit: api.CommitID(args.BaseRev),
		TargetRef:  targetRef,
		Patch:      []byte(patch),
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

func makePhabClientForOrigin(ctx context.Context, logger log.Logger, db database.DB, origin string) (*phabricator.Client, error) {
	opt := database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindPhabricator},
		LimitOffset: &database.LimitOffset{
			Limit: 500, // The number is randomly chosen
		},
	}
	for {
		svcs, err := db.ExternalServices().List(ctx, opt)
		if err != nil {
			return nil, errors.Wrap(err, "list")
		}
		if len(svcs) == 0 {
			break // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advance the cursor

		for _, svc := range svcs {
			cfg, err := extsvc.ParseEncryptableConfig(ctx, svc.Kind, svc.Config)
			if err != nil {
				return nil, errors.Wrap(err, "parse config")
			}

			var conn *schema.PhabricatorConnection
			switch c := cfg.(type) {
			case *schema.PhabricatorConnection:
				conn = c
			default:
				err := errors.Errorf("want *schema.PhabricatorConnection but got %T", cfg)
				logger.Error("makePhabClientForOrigin", log.Error(err))
				continue
			}

			if conn.Url != origin {
				continue
			}

			if conn.Token == "" {
				return nil, errors.Errorf("no phabricator token was given for: %s", origin)
			}

			return phabricator.NewClient(ctx, conn.Url, conn.Token, nil)
		}

		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}

	return nil, errors.Errorf("no phabricator was configured for: %s", origin)
}

type KeyValuePair struct {
	key   string
	value *string
}

func (k KeyValuePair) Key() string {
	return k.key
}

func (k KeyValuePair) Value() *string {
	return k.value
}

// Deprecated: Use AddRepoMetadata instead.
func (r *schemaResolver) AddRepoKeyValuePair(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	return r.AddRepoMetadata(ctx, args)
}

func (r *schemaResolver) AddRepoMetadata(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", false) {
		return nil, errors.New("'repository-metadata' feature flag is not enabled")
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	return &EmptyResponse{}, r.db.RepoKVPs().Create(ctx, repoID, database.KeyValuePair{Key: args.Key, Value: args.Value})
}

// Deprecated: Use UpdateRepoMetadata instead.
func (r *schemaResolver) UpdateRepoKeyValuePair(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	return r.UpdateRepoMetadata(ctx, args)
}

func (r *schemaResolver) UpdateRepoMetadata(ctx context.Context, args struct {
	Repo  graphql.ID
	Key   string
	Value *string
},
) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", false) {
		return nil, errors.New("'repository-metadata' feature flag is not enabled")
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	_, err = r.db.RepoKVPs().Update(ctx, repoID, database.KeyValuePair{Key: args.Key, Value: args.Value})
	return &EmptyResponse{}, err
}

// Deprecated: Use DeleteRepoMetadata instead.
func (r *schemaResolver) DeleteRepoKeyValuePair(ctx context.Context, args struct {
	Repo graphql.ID
	Key  string
},
) (*EmptyResponse, error) {
	return r.DeleteRepoMetadata(ctx, args)
}

func (r *schemaResolver) DeleteRepoMetadata(ctx context.Context, args struct {
	Repo graphql.ID
	Key  string
},
) (*EmptyResponse, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return &EmptyResponse{}, err
	}

	if !featureflag.FromContext(ctx).GetBoolOr("repository-metadata", false) {
		return nil, errors.New("'repository-metadata' feature flag is not enabled")
	}

	repoID, err := UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return &EmptyResponse{}, err
	}

	return &EmptyResponse{}, r.db.RepoKVPs().Delete(ctx, repoID, args.Key)
}

func (r *RepositoryResolver) IngestedCodeowners(ctx context.Context) (CodeownersIngestedFileResolver, error) {
	return EnterpriseResolvers.ownResolver.RepoIngestedCodeowners(ctx, r.IDInt32())
}
