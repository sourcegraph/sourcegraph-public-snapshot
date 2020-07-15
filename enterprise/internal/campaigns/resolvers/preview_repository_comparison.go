package resolvers

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func NewPreviewRepositoryComparisonResolver(ctx context.Context, repo *graphqlbackend.RepositoryResolver, baseRev, patch string) (*previewRepositoryComparisonResolver, error) {
	args := &graphqlbackend.RepositoryCommitArgs{Rev: baseRev}
	commit, err := repo.Commit(ctx, args)
	if err != nil {
		return nil, err
	}
	return &previewRepositoryComparisonResolver{
		repo:   repo,
		commit: commit,
		patch:  patch,
	}, nil
}

type previewRepositoryComparisonResolver struct {
	repo   *graphqlbackend.RepositoryResolver
	commit *graphqlbackend.GitCommitResolver
	patch  string
}

func (r *previewRepositoryComparisonResolver) ToPreviewRepositoryComparison() (graphqlbackend.PreviewRepositoryComparisonResolver, bool) {
	return r, true
}

func (r *previewRepositoryComparisonResolver) ToRepositoryComparison() (*graphqlbackend.RepositoryComparisonResolver, bool) {
	return nil, false
}

func (r *previewRepositoryComparisonResolver) BaseRepository() *graphqlbackend.RepositoryResolver {
	return r.repo
}

func (r *previewRepositoryComparisonResolver) FileDiffs(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) (graphqlbackend.FileDiffConnection, error) {
	return graphqlbackend.NewFileDiffConnectionResolver(r.commit, r.commit, args, fileDiffConnectionCompute(r.patch), previewNewFile), nil
}

func fileDiffConnectionCompute(patch string) func(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
	var (
		once        sync.Once
		fileDiffs   []*diff.FileDiff
		afterIdx    int32
		hasNextPage bool
		err         error
	)
	return func(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
		once.Do(func() {
			if args.After != nil {
				parsedIdx, err := strconv.ParseInt(*args.After, 0, 32)
				if err != nil {
					return
				}
				if parsedIdx < 0 {
					parsedIdx = 0
				}
				afterIdx = int32(parsedIdx)
			}
			totalAmount := afterIdx
			if args.First != nil {
				totalAmount += *args.First
			}

			dr := diff.NewMultiFileDiffReader(strings.NewReader(patch))
			for {
				var fileDiff *diff.FileDiff
				fileDiff, err = dr.ReadFile()
				if err == io.EOF {
					err = nil
					break
				}
				if err != nil {
					return
				}
				fileDiffs = append(fileDiffs, fileDiff)
				if len(fileDiffs) == int(totalAmount) {
					// Check for hasNextPage.
					_, err = dr.ReadFile()
					if err != nil && err != io.EOF {
						return
					}
					if err == io.EOF {
						err = nil
					} else {
						hasNextPage = true
					}
					break
				}
			}
		})
		return fileDiffs, afterIdx, hasNextPage, err
	}
}

func previewNewFile(r *graphqlbackend.FileDiffResolver) graphqlbackend.FileResolver {
	fileStat := graphqlbackend.CreateFileInfo(r.FileDiff.NewName, false)
	return graphqlbackend.NewVirtualFileResolver(fileStat, fileDiffVirtualFileContent(r))
}

func fileDiffVirtualFileContent(r *graphqlbackend.FileDiffResolver) graphqlbackend.FileContentFunc {
	var (
		once       sync.Once
		newContent string
		err        error
	)
	return func(ctx context.Context) (string, error) {
		once.Do(func() {
			var oldContent string
			if oldFile := r.OldFile(); oldFile != nil {
				var err error
				oldContent, err = r.OldFile().Content(ctx)
				if err != nil {
					return
				}
			}
			newContent = applyPatch(oldContent, r.FileDiff)
		})
		return newContent, err
	}
}
