package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/groupcache/lru"
	"github.com/ryanuber/go-glob"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-diff/diff"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/synclru"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var deltasListFilesCache = synclru.New(lru.New(50))

func (s *deltas) ListFiles(ctx context.Context, op *sourcegraph.DeltasListFilesOp) (*sourcegraph.DeltaFiles, error) {
	ds := op.Ds
	opt := op.Opt

	// SECURITY NOTE: If these auth checks are moved or removed, we
	// MUST remove the code below that satisfies this request from the
	// cache, since we can't be sure that the user is authorized to
	// view the result.
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.ListFiles", ds.Base.URI); err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.ListFiles", ds.Head.URI); err != nil {
		return nil, err
	}

	// Make sure we've fully resolved the RepoRevSpecs. If we haven't,
	// then they will need to be re-resolved in each call to
	// RepoTree.Get that we issue, which will seriously degrade
	// performance.
	resolveAndCacheRepoRevAndBranchExistence := func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) error {
		if repoRev.Resolved() {
			// The repo rev appears resolved already -- but it might have been
			// deleted, thus making any URLs we would emit for Rev instead of CommitID
			// invalid. Check if the rev/branch was deleted:
			//
			// TODO(slimsag): write a test exactly for this case.
			unresolvedRev := *repoRev
			unresolvedRev.CommitID = ""
			if err := (&repos{}).resolveRepoRev(ctx, &unresolvedRev); errcode.GRPC(err) == codes.NotFound {
				// Rev no longer exists, so fallback to the CommitID instead. This is a
				// last-ditch effort to ensure tokenized source displays well in diffs
				// that are very old / have had one or more of their revs/branches
				// deleted.
				repoRev.Rev = repoRev.CommitID
			} else if err != nil {
				return err
			}
		}
		return (&repos{}).resolveRepoRev(ctx, repoRev)
	}
	if err := resolveAndCacheRepoRevAndBranchExistence(ctx, &ds.Base); err != nil {
		return nil, err
	}
	if err := resolveAndCacheRepoRevAndBranchExistence(ctx, &ds.Head); err != nil {
		return nil, err
	}

	if opt == nil {
		opt = &sourcegraph.DeltaListFilesOptions{}
	}

	// Construct cache key and check if cached.
	//
	// SECURITY NOTE: We are only able to cache these because we've
	// checked the user's authentication above. If those checks are
	// removed, we can't return the cached values without leaking
	// private data!
	op.Ds = ds
	op.Opt = opt
	cacheKey, err := json.Marshal(op)
	if err != nil {
		return nil, err
	}
	if res, found := deltasListFilesCache.Get(string(cacheKey)); found {
		return res.(*sourcegraph.DeltaFiles), nil
	}

	fdiffs, delta, err := s.diff(ctx, ds)
	if err != nil {
		return nil, err
	}

	filtered := make(map[*diff.FileDiff]bool)
	if opt.Filter != "" {
		filter := opt.Filter
		expected := true
		if filter[0] == '!' {
			filter = filter[1:]
			expected = false
		}
		for _, fdiff := range fdiffs {
			if (strings.HasPrefix(fdiff.OrigName, filter) || strings.HasPrefix(fdiff.NewName, filter)) != expected {
				filtered[fdiff] = true
			}
		}
	} else if len(opt.Ignore) > 0 {
		for _, fdiff := range fdiffs {
			for _, pattern := range opt.Ignore {
				if glob.Glob(pattern, fdiff.OrigName) || glob.Glob(pattern, fdiff.NewName) {
					filtered[fdiff] = true
				}
			}
		}
	}

	files, err := parseMultiFileDiffs(ctx, delta, fdiffs, filtered, opt)
	if err != nil {
		return nil, err
	}

	deltasListFilesCache.Add(string(cacheKey), files)
	return files, nil
}

func (s *deltas) diff(ctx context.Context, ds sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error) {
	if s.mockDiffFunc != nil {
		return s.mockDiffFunc(ctx, ds)
	}

	delta, err := s.Get(ctx, &ds)
	if err != nil {
		return nil, nil, err
	}
	ds = delta.DeltaSpec()

	baseVCSRepo, err := store.RepoVCSFromContext(ctx).Open(ctx, delta.BaseRepo.URI)
	if err != nil {
		return nil, nil, err
	}

	if ds.Base.RepoSpec != ds.Head.RepoSpec {
		return nil, nil, errors.New("base and head repo must be identical")
	}

	var vcsDiff *vcs.Diff
	diffOpt := &vcs.DiffOptions{
		DetectRenames: true,
		OrigPrefix:    "",
		NewPrefix:     "",

		// We want `git diff base...head` not `git diff base..head` or
		// else branches with base merge commits show diffs that
		// include those merges, which isn't what we want (since those
		// merge commits are already reflected in the base).
		ExcludeReachableFromBoth: true,
	}

	vcsDiff, err = baseVCSRepo.Diff(vcs.CommitID(ds.Base.CommitID), vcs.CommitID(ds.Head.CommitID), diffOpt)
	if err != nil {
		return nil, nil, err
	}

	fdiffs, err := diff.ParseMultiFileDiff([]byte(vcsDiff.Raw))
	if err != nil {
		return nil, nil, err
	}
	return fdiffs, delta, nil
}

func parseMultiFileDiffs(ctx context.Context, delta *sourcegraph.Delta, fdiffs []*diff.FileDiff, filtered map[*diff.FileDiff]bool, opt *sourcegraph.DeltaListFilesOptions) (*sourcegraph.DeltaFiles, error) {
	fds := make([]*sourcegraph.FileDiff, len(fdiffs))
	for i, fd := range fdiffs {
		parseRenames(fd)
		pre, post := getPrePostImage(fd.Extended)
		fds[i] = &sourcegraph.FileDiff{
			FileDiff:  *fd,
			Stats:     fd.Stat(),
			PreImage:  pre,
			PostImage: post,
		}
		if _, filtered := filtered[fd]; filtered {
			fds[i].FileDiff.Hunks = nil
			fds[i].Filtered = true
			continue
		}
	}
	files := &sourcegraph.DeltaFiles{
		FileDiffs: fds,
		Delta:     delta,
	}
	files.Stats = files.DiffStat()
	return files, nil
}

// parseRenames checks if this file diff is barely a rename and updates
// it's OrigName and NewName values accordingly from extended headers
// "rename from <path>" and "rename to <path>" if available.
// This only occurs on renames with similarity index at 100% which contain
// no hunks.
func parseRenames(fd *diff.FileDiff) {
	if fd.Hunks != nil || fd.OrigName != "" {
		// this is not a rename
		return
	}
	var prefixFrom = "rename from "
	var prefixTo = "rename to "
	for _, h := range fd.Extended {
		if strings.HasPrefix(h, prefixFrom) {
			fd.OrigName = h[len(prefixFrom):]
			continue
		}
		if strings.HasPrefix(h, prefixTo) {
			fd.NewName = h[len(prefixTo):]
			break
		}
	}
}

// getPrePostImage searches for a diff's index header inside a list
// of headers and if found, returns the pre and post commit ID or
// empty strings.
func getPrePostImage(headers []string) (pre, post string) {
	for _, h := range headers {
		if strings.HasPrefix(h, "index") {
			n, err := fmt.Sscanf(h, "index %40s..%40s", &pre, &post)
			if n == 2 && err == nil {
				if pre == strings.Repeat("0", 40) {
					pre = ""
				}
				if post == strings.Repeat("0", 40) {
					post = ""
				}
				return
			}
			break
		}
	}
	return "", ""
}
