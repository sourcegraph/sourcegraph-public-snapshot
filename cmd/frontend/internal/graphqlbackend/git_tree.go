package graphqlbackend

import (
	"context"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (r *gitTreeEntryResolver) IsRoot() bool {
	path := path.Clean(r.path)
	return path == "/" || path == "." || path == ""
}

type gitTreeEntryConnectionArgs struct {
	connectionArgs
	Recursive bool
}

func (r *gitTreeEntryResolver) Entries(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*gitTreeEntryResolver, error) {
	return r.entries(ctx, args, nil)
}

func (r *gitTreeEntryResolver) Directories(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*gitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi os.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *gitTreeEntryResolver) Files(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*gitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi os.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *gitTreeEntryResolver) entries(ctx context.Context, args *gitTreeEntryConnectionArgs, filter func(fi os.FileInfo) bool) ([]*gitTreeEntryResolver, error) {
	entries, err := git.ReadDir(ctx, backend.CachedGitRepo(r.commit.repo.repo), api.CommitID(r.commit.oid), r.path, r.isRecursive || args.Recursive)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") { // TODO proper error value
			// empty tree is not an error
		} else {
			return nil, err
		}
	}

	sort.Sort(byDirectory(entries))

	if args.First != nil && len(entries) > int(*args.First) {
		entries = entries[:int(*args.First)]
	}

	var prefix string
	if r.path != "" {
		prefix = r.path + "/"
	}

	var l []*gitTreeEntryResolver
	for _, entry := range entries {
		if filter == nil || filter(entry) {
			l = append(l, &gitTreeEntryResolver{
				commit: r.commit,
				path:   prefix + entry.Name(), // relies on git paths being cleaned already
				stat:   entry,
			})
		}
	}
	return l, nil
}

type byDirectory []os.FileInfo

func (s byDirectory) Len() int {
	return len(s)
}

func (s byDirectory) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byDirectory) Less(i, j int) bool {
	if s[i].IsDir() && !s[j].IsDir() {
		return true
	}

	if !s[i].IsDir() && s[j].IsDir() {
		return false
	}

	return s[i].Name() < s[j].Name()
}
