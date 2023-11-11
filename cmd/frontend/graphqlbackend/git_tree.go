package graphqlbackend

import (
	"context"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *GitTreeEntryResolver) IsRoot() bool {
	cleanPath := path.Clean(r.Path())
	return cleanPath == "/" || cleanPath == "." || cleanPath == ""
}

type gitTreeEntryConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Recursive bool
	// If recurseSingleChild is true, and the requested path only has a single child
	// that is a directory, we recurse into that directory.
	// Only respected when Recursive is false.
	// Example:
	// FS structure:
	// /folder/subfolder/file
	// Request for entries of /folder
	// We don't return only /folder/subfolder, and instead also travserse into subfolder
	// until we hit something with more than 1 child that's not a dir.
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
	tr, ctx := trace.New(ctx, "GitTreeEntryResolver.entries")
	defer tr.EndWithErr(&err)

	if args.First != nil && *args.First < 0 {
		return nil, errors.Newf("invalid argument for first, must be non-negative")
	}

	if args.Recursive && args.RecursiveSingleChild {
		// No extra work needed, recursive includes all these.
		args.RecursiveSingleChild = false
	}

	// If RecursiveSingleChild is true, we also get all files recursively. Otherwise, we would
	// have to do a readdir for every single directory to see if it has only one child (and nested)
	// dirs, too.
	entries, err := r.gitserverClient.ReadDir(ctx, r.commit.repoResolver.RepoName(), api.CommitID(r.commit.OID()), r.Path(), args.Recursive)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") { // TODO proper error value
			// empty tree is not an error
		} else {
			return nil, err
		}
	}

	// When using recursive: true on gitserverClient.ReadDir, we get entries for
	// all parent trees (directories) from git as well, so we filter those out.
	// Ideally, we fix this in the ReadDir API, but this might have other unforseen
	// side-effects so we will revisit that later.
	// Example output from git for ls-tree cmd/gitserver with -r -t (recursive: true):
	// cmd
	// gitserver
	// [...] files in cmd/gitserver and deeper.
	// To drop those, we just have to drop as many entries as the level of nesting
	// r.Path is at.
	if args.Recursive && !r.IsRoot() {
		entries = entries[len(strings.Split(strings.Trim(r.Path(), "/"), "/")):]
	}

	if !args.Recursive && args.RecursiveSingleChild && len(entries) == 1 && entries[0].IsDir() {
		opts := GitTreeEntryResolverOpts{
			Commit: r.Commit(),
			Stat:   entries[0],
		}
		r := NewGitTreeEntryResolver(r.db, r.gitserverClient, opts)

		subEntries, err := r.entries(ctx, &gitTreeEntryConnectionArgs{
			RecursiveSingleChild: true,
		}, nil)
		if err != nil {
			return nil, err
		}
		for _, e := range subEntries {
			entries = append(entries, e.stat)
		}
	}

	maxResolvers := len(entries)
	if args.First != nil && int(*args.First) < maxResolvers {
		maxResolvers = int(*args.First)
	}

	sort.Sort(byDirectory(entries))

	resolvers := make([]*GitTreeEntryResolver, 0, maxResolvers)
	seen := 0
	for _, entry := range entries {
		if seen == maxResolvers {
			break
		}

		// Apply any additional filtering
		if filter == nil || filter(entry) {
			seen++
			opts := GitTreeEntryResolverOpts{
				Commit: r.Commit(),
				Stat:   entry,
			}
			resolvers = append(resolvers, NewGitTreeEntryResolver(r.db, r.gitserverClient, opts))
		}
	}

	if args.Ancestors && !r.IsRoot() {
		p := r.Path()
		p = strings.Trim(p, "/")
		parent := NewGitTreeEntryResolver(r.db, r.gitserverClient, GitTreeEntryResolverOpts{
			Commit: r.commit,
			Stat:   CreateFileInfo(filepath.Dir(p), true),
		})
		parentEntries, err := parent.entries(ctx, &gitTreeEntryConnectionArgs{
			Ancestors: true,
		}, nil)
		if err != nil {
			return nil, err
		}
		resolvers = append(parentEntries, resolvers...)
	}

	return resolvers, nil
}

// byDirectory implements sort.Sortable and orders directories before files.
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
