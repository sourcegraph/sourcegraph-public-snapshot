package graphqlbackend

import (
	"context"
	"path"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

type treeResolver struct {
	commit commitSpec
	path   string
	tree   *sourcegraph.TreeEntry
}

func makeTreeResolver(ctx context.Context, commit commitSpec, path string) (*treeResolver, error) {
	tree, err := backend.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
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
