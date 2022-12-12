package graphqlbackend

import (
	"context"
	"io/fs"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *GitTreeEntryResolver) IsRoot() bool {
	path := path.Clean(r.Path())
	return path == "/" || path == "." || path == ""
}

type gitTreeEntrySharedArgs struct {
	Recursive bool
	// If recurseSingleChild is true, we will return a flat list of every
	// directory and file in a single-child nest.
	RecursiveSingleChild bool
}
type gitTreeEntryArgs struct {
	graphqlutil.ConnectionArgs
	gitTreeEntrySharedArgs
}

func (r *GitTreeEntryResolver) Entries(ctx context.Context, args *gitTreeEntryArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, nil)
}

type gitTreeEntryConnectionArgs struct {
	graphqlutil.ConnectionResolverArgs
	gitTreeEntrySharedArgs
}

func (r *GitTreeEntryResolver) EntriesConnection(ctx context.Context, args *gitTreeEntryConnectionArgs) (*graphqlutil.ConnectionResolver[GitTreeEntryResolver], error) {
	connectionArgs := &graphqlutil.ConnectionResolverArgs{
		First:  args.First,
		Last:   args.Last,
		After:  args.After,
		Before: args.Before,
	}
	connectionStore := &entriesConnectionStore{
		db:             r.db,
		args:           &args.gitTreeEntrySharedArgs,
		connectionArgs: connectionArgs,
		r:              r,
	}
	return graphqlutil.NewConnectionResolver[GitTreeEntryResolver](connectionStore, connectionArgs)
}

func (r *GitTreeEntryResolver) Directories(ctx context.Context, args *gitTreeEntryArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) Files(ctx context.Context, args *gitTreeEntryArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) entries(ctx context.Context, args *gitTreeEntryArgs, filter func(fi fs.FileInfo) bool) ([]*GitTreeEntryResolver, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "tree.entries")
	defer span.Finish()

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
			l = append(l, NewGitTreeEntryResolver(r.db, r.gitserverClient, r.commit, entry))
		}
	}

	// Update after filtering
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

type entriesConnectionStore struct {
	db             database.DB
	args           *gitTreeEntrySharedArgs
	connectionArgs *graphqlutil.ConnectionResolverArgs

	r *GitTreeEntryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*GitTreeEntryResolver
	err     error
}

func (s *entriesConnectionStore) MarshalCursor(node *GitTreeEntryResolver) (*string, error) {
	if s.results == nil {
		return nil, errors.New("results not initialized yet")
	}
	if node == nil {
		return nil, errors.New("node is nil")
	}
	position := -1
	for i, r := range s.results {
		if r.Path() == node.Path() {
			position = i
			break
		}
	}
	if position == -1 {
		return nil, errors.New("node not found in results")
	}
	cursor := strconv.Itoa(position)
	return &cursor, nil
}

func (s *entriesConnectionStore) UnMarshalCursor(cursor string) (*int32, error) {
	i64, err := strconv.ParseInt(cursor, 10, 32)
	if err != nil {
		return nil, err
	}
	position := int32(i64)
	return &position, nil
}

func (s *entriesConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	results, err := s.compute(ctx)
	num := int32(len(results))
	return &num, err
}

func (s *entriesConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*GitTreeEntryResolver, error) {
	results, err := s.compute(ctx)
	if err != nil {
		return nil, err
	}

	results, _, err = offsetBasedCursorSlice(results, args)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *entriesConnectionStore) compute(ctx context.Context) ([]*GitTreeEntryResolver, error) {
	s.once.Do(func() {
		args := &gitTreeEntryArgs{graphqlutil.ConnectionArgs{}, *s.args}
		s.results, s.err = s.r.entries(ctx, args, nil)
	})
	return s.results, s.err
}

func offsetBasedCursorSlice[T any](nodes []T, args *database.PaginationArgs) ([]T, int32, error) {
	start := int32(0)
	end := int32(0)
	totalFloat := float64(len(nodes))
	if args.First != nil {
		if args.After != nil {
			start = int32(math.Min(float64(*args.After)+1, totalFloat))
		}
		end = int32(math.Min(float64(start+*args.First), totalFloat))
	} else if args.Last != nil {
		end = int32(totalFloat)
		if args.Before != nil {
			end = int32(math.Max(float64(*args.Before), 0))
		}
		start = int32(math.Max(float64(end-*args.Last), 0))
	} else {
		return nil, 0, errors.New(`args.First and args.Last are nil`)
	}

	nodes = nodes[start:end]

	if args.Last != nil {
		// Invert order because the abstraction expects this for a backward pagination
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}

	return nodes, start, nil
}
