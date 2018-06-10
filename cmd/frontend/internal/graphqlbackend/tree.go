package graphqlbackend

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
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

	entries, err := git.ReadDir(ctx, backend.CachedGitRepo(commit.repo.repo), api.CommitID(commit.oid), path, recursive)
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

func (r *treeResolver) toFileResolvers(filter func(fi os.FileInfo) bool, alloc int) []*gitTreeEntryResolver {
	var prefix string
	if r.path != "" {
		prefix = r.path + "/"
	}

	l := make([]*gitTreeEntryResolver, 0, alloc)
	for _, entry := range r.entries {
		if filter == nil || filter(entry) {
			l = append(l, &gitTreeEntryResolver{
				commit: r.commit,
				path:   prefix + entry.Name(), // relies on git paths being cleaned already
				stat:   entry,
			})
		}
	}
	return l
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

func (r *treeResolver) Entries(args *connectionArgs) []*gitTreeEntryResolver {
	sort.Sort(byDirectory(r.entries))
	resolvers := r.toFileResolvers(nil, len(r.entries))
	if args.First != nil && len(r.entries) > int(*args.First) {
		return resolvers[:int(*args.First)]
	}
	return resolvers
}

func (r *treeResolver) Directories(args *connectionArgs) []*gitTreeEntryResolver {
	resolvers := r.toFileResolvers(func(fi os.FileInfo) bool {
		return fi.Mode().IsDir()
	}, len(r.entries)/8) // heuristic: 1/8 of the entries in a repo are dirs
	if args.First != nil && len(resolvers) > int(*args.First) {
		return resolvers[:int(*args.First)]
	}

	return resolvers
}

func (r *treeResolver) Files(args *connectionArgs) []*gitTreeEntryResolver {
	resolvers := r.toFileResolvers(func(fi os.FileInfo) bool {
		return !fi.Mode().IsDir()
	}, len(r.entries))
	if args.First != nil && len(resolvers) > int(*args.First) {
		return resolvers[:int(*args.First)]
	}

	return resolvers
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

func (r *gitTreeEntryResolver) Tree(ctx context.Context) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, r.path, false)
}
