package graphqlbackend

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type treeResolver struct {
	commit *gitCommitResolver

	path    string
	entries []os.FileInfo
}

func makeTreeResolver(ctx context.Context, commit *gitCommitResolver, path string, recursive bool) (*treeResolver, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if recursive && path != "" {
		return nil, errors.New("not implemented")
	}

	vcsrepo := backend.Repos.CachedVCS(commit.repo.repo)
	entries, err := vcsrepo.ReadDir(ctx, api.CommitID(commit.oid), path, recursive)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") { // TODO proper error value
			// empty tree is not an error
		} else {
			return nil, err
		}
	}

	return &treeResolver{
		commit:  commit,
		path:    path,
		entries: entries,
	}, nil
}

func (r *treeResolver) toFileResolvers(filter func(fi os.FileInfo) bool, alloc int) []*fileResolver {
	var prefix string
	if r.path != "" {
		prefix = r.path + "/"
	}

	l := make([]*fileResolver, 0, alloc)
	for _, entry := range r.entries {
		if filter == nil || filter(entry) {
			l = append(l, &fileResolver{
				commit: r.commit,
				path:   prefix + entry.Name(), // relies on git paths being cleaned already
				stat:   entry,
			})
		}
	}
	return l
}

func (r *treeResolver) Entries() []*fileResolver {
	return r.toFileResolvers(nil, len(r.entries))
}

func (r *treeResolver) Directories() []*fileResolver {
	return r.toFileResolvers(func(fi os.FileInfo) bool {
		return fi.Mode().IsDir()
	}, len(r.entries)/8) // heuristic: 1/8 of the entries in a repo are dirs
}

func (r *treeResolver) Files() []*fileResolver {
	return r.toFileResolvers(func(fi os.FileInfo) bool {
		return !fi.Mode().IsDir()
	}, len(r.entries))
}

// InternalRaw returns the raw tree encoded for consumption by the frontend tree component.
//
// The structure is tightly coupled to the frontend tree component's implementation, for maximum
// speed.
func (r *treeResolver) InternalRaw() string {
	var prefix string
	if r.path != "" {
		prefix = r.path + "/"
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(r.entries)*20))
	for i, entry := range r.entries {
		if entry.IsDir() {
			continue
		}
		buf.WriteString(prefix + entry.Name())
		if i != len(r.entries)-1 {
			buf.WriteByte('\x00')
		}
	}
	return buf.String()
}

func (r *fileResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, r.path, false)
}
