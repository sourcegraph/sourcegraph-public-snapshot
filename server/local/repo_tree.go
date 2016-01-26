package local

import (
	"math"
	"sort"
	"strings"

	"github.com/cznic/mathutil"
	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

var RepoTree sourcegraph.RepoTreeServer = &repoTree{}

type repoTree struct{}

var _ sourcegraph.RepoTreeServer = (*repoTree)(nil)

func (s *repoTree) Get(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoTree.Get", op.Entry.RepoRev.URI); err != nil {
		return nil, err
	}

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
		BasicTreeEntry: entry0.BasicTreeEntry,
	}
	if entry0.Type == sourcegraph.FileEntry {
		entry.FileRange = &entry0.FileRange
	}

	switch {
	case opt.TokenizedSource && opt.Formatted:
		return nil, grpc.Errorf(codes.InvalidArgument, "at most one of TokenizedSource and Formatted may be specified")

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
func (s *repoTree) getFromVCS(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, opt *sourcegraph.GetFileOptions) (*sourcegraph.FileWithRange, error) {
	if opt == nil {
		opt = &sourcegraph.GetFileOptions{}
	}

	if err := (&repos{}).resolveRepoRev(ctx, &entrySpec.RepoRev); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, entrySpec.RepoRev.URI)
	if err != nil {
		return nil, err
	}

	commit := vcs.CommitID(entrySpec.RepoRev.CommitID)

	fi, err := vcsrepo.Lstat(commit, entrySpec.Path)
	if err != nil {
		return nil, err
	}

	e := newTreeEntry(fi)
	fwr := sourcegraph.FileWithRange{BasicTreeEntry: e}

	if fi.Mode().IsDir() {
		ee, err := readDir(vcsrepo, commit, entrySpec.Path, int(opt.RecurseSingleSubfolderLimit), true)
		if err != nil {
			return nil, err
		}
		sort.Sort(TreeEntriesByTypeByName(ee))
		e.Entries = ee
	} else if fi.Mode().IsRegular() {
		contents, err := vcsrepo.ReadFile(commit, entrySpec.Path)
		if err != nil {
			return nil, err
		}

		e.Contents = contents

		if empty := (sourcegraph.GetFileOptions{}); *opt != empty {
			fr, _, err := computeFileRange(contents, *opt)
			if err != nil {
				return nil, err
			}

			// Trim to only requested range.
			e.Contents = e.Contents[fr.StartByte:fr.EndByte]
			fwr.FileRange = *fr
		}
	}

	return &fwr, nil
}

func (s *repoTree) List(ctx context.Context, op *sourcegraph.RepoTreeListOp) (*sourcegraph.RepoTreeListResult, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoTree.List", op.Rev.URI); err != nil {
		return nil, err
	}

	repoRevSpec := op.Rev

	cacheOnCommitID(ctx, repoRevSpec.CommitID)

	if err := (&repos{}).resolveRepoRev(ctx, &repoRevSpec); err != nil {
		return nil, err
	}

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRevSpec.URI)
	if err != nil {
		return nil, err
	}

	infos, err := vcsrepo.ReadDir(vcs.CommitID(repoRevSpec.CommitID), ".", true)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, info := range infos {
		if !info.IsDir() {
			files = append(files, info.Name())
		}
	}

	return &sourcegraph.RepoTreeListResult{Files: files}, nil
}

func (s *repoTree) Search(ctx context.Context, op *sourcegraph.RepoTreeSearchOp) (*sourcegraph.VCSSearchResultList, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoTree.List", op.Rev.URI); err != nil {
		return nil, err
	}

	repoRev := op.Rev
	opt := op.Opt
	if opt == nil || strings.TrimSpace(opt.Query) == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "opt and opt.Query must be set")
	}

	cacheOnCommitID(ctx, repoRev.CommitID)

	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoRev.URI)
	if err != nil {
		return nil, err
	}

	if !isAbsCommitID(repoRev.CommitID) {
		return nil, grpc.Errorf(codes.InvalidArgument, "absolute commit ID required (got %q)", repoRev.CommitID)
	}

	if repoRev.Rev == "" {
		repoRev.Rev = repoRev.CommitID
	}

	origN, origOffset := opt.SearchOptions.N, opt.SearchOptions.Offset
	// Get all of the matches in the repo so we can count the total.
	opt.SearchOptions.N, opt.SearchOptions.Offset = math.MaxInt32, 0
	res, err := vcsrepo.Search(vcs.CommitID(repoRev.CommitID), opt.SearchOptions)
	if err != nil {
		return nil, err
	}

	total := len(res)
	// Paginate the results.
	if int(origOffset) > total {
		return nil, grpc.Errorf(codes.InvalidArgument, "page offset bounds out of range")
	}
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
						GetFileOptions: sourcegraph.GetFileOptions{
							FileRange: sourcegraph.FileRange{
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
