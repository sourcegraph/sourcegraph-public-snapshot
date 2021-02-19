package graphqlbackend

import (
	"context"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *GitTreeEntryResolver) IsRoot() bool {
	path := path.Clean(r.Path())
	return path == "/" || path == "." || path == ""
}

type gitTreeEntryConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Recursive bool
	// If recurseSingleChild is true, we will return a flat list of every
	// directory and file in a single-child nest.
	RecursiveSingleChild bool
}

func (r *GitTreeEntryResolver) Entries(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, nil)
}

func (r *GitTreeEntryResolver) Directories(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi os.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) Files(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi os.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) entries(ctx context.Context, args *gitTreeEntryConnectionArgs, filter func(fi os.FileInfo) bool) ([]*GitTreeEntryResolver, error) {
	entries, err := git.ReadDir(
		ctx,
		r.commit.repoResolver.name,
		api.CommitID(r.commit.OID()),
		r.Path(),
		r.isRecursive || args.Recursive,
	)
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

	hasSingleChild := len(entries) == 1
	var l []*GitTreeEntryResolver
	for _, entry := range entries {
		if filter == nil || filter(entry) {
			l = append(l, &GitTreeEntryResolver{
				db:            r.db,
				commit:        r.commit,
				stat:          entry,
				isSingleChild: &hasSingleChild,
			})
		}
	}

	if !args.Recursive && args.RecursiveSingleChild && len(l) == 1 {
		subEntries, err := l[0].entries(ctx, args, filter)
		if err != nil {
			return nil, err
		}
		l = append(l, subEntries...)
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
