package graphqlbackend

import (
	"context"
	"path"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
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
	repoID int32
	repo   *repositoryResolver

	oid       gitObjectID
	author    signatureResolver
	committer *signatureResolver
	message   string
}

func toGitCommitResolver(repo *repositoryResolver, commit *vcs.Commit) *gitCommitResolver {
	authorResolver := toSignatureResolver(&commit.Author)
	return &gitCommitResolver{
		repo:      repo,
		oid:       gitObjectID(commit.ID),
		author:    *authorResolver,
		committer: toSignatureResolver(commit.Committer),
		message:   commit.Message,
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

func (r *gitCommitResolver) repositoryID() graphql.ID {
	if r.repo != nil {
		return r.repo.ID()
	}
	return marshalRepositoryID(r.repoID)
}

func (r *gitCommitResolver) repositoryIDInt32() int32 {
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
		name:   path.Base(args.Path),
		path:   args.Path,
	}, nil
}

func (r *gitCommitResolver) Languages(ctx context.Context) ([]string, error) {
	inventory, err := backend.Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{
		Repo:     r.repo.repo.ID,
		CommitID: string(r.oid),
	})
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
}) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		headCommitID: string(r.oid),
		first:        args.connectionArgs.First,
		query:        args.Query,
		repo:         r.repo,
	}
}
