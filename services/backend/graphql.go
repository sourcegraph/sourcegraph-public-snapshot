package backend

import (
	"context"
	"path"
	"strconv"

	graphql "github.com/neelance/graphql-go"

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

type queryResolver struct{}

func (r *queryResolver) Root() *rootResolver {
	return &rootResolver{}
}

type rootResolver struct{}

func (r *rootResolver) Repository(ctx context.Context, args *struct{ ID string }) (*repositoryResolver, error) {
	id, err := strconv.Atoi(args.ID)
	if err != nil {
		return nil, err
	}
	repo, err := localstore.Repos.Get(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func (r *rootResolver) RepositoryByURI(ctx context.Context, args *struct{ URI string }) (*repositoryResolver, error) {
	repo, err := localstore.Repos.GetByURI(ctx, args.URI)
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

type repositoryResolver struct {
	repo *sourcegraph.Repo
}

func (r *repositoryResolver) ID() string {
	return strconv.Itoa(int(r.repo.ID))
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
	return &commitResolver{r.repo.ID, rev.CommitID}, nil
}

func (r *repositoryResolver) Latest(ctx context.Context) (*commitResolver, error) {
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
		Repo: r.repo.ID,
	})
	if err != nil {
		return nil, err
	}
	return &commitResolver{r.repo.ID, rev.CommitID}, nil
}

type commitResolver struct {
	repoID   int32
	commitID string
}

func (r *commitResolver) ID() string {
	return r.commitID
}

func (r *commitResolver) Tree(ctx context.Context, args *struct{ Path string }) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.repoID, r.commitID, args.Path)
}

type treeResolver struct {
	repoID   int32
	commitID string
	path     string
	tree     *sourcegraph.TreeEntry
}

func makeTreeResolver(ctx context.Context, repoID int32, commitID string, path string) (*treeResolver, error) {
	tree, err := RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{
				Repo:     repoID,
				CommitID: commitID,
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
		repoID:   repoID,
		commitID: commitID,
		path:     path,
		tree:     tree,
	}, nil
}

func (r *treeResolver) Directories() []*entryResolver {
	var l []*entryResolver
	for _, entry := range r.tree.Entries {
		if entry.Type == sourcegraph.DirEntry {
			l = append(l, &entryResolver{
				repoID:   r.repoID,
				commitID: r.commitID,
				path:     path.Join(r.path, entry.Name),
				entry:    entry,
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
				repoID:   r.repoID,
				commitID: r.commitID,
				path:     path.Join(r.path, entry.Name),
				entry:    entry,
			})
		}
	}
	return l
}

type entryResolver struct {
	repoID   int32
	commitID string
	path     string
	entry    *sourcegraph.BasicTreeEntry
}

func (r *entryResolver) Name() string {
	return r.entry.Name
}

func (r *entryResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.repoID, r.commitID, r.path)
}

func (r *entryResolver) Content() *blobResolver {
	return &blobResolver{}
}

type blobResolver struct{}

func (r *blobResolver) Bytes(ctx context.Context) (string, error) {
	return "TODO", nil
}
