package graphqlbackend

import (
	"context"
	"errors"
	"os"
	"path"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type treeResolver struct {
	commit  commitSpec
	path    string
	entries []os.FileInfo
}

func makeTreeResolver(ctx context.Context, commit commitSpec, path string, recursive bool) (*treeResolver, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, commit.RepoID)
	if err != nil {
		return nil, err
	}

	if recursive && path != "" {
		return nil, errors.New("not implemented")
	}

	entries, err := vcsrepo.ReadDir(ctx, vcs.CommitID(commit.CommitID), path, recursive)
	if err != nil {
		if err.Error() == "file does not exist" { // TODO proper error value
			return nil, nil
		}
		return nil, err
	}

	if recursive {
		var files []os.FileInfo
		for _, info := range entries {
			if !info.Mode().IsDir() {
				files = append(files, info)
			}
		}
		entries = files
	}
	return &treeResolver{
		commit:  commit,
		path:    path,
		entries: entries,
	}, nil
}

func (r *treeResolver) Directories() []*fileResolver {
	var l []*fileResolver
	for _, entry := range r.entries {
		if entry.Mode().IsDir() {
			l = append(l, &fileResolver{
				commit: r.commit,
				name:   entry.Name(),
				path:   path.Join(r.path, entry.Name()),
			})
		}
	}
	return l
}

func (r *treeResolver) Files() []*fileResolver {
	var l []*fileResolver
	for _, entry := range r.entries {
		if !entry.Mode().IsDir() {
			l = append(l, &fileResolver{
				commit: r.commit,
				name:   entry.Name(),
				path:   path.Join(r.path, entry.Name()),
			})
		}
	}
	return l
}

func (r *fileResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, r.path, false)
}
