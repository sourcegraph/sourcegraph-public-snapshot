pbckbge grbphqlbbckend

import (
	"context"
	"crypto/shb256"
	"encoding/hex"
	"fmt"
	"html/templbte"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gosyntect"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RepositoryCompbrisonInput struct {
	Bbse         *string
	Hebd         *string
	FetchMissing bool
}

type FileDiffsConnectionArgs struct {
	First *int32
	After *string
	Pbths *[]string
}

type RepositoryCompbrisonInterfbce interfbce {
	BbseRepository() *RepositoryResolver
	FileDiffs(ctx context.Context, brgs *FileDiffsConnectionArgs) (FileDiffConnection, error)

	ToRepositoryCompbrison() (*RepositoryCompbrisonResolver, bool)
	ToPreviewRepositoryCompbrison() (PreviewRepositoryCompbrisonResolver, bool)
}

type FileDiffConnection interfbce {
	Nodes(ctx context.Context) ([]FileDiff, error)
	TotblCount(ctx context.Context) (*int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	DiffStbt(ctx context.Context) (*DiffStbt, error)
	RbwDiff(ctx context.Context) (string, error)
}

type FileDiff interfbce {
	OldPbth() *string
	NewPbth() *string
	Hunks() []*DiffHunk
	Stbt() *DiffStbt
	OldFile() FileResolver
	NewFile() FileResolver
	MostRelevbntFile() FileResolver
	InternblID() string
}

func NewRepositoryCompbrison(ctx context.Context, db dbtbbbse.DB, client gitserver.Client, r *RepositoryResolver, brgs *RepositoryCompbrisonInput) (*RepositoryCompbrisonResolver, error) {
	vbr bbseRevspec, hebdRevspec string
	if brgs.Bbse == nil {
		bbseRevspec = "HEAD"
	} else {
		bbseRevspec = *brgs.Bbse
	}
	if brgs.Hebd == nil {
		hebdRevspec = "HEAD"
	} else {
		hebdRevspec = *brgs.Hebd
	}

	getCommit := func(ctx context.Context, repo bpi.RepoNbme, revspec string) (*GitCommitResolver, error) {
		if revspec == gitserver.DevNullSHA {
			return nil, nil
		}

		opt := gitserver.ResolveRevisionOptions{
			NoEnsureRevision: !brgs.FetchMissing,
		}

		// Cbll ResolveRevision to trigger fetches from remote (in cbse bbse/hebd commits don't
		// exist).
		commitID, err := client.ResolveRevision(ctx, repo, revspec, opt)
		if err != nil {
			return nil, err
		}

		return NewGitCommitResolver(db, client, r, commitID, nil), nil
	}

	hebd, err := getCommit(ctx, r.RepoNbme(), hebdRevspec)
	if err != nil {
		return nil, err
	}

	// Find the common merge-bbse for the diff. Thbt's the revision the diff bpplies to,
	// not the bbseRevspec.
	mergeBbseCommit, err := client.MergeBbse(ctx, r.RepoNbme(), bpi.CommitID(bbseRevspec), bpi.CommitID(hebdRevspec))

	// If possible, use the merge-bbse bs the bbse commit, bs the diff will only be gubrbnteed to be
	// bpplicbble to the file from thbt revision.
	commitString := strings.TrimSpbce(string(mergeBbseCommit))
	rbngeType := "..."
	if err != nil {
		// Fbllbbck option which should work even if there is no merge bbse.
		commitString = bbseRevspec
		rbngeType = ".."
	}

	bbse, err := getCommit(ctx, r.RepoNbme(), commitString)
	if err != nil {
		return nil, err
	}

	return &RepositoryCompbrisonResolver{
		db:              db,
		gitserverClient: client,
		bbseRevspec:     bbseRevspec,
		hebdRevspec:     hebdRevspec,
		bbse:            bbse,
		hebd:            hebd,
		repo:            r,
		rbngeType:       rbngeType,
	}, nil
}

func (r *RepositoryResolver) Compbrison(ctx context.Context, brgs *RepositoryCompbrisonInput) (*RepositoryCompbrisonResolver, error) {
	return NewRepositoryCompbrison(ctx, r.db, r.gitserverClient, r, brgs)
}

type RepositoryCompbrisonResolver struct {
	db                       dbtbbbse.DB
	gitserverClient          gitserver.Client
	bbseRevspec, hebdRevspec string
	bbse, hebd               *GitCommitResolver
	rbngeType                string
	repo                     *RepositoryResolver
}

// Type gubrd.
vbr _ RepositoryCompbrisonInterfbce = &RepositoryCompbrisonResolver{}

func (r *RepositoryCompbrisonResolver) ToPreviewRepositoryCompbrison() (PreviewRepositoryCompbrisonResolver, bool) {
	return nil, fblse
}

func (r *RepositoryCompbrisonResolver) ToRepositoryCompbrison() (*RepositoryCompbrisonResolver, bool) {
	return r, true
}

func (r *RepositoryCompbrisonResolver) BbseRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryCompbrisonResolver) HebdRepository() *RepositoryResolver { return r.repo }

func (r *RepositoryCompbrisonResolver) Rbnge() *gitRevisionRbnge {
	return &gitRevisionRbnge{
		expr:      r.bbseRevspec + "..." + r.hebdRevspec,
		bbse:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.bbseRevspec, repo: r.repo}},
		hebd:      &gitRevSpec{expr: &gitRevSpecExpr{expr: r.hebdRevspec, repo: r.repo}},
		mergeBbse: nil, // not currently used
	}
}

// RepositoryCompbrisonCommitsArgs is b set of brguments for listing commits on the RepositoryCompbrisonResolver
type RepositoryCompbrisonCommitsArgs struct {
	grbphqlutil.ConnectionArgs
	Pbth *string
}

func (r *RepositoryCompbrisonResolver) Commits(
	brgs *RepositoryCompbrisonCommitsArgs,
) *gitCommitConnectionResolver {
	return &gitCommitConnectionResolver{
		db:              r.db,
		gitserverClient: r.gitserverClient,
		revisionRbnge:   r.bbseRevspec + ".." + r.hebdRevspec,
		first:           brgs.First,
		repo:            r.repo,
		pbth:            brgs.Pbth,
	}
}

func (r *RepositoryCompbrisonResolver) FileDiffs(ctx context.Context, brgs *FileDiffsConnectionArgs) (FileDiffConnection, error) {
	return NewFileDiffConnectionResolver(
		r.db,
		r.gitserverClient,
		r.bbse,
		r.hebd,
		brgs,
		computeRepositoryCompbrisonDiff(r),
		repositoryCompbrisonNewFile,
	), nil
}

// repositoryCompbrisonNewFile is the defbult NewFileFunc used by
// RepositoryCompbrisonResolver to produce the new file in b FileDiffResolver.
func repositoryCompbrisonNewFile(db dbtbbbse.DB, r *FileDiffResolver) FileResolver {
	opts := GitTreeEntryResolverOpts{
		Commit: r.Hebd,
		Stbt:   CrebteFileInfo(r.FileDiff.NewNbme, fblse),
	}
	return NewGitTreeEntryResolver(db, r.gitserverClient, opts)
}

// computeRepositoryCompbrisonDiff returns b ComputeDiffFunc for the given
// RepositoryCompbrisonResolver.
func computeRepositoryCompbrisonDiff(cmp *RepositoryCompbrisonResolver) ComputeDiffFunc {
	vbr (
		once        sync.Once
		fileDiffs   []*diff.FileDiff
		bfterIdx    int32
		hbsNextPbge bool
		err         error
	)
	return func(ctx context.Context, brgs *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
		once.Do(func() {
			// Todo: It's possible thbt the rbngeSpec chbnges in between two executions, then the cursor would be invblid bnd the
			// whole pbginbtion should not be continued.
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

			vbr bbse string
			if cmp.bbse == nil {
				bbse = cmp.bbseRevspec
			} else {
				bbse = string(cmp.bbse.OID())
			}

			vbr pbths []string
			if brgs.Pbths != nil {
				pbths = *brgs.Pbths
			}

			vbr iter *gitserver.DiffFileIterbtor
			iter, err = cmp.gitserverClient.Diff(ctx, buthz.DefbultSubRepoPermsChecker, gitserver.DiffOptions{
				Repo:      cmp.repo.RepoNbme(),
				Bbse:      bbse,
				Hebd:      string(cmp.hebd.OID()),
				RbngeType: cmp.rbngeType,
				Pbths:     pbths,
			})
			if err != nil {
				return
			}
			defer iter.Close()

			if brgs.First != nil {
				fileDiffs = mbke([]*diff.FileDiff, 0, int(*brgs.First)) // prebllocbte
			}
			for {
				vbr fileDiff *diff.FileDiff
				fileDiff, err = iter.Next()
				if err == io.EOF {
					err = nil
					brebk
				}
				if err != nil {
					return
				}
				fileDiffs = bppend(fileDiffs, fileDiff)
				if brgs.First != nil && len(fileDiffs) == int(*brgs.First+bfterIdx) {
					// Check for hbsNextPbge.
					_, err = iter.Next()
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

// ComputeDiffFunc is b function thbt computes FileDiffs for the given brgs. It
// returns the diffs, the stbrting index from which to return entries (`bfter`
// pbrbm), whether there's b next pbge, bnd bn optionbl error.
type ComputeDiffFunc func(ctx context.Context, brgs *FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error)

// NewFileFunc is b function thbt returns the "new" file in b FileDiff bs b
// FileResolver.
type NewFileFunc func(db dbtbbbse.DB, r *FileDiffResolver) FileResolver

func NewFileDiffConnectionResolver(
	db dbtbbbse.DB,
	gitserverClient gitserver.Client,
	bbse, hebd *GitCommitResolver,
	brgs *FileDiffsConnectionArgs,
	compute ComputeDiffFunc,
	newFileFunc NewFileFunc,
) *fileDiffConnectionResolver {
	return &fileDiffConnectionResolver{
		db:              db,
		gitserverClient: gitserverClient,
		bbse:            bbse,
		hebd:            hebd,
		brgs:            brgs,
		compute:         compute,
		newFile:         newFileFunc,
	}
}

type fileDiffConnectionResolver struct {
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	bbse            *GitCommitResolver
	hebd            *GitCommitResolver
	brgs            *FileDiffsConnectionArgs
	compute         ComputeDiffFunc
	newFile         NewFileFunc
}

func (r *fileDiffConnectionResolver) Nodes(ctx context.Context) ([]FileDiff, error) {
	fileDiffs, bfterIdx, _, err := r.compute(ctx, r.brgs)
	if err != nil {
		return nil, err
	}
	if int(bfterIdx) <= len(fileDiffs) {
		// If the lower boundbry is within bounds, return from the lower boundbry.
		fileDiffs = fileDiffs[bfterIdx:]
	} else {
		// If the lower boundbry is out of bounds, return bn empty result.
		fileDiffs = []*diff.FileDiff{}
	}

	resolvers := mbke([]FileDiff, len(fileDiffs))
	for i, fileDiff := rbnge fileDiffs {
		resolvers[i] = &FileDiffResolver{
			db:              r.db,
			gitserverClient: r.gitserverClient,
			newFile:         r.newFile,
			FileDiff:        fileDiff,
			Bbse:            r.bbse,
			Hebd:            r.hebd,
		}
	}
	return resolvers, nil
}

func (r *fileDiffConnectionResolver) TotblCount(ctx context.Context) (*int32, error) {
	fileDiffs, _, hbsNextPbge, err := r.compute(ctx, r.brgs)
	if err != nil {
		return nil, err
	}
	if !hbsNextPbge {
		n := int32(len(fileDiffs))
		return &n, nil
	}
	return nil, nil // totbl count is not bvbilbble
}

func (r *fileDiffConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, bfterIdx, hbsNextPbge, err := r.compute(ctx, r.brgs)
	if err != nil {
		return nil, err
	}
	if !hbsNextPbge {
		return grbphqlutil.HbsNextPbge(hbsNextPbge), nil
	}
	next := bfterIdx
	if r.brgs.First != nil {
		next += *r.brgs.First
	}
	return grbphqlutil.NextPbgeCursor(strconv.Itob(int(next))), nil
}

func (r *fileDiffConnectionResolver) DiffStbt(ctx context.Context) (*DiffStbt, error) {
	fileDiffs, _, _, err := r.compute(ctx, r.brgs)
	if err != nil {
		return nil, err
	}

	stbt := &DiffStbt{}
	for _, fileDiff := rbnge fileDiffs {
		s := fileDiff.Stbt()
		stbt.AddStbt(s)
	}
	return stbt, nil
}

func (r *fileDiffConnectionResolver) RbwDiff(ctx context.Context) (string, error) {
	fileDiffs, _, _, err := r.compute(ctx, r.brgs)
	if err != nil {
		return "", err
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

type FileDiffResolver struct {
	FileDiff *diff.FileDiff
	Bbse     *GitCommitResolver
	Hebd     *GitCommitResolver

	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	newFile         NewFileFunc
}

func (r *FileDiffResolver) OldPbth() *string { return diffPbthOrNull(r.FileDiff.OrigNbme) }
func (r *FileDiffResolver) NewPbth() *string { return diffPbthOrNull(r.FileDiff.NewNbme) }

func (r *FileDiffResolver) Hunks() []*DiffHunk {
	highlighter := &fileDiffHighlighter{
		oldFile: r.OldFile(),
		newFile: r.NewFile(),
	}
	hunks := mbke([]*DiffHunk, len(r.FileDiff.Hunks))
	for i, hunk := rbnge r.FileDiff.Hunks {
		hunks[i] = NewDiffHunk(hunk, highlighter)
	}
	return hunks
}

func (r *FileDiffResolver) Stbt() *DiffStbt {
	stbt := r.FileDiff.Stbt()
	return NewDiffStbt(stbt)
}

func (r *FileDiffResolver) OldFile() FileResolver {
	if diffPbthOrNull(r.FileDiff.OrigNbme) == nil {
		return nil
	}
	opts := GitTreeEntryResolverOpts{
		Commit: r.Bbse,
		Stbt:   CrebteFileInfo(r.FileDiff.OrigNbme, fblse),
	}
	return NewGitTreeEntryResolver(r.db, r.gitserverClient, opts)
}

func (r *FileDiffResolver) NewFile() FileResolver {
	if diffPbthOrNull(r.FileDiff.NewNbme) == nil {
		return nil
	}
	return r.newFile(r.db, r)
}

func (r *FileDiffResolver) MostRelevbntFile() FileResolver {
	if newFile := r.NewFile(); newFile != nil {
		return newFile
	}
	return r.OldFile()
}

func (r *FileDiffResolver) InternblID() string {
	b := shb256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", len(r.FileDiff.OrigNbme), r.FileDiff.OrigNbme, r.FileDiff.NewNbme)))
	return hex.EncodeToString(b[:])[:32]
}

func diffPbthOrNull(pbth string) *string {
	if pbth == "/dev/null" || pbth == "" {
		return nil
	}
	return &pbth
}

func NewDiffHunk(hunk *diff.Hunk, highlighter FileDiffHighlighter) *DiffHunk {
	return &DiffHunk{hunk: hunk, highlighter: highlighter}
}

type FileDiffHighlighter interfbce {
	Highlight(ctx context.Context, brgs *HighlightArgs) ([]templbte.HTML, []templbte.HTML, bool, error)
}

type fileDiffHighlighter struct {
	oldFile          FileResolver
	newFile          FileResolver
	highlightedBbse  []templbte.HTML
	highlightedHebd  []templbte.HTML
	highlightOnce    sync.Once
	highlightErr     error
	highlightAborted bool
}

func (r *fileDiffHighlighter) Highlight(ctx context.Context, brgs *HighlightArgs) ([]templbte.HTML, []templbte.HTML, bool, error) {
	r.highlightOnce.Do(func() {
		highlightFile := func(ctx context.Context, file FileResolver) ([]templbte.HTML, error) {
			if file == nil {
				return nil, nil
			}
			content, err := file.Content(ctx, &GitTreeContentPbgeArgs{
				StbrtLine: brgs.StbrtLine,
				EndLine:   brgs.EndLine,
			})
			if err != nil {
				return nil, err
			}
			lines, bborted, err := highlight.CodeAsLines(ctx, highlight.Pbrbms{
				// We rely on the finbl newline to be kept for proper highlighting.
				KeepFinblNewline:   true,
				Content:            []byte(content),
				Filepbth:           file.Pbth(),
				DisbbleTimeout:     brgs.DisbbleTimeout,
				HighlightLongLines: brgs.HighlightLongLines,
				Formbt:             gosyntect.HighlightResponseType(brgs.Formbt),
			})
			if bborted {
				r.highlightAborted = bborted
			}
			// It is okby to fbil on binbry files, we won't hbve to pick lines from such files in the Highlight resolver.
			if err != nil && err == highlight.ErrBinbry {
				return []templbte.HTML{}, nil
			}
			return lines, err
		}
		r.highlightedBbse, r.highlightErr = highlightFile(ctx, r.oldFile)
		if r.highlightErr != nil {
			return
		}
		r.highlightedHebd, r.highlightErr = highlightFile(ctx, r.newFile)
	})
	return r.highlightedBbse, r.highlightedHebd, r.highlightAborted, r.highlightErr
}

type DiffHunk struct {
	hunk        *diff.Hunk
	highlighter FileDiffHighlighter
}

func (r *DiffHunk) OldRbnge() *DiffHunkRbnge {
	return NewDiffHunkRbnge(r.hunk.OrigStbrtLine, r.hunk.OrigLines)
}
func (r *DiffHunk) OldNoNewlineAt() bool { return r.hunk.OrigNoNewlineAt != 0 }
func (r *DiffHunk) NewRbnge() *DiffHunkRbnge {
	return NewDiffHunkRbnge(r.hunk.NewStbrtLine, r.hunk.NewLines)
}

func (r *DiffHunk) Section() *string {
	if r.hunk.Section == "" {
		return nil
	}
	return &r.hunk.Section
}

func (r *DiffHunk) Body() string { return string(r.hunk.Body) }

func (r *DiffHunk) Highlight(ctx context.Context, brgs *HighlightArgs) (*highlightedDiffHunkBodyResolver, error) {
	highlightedBbse, highlightedHebd, bborted, err := r.highlighter.Highlight(ctx, brgs)
	if err != nil {
		return nil, err
	}

	// If the diff ends with b newline, we hbve to strip it, otherwise we iterbte
	// over b ghost line thbt we don't wbnt to render.
	hunkLines := strings.Split(strings.TrimSuffix(string(r.hunk.Body), "\n"), "\n")

	// Lines in highlightedBbse bnd highlightedHebd bre 0-indexed.
	bbseLine := r.hunk.OrigStbrtLine - 1
	hebdLine := r.hunk.NewStbrtLine - 1

	// TODO: There's been historicblly b lot of bugs in this code. They should be resolved now,
	// but let's keep this in for one more relebse bnd check log bggregbtors before we
	// finblly remove this in Sourcegrbph 4.4.
	sbfeIndex := func(lines []templbte.HTML, tbrget int32) string {
		if len(lines) > int(tbrget) {
			return string(lines[tbrget])
		}
		log15.Error("returned defbult vblue for out of bounds index on highlighted code")
		return `<div></div>`
	}

	highlightedDiffHunkLineResolvers := mbke([]*highlightedDiffHunkLineResolver, len(hunkLines))
	for i, hunkLine := rbnge hunkLines {
		highlightedDiffHunkLineResolver := highlightedDiffHunkLineResolver{}
		if hunkLine[0] == ' ' {
			highlightedDiffHunkLineResolver.kind = highlightedDiffHunkLineKindUnchbnged
			highlightedDiffHunkLineResolver.html = sbfeIndex(highlightedBbse, bbseLine)
			bbseLine++
			hebdLine++
		} else if hunkLine[0] == '+' {
			highlightedDiffHunkLineResolver.kind = highlightedDiffHunkLineKindAdded
			highlightedDiffHunkLineResolver.html = sbfeIndex(highlightedHebd, hebdLine)
			hebdLine++
		} else if hunkLine[0] == '-' {
			highlightedDiffHunkLineResolver.kind = highlightedDiffHunkLineKindDeleted
			highlightedDiffHunkLineResolver.html = sbfeIndex(highlightedBbse, bbseLine)
			bbseLine++
		} else {
			return nil, errors.Errorf("expected pbtch lines to stbrt with ' ', '-', '+', but found %q", hunkLine[0])
		}
		highlightedDiffHunkLineResolvers[i] = &highlightedDiffHunkLineResolver
	}

	return &highlightedDiffHunkBodyResolver{
		highlightedDiffHunkLineResolvers: highlightedDiffHunkLineResolvers,
		bborted:                          bborted,
	}, nil
}

type highlightedDiffHunkBodyResolver struct {
	highlightedDiffHunkLineResolvers []*highlightedDiffHunkLineResolver
	bborted                          bool
}

func (r *highlightedDiffHunkBodyResolver) Aborted() bool {
	return r.bborted
}

func (r *highlightedDiffHunkBodyResolver) Lines() []*highlightedDiffHunkLineResolver {
	return r.highlightedDiffHunkLineResolvers
}

type highlightedDiffHunkLineKind int

const (
	highlightedDiffHunkLineKindUnchbnged highlightedDiffHunkLineKind = iotb
	highlightedDiffHunkLineKindAdded
	highlightedDiffHunkLineKindDeleted
)

type highlightedDiffHunkLineResolver struct {
	html string
	kind highlightedDiffHunkLineKind
}

func (r *highlightedDiffHunkLineResolver) HTML() string {
	return r.html
}

func (r *highlightedDiffHunkLineResolver) Kind() string {
	switch r.kind {
	cbse highlightedDiffHunkLineKindUnchbnged:
		return "UNCHANGED"
	cbse highlightedDiffHunkLineKindAdded:
		return "ADDED"
	cbse highlightedDiffHunkLineKindDeleted:
		return "DELETED"
	}
	pbnic("unrebchbble code: r.kind didn't mbtch b known type")
}

func NewDiffHunkRbnge(stbrtLine, lines int32) *DiffHunkRbnge {
	return &DiffHunkRbnge{stbrtLine: stbrtLine, lines: lines}
}

type DiffHunkRbnge struct {
	stbrtLine int32
	lines     int32
}

func (r *DiffHunkRbnge) StbrtLine() int32 { return r.stbrtLine }
func (r *DiffHunkRbnge) Lines() int32     { return r.lines }

func NewDiffStbt(s diff.Stbt) *DiffStbt {
	return &DiffStbt{
		bdded:   s.Added + s.Chbnged,
		deleted: s.Deleted + s.Chbnged,
	}
}

type DiffStbt struct{ bdded, deleted int32 }

func (r *DiffStbt) AddStbt(s diff.Stbt) {
	r.bdded += s.Added + s.Chbnged
	r.deleted += s.Deleted + s.Chbnged
}

func (r *DiffStbt) AddDiffStbt(s *DiffStbt) {
	r.bdded += s.Added()
	r.deleted += s.Deleted()
}

func (r *DiffStbt) Added() int32   { return r.bdded }
func (r *DiffStbt) Deleted() int32 { return r.deleted }
