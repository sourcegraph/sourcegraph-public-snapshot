package graphqlbackend

import (
	"context"
	"errors"
	"path"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

type treeResolver struct {
	commit commitSpec
	path   string
	tree   *sourcegraph.TreeEntry
}

func makeTreeResolver(ctx context.Context, commit commitSpec, path string, recursive bool) (*treeResolver, error) {
	if recursive {
		if path != "" {
			return nil, errors.New("not implemented")
		}
		list, err := backend.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{ // TODO merge with RepoTree.Get
			Rev: sourcegraph.RepoRevSpec{
				Repo:     commit.RepoID,
				CommitID: commit.CommitID,
			},
		})
		if err != nil {
			if err.Error() == "file does not exist" { // TODO proper error value
				return nil, nil
			}
			return nil, err
		}

		entries := make([]*sourcegraph.BasicTreeEntry, len(list.Files))
		for i, name := range list.Files {
			entries[i] = &sourcegraph.BasicTreeEntry{Name: name, Type: sourcegraph.FileEntry}
		}
		return &treeResolver{
			commit: commit,
			path:   path,
			tree: &sourcegraph.TreeEntry{
				BasicTreeEntry: &sourcegraph.BasicTreeEntry{
					Entries: entries,
				},
			},
		}, nil
	}

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
		if err.Error() == "file does not exist" { // TODO proper error value
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
				name:   entry.Name,
				path:   path.Join(r.path, entry.Name),
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
				name:   entry.Name,
				path:   path.Join(r.path, entry.Name),
			})
		}
	}
	return l
}

type entryResolver struct {
	commit commitSpec
	name   string
	path   string
}

func (r *entryResolver) Name() string {
	return r.name
}

func (r *entryResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, r.path, false)
}

func (r *entryResolver) Content(ctx context.Context) (string, error) {
	file, err := backend.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{
				Repo:     r.commit.RepoID,
				CommitID: r.commit.CommitID,
			},
			Path: r.path,
		},
	})
	if err != nil {
		return "", err
	}
	return string(file.Contents), nil
}
