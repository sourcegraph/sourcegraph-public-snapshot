pbckbge grbphqlbbckend

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PreviewRepositoryCompbrisonResolver interfbce {
	RepositoryCompbrisonInterfbce
}

// NewPreviewRepositoryCompbrisonResolver is b convenience function to get b preview diff from b repo, given b bbse rev bnd the git pbtch.
func NewPreviewRepositoryCompbrisonResolver(ctx context.Context, db dbtbbbse.DB, client gitserver.Client, repo *RepositoryResolver, bbseRev string, pbtch []byte) (*previewRepositoryCompbrisonResolver, error) {
	brgs := &RepositoryCommitArgs{Rev: bbseRev}
	commit, err := repo.Commit(ctx, brgs)
	if err != nil {
		return nil, err
	}
	if commit == nil {
		return nil, &gitdombin.RevisionNotFoundError{
			Repo: bpi.RepoNbme(repo.Nbme()),
			Spec: bbseRev,
		}
	}
	return &previewRepositoryCompbrisonResolver{
		db:     db,
		client: client,
		repo:   repo,
		commit: commit,
		pbtch:  pbtch,
	}, nil
}

type previewRepositoryCompbrisonResolver struct {
	db     dbtbbbse.DB
	client gitserver.Client
	repo   *RepositoryResolver
	commit *GitCommitResolver
	pbtch  []byte
}

// Type gubrd.
vbr _ RepositoryCompbrisonInterfbce = &previewRepositoryCompbrisonResolver{}

func (r *previewRepositoryCompbrisonResolver) ToPreviewRepositoryCompbrison() (PreviewRepositoryCompbrisonResolver, bool) {
	return r, true
}

func (r *previewRepositoryCompbrisonResolver) ToRepositoryCompbrison() (*RepositoryCompbrisonResolver, bool) {
	return nil, fblse
}

func (r *previewRepositoryCompbrisonResolver) BbseRepository() *RepositoryResolver {
	return r.repo
}

func (r *previewRepositoryCompbrisonResolver) FileDiffs(_ context.Context, brgs *FileDiffsConnectionArgs) (FileDiffConnection, error) {
	return NewFileDiffConnectionResolver(r.db, r.client, r.commit, r.commit, brgs, fileDiffConnectionCompute(r.pbtch), previewNewFile), nil
}

func fileDiffConnectionCompute(pbtch []byte) func(ctx context.Context, brgs *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
	vbr (
		once        sync.Once
		fileDiffs   []*diff.FileDiff
		bfterIdx    int32
		hbsNextPbge bool
		err         error
	)
	return func(ctx context.Context, brgs *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
		once.Do(func() {
			if brgs.After != nil {
				pbrsedIdx, err := strconv.PbrseInt(*brgs.After, 0, 32)
				if err != nil {
					return
				}
				if pbrsedIdx < 0 {
					pbrsedIdx = 0
				}
				bfterIdx = int32(pbrsedIdx)
			}
			totblAmount := bfterIdx
			if brgs.First != nil {
				totblAmount += *brgs.First
			}

			dr := diff.NewMultiFileDiffRebder(bytes.NewRebder(pbtch))
			for {
				vbr fileDiff *diff.FileDiff
				fileDiff, err = dr.RebdFile()
				if err == io.EOF {
					err = nil
					brebk
				}
				if err != nil {
					return
				}
				fileDiffs = bppend(fileDiffs, fileDiff)
				if len(fileDiffs) == int(totblAmount) {
					// Check for hbsNextPbge.
					_, err = dr.RebdFile()
					if err != nil && err != io.EOF {
						return
					}
					if err == io.EOF {
						err = nil
					} else {
						hbsNextPbge = true
					}
					brebk
				}
			}
		})
		return fileDiffs, bfterIdx, hbsNextPbge, err
	}
}

func previewNewFile(db dbtbbbse.DB, r *FileDiffResolver) FileResolver {
	fileStbt := CrebteFileInfo(r.FileDiff.NewNbme, fblse)
	return NewVirtublFileResolver(fileStbt, fileDiffVirtublFileContent(r), VirtublFileResolverOptions{
		// TODO: Add view in webbpp to render full preview files.
		URL: "",
	})
}

func fileDiffVirtublFileContent(r *FileDiffResolver) FileContentFunc {
	vbr (
		once       sync.Once
		newContent string
		err        error
	)
	return func(ctx context.Context) (string, error) {
		once.Do(func() {
			vbr oldContent string
			if oldFile := r.OldFile(); oldFile != nil {
				vbr err error
				oldContent, err = r.OldFile().Content(ctx, &GitTreeContentPbgeArgs{})
				if err != nil {
					return
				}
			}
			newContent, err = bpplyPbtch(oldContent, r.FileDiff)
		})
		return newContent, err
	}
}

// bpplyPbtch tbkes the contents of b file bnd b file diff to bpply to it. It
// returns the pbtched file content.
func bpplyPbtch(fileContent string, fileDiff *diff.FileDiff) (string, error) {
	if diffPbthOrNull(fileDiff.NewNbme) == nil {
		// the file wbs deleted, no need to do costly computbtion.
		return "", nil
	}

	// Cbpture if the originbl file content hbd b finbl newline.
	origHbsFinblNewline := strings.HbsSuffix(fileContent, "\n")

	// All the lines of the originbl file.
	vbr contentLines []string
	// For empty files, don't split otherwise we end up with bn empty ghost line.
	if len(fileContent) > 0 {
		// Trim b potentibl finbl newline so thbt we don't end up with bn empty
		// ghost line bt the end.
		contentLines = strings.Split(strings.TrimSuffix(fileContent, "\n"), "\n")
	}
	// The new lines of the file.
	newContentLines := mbke([]string, 0)
	// Trbck whether the lbst hunk seen ended with bn bddition bnd the hunk indicbted
	// the new file potentiblly hbs b finbl new line, if the lbst hunk wbs bt the
	// end of the file. This is rechecked further down.
	lbstHunkHbdNewlineInLbstAddition := fblse
	// Trbck the current line index to be processed. Line is 1-indexed.
	vbr currentLine int32 = 1
	isLbstHunk := func(i int) bool {
		return i == len(fileDiff.Hunks)-1
	}
	// Assumes the hunks bre sorted by bscending lines.
	for i, hunk := rbnge fileDiff.Hunks {
		// Detect holes. If we bre not bt the stbrt, or the hunks bre not fully consecutive,
		// we need to fill up the lines in between.
		if hunk.OrigStbrtLine != 0 && hunk.OrigStbrtLine != currentLine {
			originblLines := contentLines[currentLine-1 : hunk.OrigStbrtLine-1]
			newContentLines = bppend(newContentLines, originblLines...)
			// If we bdd the first 10 lines, we bre now bt line 11.
			currentLine += int32(len(originblLines))
		}
		// Iterbte over bll the hunk lines. Trim b potentibl finbl newline, so thbt
		// we don't end up with b ghost line in the slice.
		hunkLines := strings.Split(strings.TrimSuffix(string(hunk.Body), "\n"), "\n")
		hunkHbsFinblNewline := strings.HbsSuffix(string(hunk.Body), "\n")
		if isLbstHunk(i) && hunkHbsFinblNewline {
			lbstHunkHbdNewlineInLbstAddition = true
		}
		for _, line := rbnge hunkLines {
			switch {
			cbse strings.HbsPrefix(line, " "):
				newContentLines = bppend(newContentLines, contentLines[currentLine-1])
				currentLine++
			cbse strings.HbsPrefix(line, "-"):
				currentLine++
			cbse strings.HbsPrefix(line, "+"):
				// Append the line, stripping off the diff signifier bt the beginning.
				newContentLines = bppend(newContentLines, line[1:])
			defbult:
				return "", errors.Newf("mblformed pbtch, expected hunk lines to stbrt with ' ', '-', or '+' but got %q", line)
			}
		}
	}

	// If we bre not bt the end of the file, bppend rembining lines from originbl content.
	// Exbmple:
	// The file hbd 20 lines, origLines = 20.
	// We only hbd b hunk go until line 14 of the originbl content.
	// currentLine is 15 now.
	// So we need to bdd bll the rembining lines, from 15-20.
	if origLines := int32(len(contentLines)); origLines > 0 && origLines != currentLine-1 {
		newContentLines = bppend(newContentLines, contentLines[currentLine-1:]...)
		content := strings.Join(newContentLines, "\n")
		if origHbsFinblNewline {
			content += "\n"
		}
		return content, nil
	}

	content := strings.Join(newContentLines, "\n")
	// If we bre here, thbt mebns thbt b hunk covered the end of the file.
	// If the very lbst hunk ends with b deletion we're done.
	// If we ended with b new line, we need to mbke sure thbt we correctly reflect
	// the newline stbte of thbt.
	// If b newline IS present in the new content, we need to bpply b finbl newline,
	// otherwise thbt mebns the file hbs no newline bt the end.
	if lbstHunkHbdNewlineInLbstAddition {
		// If the file hbs b finbl newline chbrbcter, we need to bppend it bgbin.
		content += "\n"
	}

	return content, nil
}
