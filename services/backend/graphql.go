package backend

import (
	"context"
	"errors"
	"path"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"

	"sourcegraph.com/sourcegraph/sourcegraph/api"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var GraphQLSchema *graphql.Schema

func init() {
	var err error
	GraphQLSchema, err = graphql.ParseSchema(api.Schema, &queryResolver{})
	if err != nil {
		panic(err)
	}
}

type nodeResolver interface {
	ID() graphql.ID
	ToRepository() (*repositoryResolver, bool)
	ToCommit() (*commitResolver, bool)
}

type nodeBase struct{}

func (*nodeBase) ToRepository() (*repositoryResolver, bool) {
	return nil, false
}

func (*nodeBase) ToCommit() (*commitResolver, bool) {
	return nil, false
}

type queryResolver struct{}

func (r *queryResolver) Root() *rootResolver {
	return &rootResolver{}
}

func (r *queryResolver) Node(ctx context.Context, args *struct{ ID graphql.ID }) (nodeResolver, error) {
	switch relay.UnmarshalKind(args.ID) {
	case "Repository":
		return repositoryByID(ctx, args.ID)
	case "Commit":
		return commitByID(ctx, args.ID)
	default:
		return nil, errors.New("invalid id")
	}
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
	return &repositoryResolver{repo: repo}, nil
}

func commitByID(ctx context.Context, id graphql.ID) (nodeResolver, error) {
	var commit commitSpec
	if err := relay.UnmarshalSpec(id, &commit); err != nil {
		return nil, err
	}
	return &commitResolver{commit: commit}, nil
}

type rootResolver struct{}

func (r *rootResolver) Repository(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	repo, err := localstore.Repos.GetByURI(ctx, args.URI)
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

type repositoryResolver struct {
	nodeBase
	repo *sourcegraph.Repo
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

func (r *repositoryResolver) Commit(ctx context.Context, args *struct{ Rev string }) (*commitResolver, error) {
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
		Rev:  args.Rev,
	})
	if err != nil {
		return nil, err
	}
	return &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}, nil
}

func (r *repositoryResolver) Latest(ctx context.Context) (*commitResolver, error) {
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
	})
	if err != nil {
		return nil, err
	}
	return &commitResolver{commit: commitSpec{r.repo.ID, rev.CommitID}}, nil
}

type commitSpec struct {
	RepoID   int32
	CommitID string
}

type commitResolver struct {
	nodeBase
	commit commitSpec
}

func (r *commitResolver) ToCommit() (*commitResolver, bool) {
	return r, true
}

func (r *commitResolver) ID() graphql.ID {
	return relay.MarshalID("Commit", r.commit)
}

func (r *commitResolver) SHA1() string {
	return r.commit.CommitID
}

func (r *commitResolver) Tree(ctx context.Context, args *struct{ Path string }) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, args.Path)
}

type treeResolver struct {
	commit commitSpec
	path   string
	tree   *sourcegraph.TreeEntry
}

func makeTreeResolver(ctx context.Context, commit commitSpec, path string) (*treeResolver, error) {
	tree, err := RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{
				Repo:     commit.RepoID,
				CommitID: commit.CommitID,
			},
			Path: path,
		},
	})
	if err != nil {
		if err.Error() == "file does not exist" {
			return nil, nil
		}
		return nil, err
	}

	return &treeResolver{
		commit: commit,
		path:   path,
		tree:   tree,
	}, nil
}

func (r *treeResolver) Directories() []*entryResolver {
	var l []*entryResolver
	for _, entry := range r.tree.Entries {
		if entry.Type == sourcegraph.DirEntry {
			l = append(l, &entryResolver{
				commit: r.commit,
				path:   path.Join(r.path, entry.Name),
				entry:  entry,
			})
		}
	}
	return l
}

func (r *treeResolver) Files() []*entryResolver {
	var l []*entryResolver
	for _, entry := range r.tree.Entries {
		if entry.Type != sourcegraph.DirEntry {
			l = append(l, &entryResolver{
				commit: r.commit,
				path:   path.Join(r.path, entry.Name),
				entry:  entry,
			})
		}
	}
	return l
}

type entryResolver struct {
	commit commitSpec
	path   string
	entry  *sourcegraph.BasicTreeEntry
}

func (r *entryResolver) Name() string {
	return r.entry.Name
}

func (r *entryResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, r.path)
}

func (r *entryResolver) Content() *blobResolver {
	return &blobResolver{}
}

type blobResolver struct{}

func (r *blobResolver) Bytes(ctx context.Context) (string, error) {
	return "TODO", nil
}
