package graphqlbackend

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
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
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) Files(ctx context.Context, args *gitTreeEntryConnectionArgs) ([]*GitTreeEntryResolver, error) {
	return r.entries(ctx, args, func(fi fs.FileInfo) bool { return !fi.Mode().IsDir() })
}

func (r *GitTreeEntryResolver) entries(ctx context.Context, args *gitTreeEntryConnectionArgs, filter func(fi fs.FileInfo) bool) ([]*GitTreeEntryResolver, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "tree.entries")
	defer span.Finish()

	// First check if we are able to view the tree at all, if not we can return early
	// and don't need to hit gitserver.
	perms, err := authz.CurrentUserPermissions(ctx, r.subRepoPerms, authz.RepoContent{
		Repo: r.commit.repoResolver.RepoName(),
		Path: r.Path(),
	})
	if err != nil {
		log15.Error("checking sub-repo permissions", "error", err)
		return nil, err
	}
	// No access
	if !perms.Include(authz.Read) {
		return nil, nil
	}

	entries, err := gitReadDir(
		ctx,
		r.subRepoPerms,
		r.commit.repoResolver.RepoName(),
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

	l := make([]*GitTreeEntryResolver, 0, len(entries))
	for _, entry := range entries {
		// Apply any additional filtering
		if filter == nil || filter(entry) {
			l = append(l, NewGitTreeEntryResolver(r.db, r.commit, entry))
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

// gitReadDir call git.ReadDir but applies sub-repo filtering to the returned entries.
func gitReadDir(ctx context.Context, srp authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	entries, err := git.ReadDir(ctx, repo, commit, path, recurse)
	if err != nil {
		return nil, err
	}

	if !srp.Enabled() {
		return entries, nil
	}

	// Filter in place, keeping entries the given actor is authorized to see.
	tr, ctx := trace.New(ctx, "gitReadDir.subRepoPerms", "")
	var (
		a            = actor.FromContext(ctx)
		errs         = &multierror.Error{}
		authorized   = 0
		resultsCount = len(entries)
	)
	defer func() {
		tr.SetError(errs.ErrorOrNil())
		tr.LazyPrintf("actor=(%s) authorized=%d unauthorized=%d",
			a.String(), authorized, resultsCount-authorized)
		tr.Finish()
	}()
	for _, entry := range entries {
		perms, err := authz.ActorPermissions(ctx, srp, a, authz.RepoContent{
			Repo: repo,
			Path: entry.Name(),
		})
		if err != nil {
			// Log error but don't propagate upwards to ensure data does not leak
			log15.Error("gitReadDir.subRepoPerms check failed",
				"actor.UID", a.UID,
				"entry.Name", entry.Name(),
				"error", err)
			errs = multierror.Append(errs, fmt.Errorf("subRepoPermsFilter: failed to check sub-repo permissions"))
		}
		if perms.Include(authz.Read) {
			entries[authorized] = entry
			authorized++
		}
	}
	// Only keep authorized matches
	entries = entries[:authorized]

	return entries, nil
}
