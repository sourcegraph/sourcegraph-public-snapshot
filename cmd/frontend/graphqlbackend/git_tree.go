package graphqlbackend

import (
	"context"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func (r *GitTreeEntryResolver) IsRoot() bool {
	cleanPath := path.Clean(r.Path())
	return cleanPath == "/" || cleanPath == "." || cleanPath == ""
}

type gitTreeEntryConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Recursive bool
	// If recurseSingleChild is true, we will return a flat list of every
	// directory and file in a single-child nest.
	RecursiveSingleChild bool
	// If Ancestors is true and the tree is loaded from a subdirectory, we will
	// return a flat list of all entries in all parent directories.
	Ancestors bool
}

func (r *GitTreeEntryResolver) Entries(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, nil)
}

func (r *GitTreeEntryResolver) Directories(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) Files(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) entries(ctx context.Context, args *gitTreeEntryConnectionArgs, filter func(fi fs.FileInfo) bool) (_ []*GitTreeEntryResolver, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "GitTreeEntryResolver", "entries")
	defer tr.FinishWithErr(&err)

	entries, err := r.gitserverClient.ReadDir(ctx, authz.DefaultSubRepoPermsChecker, r.commit.repoResolver.RepoName(), api.CommitID(r.commit.OID()), r.Path(), r.isRecursive || args.Recursive)
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

	l := make([]*GitTreeEntryResolver, 0, len(entries))
	for _, entry := range entries {
		// Apply any additional filtering

		if filter == nil || filter(entry) {
			opts := GitTreeEntryResolverOpts{
				Commit: r.Commit(),
				Stat:   entry,
			}
			l = append(l, NewGitTreeEntryResolver(r.db, r.gitserverClient, opts))
		}
	}

	// Update endLine filtering
	hasSingleChild := len(l) == 1
	for i := range l {
		l[i].isSingleChild = &hasSingleChild
	}

	if !args.Recursive && args.RecursiveSingleChild && len(l) == 1 {
		subEntries, err := l[0].entries(ctx, args, filter)
		if err != nil {
			return nil, err
		}
		l = append(l, subEntries...)
	}

	if args.Ancestors && !r.IsRoot() {
		var parent *GitTreeEntryResolver
		parent, err = r.parent(ctx)
		if err != nil {
			return nil, err
		}
		if parent != nil {
			parentEntries, err := parent.Entries(ctx, args)
			if err != nil {
				return nil, err
			}
			l = append(parentEntries, l...)
		}
	}

	return l, nil
}

type byDirectory []fs.FileInfo

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
