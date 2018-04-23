package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

func gitCommitByID(ctx context.Context, id graphql.ID) (*gitCommitResolver, error) {
	repoID, commitID, err := unmarshalGitCommitID(id)
	if err != nil {
		return nil, err
	}
	repo, err := repositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return repo.Commit(ctx, &struct{ Rev string }{Rev: string(commitID)})
}

type gitCommitResolver struct {
	// Either repoID or repo must be set.
	repoID api.RepoID // TODO!(sqs): can remove?
	repo   *repositoryResolver

	// inputRev is the Git revspec that the user originally requested that resolved to this Git commit. It is used
	// to avoid redirecting a user browsing a revision "mybranch" to the absolute commit ID as they follow links in the UI.
	inputRev *string

	oid       gitObjectID
	author    signatureResolver
	committer *signatureResolver
	message   string
	parents   []api.CommitID
}

func toGitCommitResolver(repo *repositoryResolver, commit *vcs.Commit) *gitCommitResolver {
	authorResolver := toSignatureResolver(&commit.Author)
	return &gitCommitResolver{
		repo:      repo,
		oid:       gitObjectID(commit.ID),
		author:    *authorResolver,
		committer: toSignatureResolver(commit.Committer),
		message:   commit.Message,
		parents:   commit.Parents,
	}
}

// gitCommitGQLID is a type used for marshaling and unmarshaling a Git commit's
// GraphQL ID.
type gitCommitGQLID struct {
	Repository graphql.ID  `json:"r"`
	CommitID   gitObjectID `json:"c"`
}

func marshalGitCommitID(repo graphql.ID, commitID gitObjectID) graphql.ID {
	return relay.MarshalID("GitCommit", gitCommitGQLID{Repository: repo, CommitID: commitID})
}

func unmarshalGitCommitID(id graphql.ID) (repoID graphql.ID, commitID gitObjectID, err error) {
	var spec gitCommitGQLID
	err = relay.UnmarshalSpec(id, &spec)
	return spec.Repository, spec.CommitID, err
}

func (r *gitCommitResolver) ID() graphql.ID { return marshalGitCommitID(r.repo.ID(), r.oid) }

func (r *gitCommitResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	if r.repo != nil {
		return r.repo, nil
	}
	return repositoryByIDInt32(ctx, r.repoID)
}

func (r *gitCommitResolver) repositoryGraphQLID() graphql.ID {
	if r.repo != nil {
		return r.repo.ID()
	}
	return marshalRepositoryID(r.repoID)
}

func (r *gitCommitResolver) repositoryDatabaseID() api.RepoID {
	if r.repo != nil {
		return r.repo.repo.ID
	}
	return r.repoID
}

func (r *gitCommitResolver) OID() gitObjectID              { return r.oid }
func (r *gitCommitResolver) AbbreviatedOID() string        { return string(r.oid)[:7] }
func (r *gitCommitResolver) Author() *signatureResolver    { return &r.author }
func (r *gitCommitResolver) Committer() *signatureResolver { return r.committer }
func (r *gitCommitResolver) Message() string               { return r.message }
func (r *gitCommitResolver) Subject() string               { return gitCommitSubject(r.message) }
func (r *gitCommitResolver) Body() *string {
	body := gitCommitBody(r.message)
	if body == "" {
		return nil
	}
	return &body
}

func (r *gitCommitResolver) Parents(ctx context.Context) ([]*gitCommitResolver, error) {
	resolvers := make([]*gitCommitResolver, len(r.parents))
	for i, parent := range r.parents {
		var err error
		resolvers[i], err = r.repo.Commit(ctx, &struct{ Rev string }{Rev: string(parent)})
		if err != nil {
			return nil, err
		}
	}
	return resolvers, nil
}

func (r *gitCommitResolver) URL() string { return r.repo.URL() + "/-/commit/" + string(r.oid) }

func (r *gitCommitResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return externallink.Commit(ctx, r.repo.repo, api.CommitID(r.oid))
}

func (r *gitCommitResolver) Tree(ctx context.Context, args *struct {
	Path      string
	Recursive bool
}) (*treeResolver, error) {
	return makeTreeResolver(ctx, r, args.Path, args.Recursive)
}

func (r *gitCommitResolver) File(ctx context.Context, args *struct {
	Path string
}) (*fileResolver, error) {
	return &fileResolver{
		commit: r,
		path:   args.Path,
	}, nil
}

func (r *gitCommitResolver) Languages(ctx context.Context) ([]string, error) {
	inventory, err := backend.Repos.GetInventory(ctx, r.repo.repo, api.CommitID(r.oid))
	if err != nil {
		return nil, err
	}

	names := make([]string, len(inventory.Languages))
	for i, l := range inventory.Languages {
		names[i] = l.Name
	}
	return names, nil
}

func (r *gitCommitResolver) Ancestors(ctx context.Context, args *struct {
	connectionArgs
	Query *string
	Path  *string
}) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		revisionRange: string(r.oid),
		first:         args.connectionArgs.First,
		query:         args.Query,
		path:          args.Path,
		repo:          r.repo,
	}
}

func (r *gitCommitResolver) BehindAhead(ctx context.Context, args *struct {
	Revspec string
}) (*behindAheadCountsResolver, error) {
	vcsrepo := backend.Repos.CachedVCS(r.repo.repo)
	counts, err := vcsrepo.BehindAhead(ctx, args.Revspec, string(r.oid))
	if err != nil {
		return nil, err
	}
	return &behindAheadCountsResolver{
		behind: int32(counts.Behind),
		ahead:  int32(counts.Ahead),
	}, nil
}

type behindAheadCountsResolver struct{ behind, ahead int32 }

func (r *behindAheadCountsResolver) Behind() int32 { return r.behind }
func (r *behindAheadCountsResolver) Ahead() int32  { return r.ahead }

func (r *gitCommitResolver) revForURL() string {
	if r.inputRev != nil && *r.inputRev != "" {
		return escapeRevspecForURL(*r.inputRev)
	}
	return string(r.oid)
}

func (r *gitCommitResolver) repoRevURL() string {
	url := r.repo.URL()
	var rev string
	if r.inputRev != nil {
		rev = *r.inputRev // use the original input rev from the user
	} else {
		rev = string(r.oid)
	}
	if rev != "" {
		return url + "@" + rev
	}
	return url
}

// gitCommitBody returns the first line of the Git commit message.
func gitCommitSubject(message string) string {
	i := strings.Index(message, "\n")
	if i == -1 {
		return message
	}
	return message[:i]
}

// gitCommitBody returns the contents of the Git commit message after the subject.
func gitCommitBody(message string) string {
	i := strings.Index(message, "\n")
	if i == -1 {
		return ""
	}
	return strings.TrimSpace(message[i:])
}
