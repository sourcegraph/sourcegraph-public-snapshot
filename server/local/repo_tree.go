package local

import (
	"math"
	"strings"

	"code.google.com/p/rog-go/parallel"
	"github.com/cznic/mathutil"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var RepoTree sourcegraph.RepoTreeServer = &repoTree{}

type repoTree struct{}

var _ sourcegraph.RepoTreeServer = (*repoTree)(nil)

func (s *repoTree) Get(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
	entrySpec := op.Entry
	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.RepoTreeGetOptions{}
	}

	cacheOnCommitID(ctx, entrySpec.RepoRev.CommitID)

	// It's OK if entrySpec is a dir. GetFileOptions will be ignored
	// by the vcsstore server in that case.
	entry0, err := s.getFromVCS(ctx, entrySpec, &opt.GetFileOptions)
	if err != nil {
		return nil, err
	}

	entry := &sourcegraph.TreeEntry{
		TreeEntry: entry0.TreeEntry,
	}
	if entry0.Type == vcsclient.FileEntry {
		entry.FileRange = &entry0.FileRange
	}

	switch {
	case opt.TokenizedSource && opt.Formatted:
		return nil, &sourcegraph.InvalidOptionsError{Reason: "at most one of TokenizedSource and Formatted may be specified"}

	case opt.TokenizedSource:
		sourceCode, err := sourcecode.Parse(ctx, entrySpec, entry0)
		if err == nil {
			entry.Contents = nil
			entry.SourceCode = sourceCode
		}
		if err != nil && err != sourcecode.ErrIsNotFile {
			return nil, err
		}

	case opt.Formatted:
		res, err := sourcecode.Format(ctx, entrySpec, entry0, opt.HighlightStrings)
		if err == nil {
			entry.FormatResult = res
		}
		if err != nil && err != sourcecode.ErrIsNotFile {
			return nil, err
		}
	}

	if opt.ContentsAsString {
		entry.ContentsString = string(entry.Contents)
		entry.Contents = nil
	}

	return entry, nil
}

// getFromVCS gets a tree entry from the vcsstore. Even though the
// return type is FileWithRange, it can return dirs too (the FileRange
// embedded struct will just be zeroed).
func (s *repoTree) getFromVCS(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, opt *vcsclient.GetFileOptions) (*vcsclient.FileWithRange, error) {
	if opt == nil {
		opt = &vcsclient.GetFileOptions{}
	}

	if err := (&repos{}).resolveRepoRev(ctx, &entrySpec.RepoRev); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, entrySpec.RepoRev.URI)
	if err != nil {
		return nil, err
	}

	fs, err := vcsrepo.FileSystem(vcs.CommitID(entrySpec.RepoRev.CommitID))
	if err != nil {
		return nil, err
	}

	return vcsclient.GetFileWithOptions(fs, entrySpec.Path, *opt)
}

func (s *repoTree) List(ctx context.Context, op *sourcegraph.RepoTreeListOp) (*sourcegraph.RepoTreeListResult, error) {
	repoRevSpec := op.Rev

	cacheOnCommitID(ctx, repoRevSpec.CommitID)

	if err := (&repos{}).resolveRepoRev(ctx, &repoRevSpec); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRevSpec.URI)
	if err != nil {
		return nil, err
	}

	if flvcs, ok := vcsrepo.(vcs.FileLister); ok {
		files, err := flvcs.ListFiles(vcs.CommitID(repoRevSpec.CommitID))
		if err != nil {
			return nil, err
		}
		return &sourcegraph.RepoTreeListResult{Files: files}, nil
	} else {
		return nil, grpc.Errorf(codes.Unimplemented, "repo does not support listing files")
	}
}

func (s *repoTree) Search(ctx context.Context, op *sourcegraph.RepoTreeSearchOp) (*sourcegraph.VCSSearchResultList, error) {
	repoRev := op.Rev
	opt := op.Opt
	if opt == nil || strings.TrimSpace(opt.Query) == "" {
		return nil, &sourcegraph.InvalidOptionsError{Reason: "opt and opt.Query must be set"}
	}

	cacheOnCommitID(ctx, repoRev.CommitID)

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}

	rcs, ok := vcsrepo.(vcs.Searcher)
	if !ok {
		// Repo does not support tree searching.
		return nil, &sourcegraph.NotImplementedError{What: "VCS searching"}
	}

	if repoRev.CommitID == "" {
		buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRev})
		if err == nil && buildInfo.LastSuccessful != nil {
			repoRev.CommitID = buildInfo.LastSuccessful.CommitID
		} else {
			// Fall back to textual result (no build, so no formatting).
			if err := (&repos{}).resolveRepoRev(ctx, &repoRev); err != nil {
				return nil, err
			}
		}
	}
	if repoRev.Rev == "" {
		repoRev.Rev = repoRev.CommitID
	}

	origN, origOffset := opt.SearchOptions.N, opt.SearchOptions.Offset
	// Get all of the matches in the repo so we can count the total.
	opt.SearchOptions.N, opt.SearchOptions.Offset = math.MaxInt32, 0
	res, err := rcs.Search(vcs.CommitID(repoRev.CommitID), opt.SearchOptions)
	if err != nil {
		return nil, err
	}

	total := len(res)
	// Paginate the results.
	res = res[origOffset:mathutil.Min(int(origOffset+origN), total)]

	if opt.Formatted {
		// Format the results in parallel since each call to RepoTree.Get is expensive
		// (on the order of ~30ms per call) due to blocking I/O constraints.
		par := parallel.NewRun(8)
		for _, res := range res {
			r := res
			par.Do(func() error {
				entrySpec := sourcegraph.TreeEntrySpec{RepoRev: repoRev, Path: r.File}
				f, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{
					Entry: entrySpec,
					Opt: &sourcegraph.RepoTreeGetOptions{
						Formatted:        true,
						HighlightStrings: []string{opt.SearchOptions.Query},
						GetFileOptions: vcsclient.GetFileOptions{
							FileRange: vcsclient.FileRange{
								StartLine: int64(r.StartLine),
								EndLine:   int64(r.EndLine),
							},
						},
					},
				})

				r.Match = f.Contents
				if err != nil {
					return err
				}
				return nil
			})
		}

		err = par.Wait()
		if err != nil {
			return nil, err
		}
	}

	return &sourcegraph.VCSSearchResultList{
		SearchResults: res,
		ListResponse: sourcegraph.ListResponse{
			Total: int32(total),
		},
	}, nil
}
