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
	// If recurseSingleChild is true, we will return a flat list of every
	// directory and file in a single-child nest.
	// Only respected when Recursive is false.
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
	entries, err := r.gitserverClient.ReadDir(ctx, r.commit.repoResolver.RepoName(), api.CommitID(r.commit.OID()), r.Path(), args.Recursive || args.RecursiveSingleChild)
	if err != nil {
		if strings.Contains(err.Error(), "file does not exist") { // TODO proper error value
			// empty tree is not an error
		} else {
			return nil, err
		}
	}

	// If RecursiveSingleChild is true, we need to filter out non-single-childs
	// again.
	filteredEntries := entries
	if args.RecursiveSingleChild {
		// Reset filteredEntries.
		filteredEntries = filteredEntries[:0]
		// Convert the entries into a tree fs.
		fs := entriesToFsTree(r.stat, entries)
		// Keep all files in the current directory.
		for _, entry := range entries {
			if normalizePath(filepath.Dir(normalizePath(entry.Name()))) == normalizePath(r.Path()) {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		// And now traverse all the children and check if there's at most 1 item
		// in it.
		var traverseFs func(*fsNode, int)
		traverseFs = func(fs *fsNode, l int) {
			if fs.IsDir() && len(fs.children) == 1 {
				if l != 0 {
					filteredEntries = append(filteredEntries, fs.node)
				}
				traverseFs(fs.children[0], l+1)
				return
			} else {
				if !fs.IsDir() {
					if l != 0 {
						filteredEntries = append(filteredEntries, fs.node)
					}
				}
			}
		}
		for _, c := range fs.children {
			traverseFs(c, 0)
		}
	}

	maxResolvers := len(filteredEntries)
	if args.First != nil && int(*args.First) < maxResolvers {
		maxResolvers = int(*args.First)
	}

	sort.Sort(byDirectory(filteredEntries))

	resolvers := make([]*GitTreeEntryResolver, 0, maxResolvers)
	seen := 0
	for _, entry := range filteredEntries {
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

type fsNode struct {
	name     string
	node     fs.FileInfo
	children []*fsNode
}

func (n *fsNode) IsDir() bool {
	return n.node.IsDir()
}

// entriesToFsTree takes a slice of fs.FileInfo entries and converts them into a
// tree structure like a file system. This makes it easy to traverse the tree
// for what would otherwise be a simple slice of strings basically.
func entriesToFsTree(root fs.FileInfo, entries []fs.FileInfo) *fsNode {
	// Make sure we order by length, this means that dirs are always inserted before files.
	sort.Slice(entries, func(i, j int) bool {
		return len(entries[i].Name()) < len(entries[j].Name())
	})
	normRoot := normalizePath(root.Name())
	tree := &fsNode{name: normRoot, node: root}
	for _, entry := range entries {
		path := normalizePath(entry.Name())
		path = strings.TrimPrefix(path, normRoot)
		segments := strings.Split(path, "/")
		curTree := tree
		for i, s := range segments {
			if i == len(segments)-1 {
				node := &fsNode{name: s, node: entry}
				curTree.children = append(curTree.children, node)
			} else {
				for _, c := range curTree.children {
					if c.name == s {
						curTree = c
						break
					}
				}
			}
		}
	}
	return tree
}

func normalizePath(path string) string {
	if path == "." || path == "/" {
		path = ""
	}
	return strings.Trim(path, "/")
}
