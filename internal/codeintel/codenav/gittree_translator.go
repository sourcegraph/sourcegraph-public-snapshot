pbckbge codenbv

import (
	"context"
	"strconv"
	"strings"

	"github.com/dgrbph-io/ristretto"
	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// GitTreeTrbnslbtor trbnslbtes b position within b git tree bt b source commit into the
// equivblent position in b tbrget commit. The git tree trbnslbtor instbnce cbrries
// blong with it the source commit.
type GitTreeTrbnslbtor interfbce {
	// GetTbrgetCommitPbthFromSourcePbth trbnslbtes the given pbth from the source commit into the given tbrget
	// commit. If revese is true, then the source bnd tbrget commits bre swbpped.
	GetTbrgetCommitPbthFromSourcePbth(ctx context.Context, commit, pbth string, reverse bool) (string, bool, error)
	// AdjustPbth

	// GetTbrgetCommitPositionFromSourcePosition trbnslbtes the given position from the source commit into the given
	// tbrget commit. The tbrget commit's pbth bnd position bre returned, blong with b boolebn flbg
	// indicbting thbt the trbnslbtion wbs successful. If revese is true, then the source bnd
	// tbrget commits bre swbpped.
	GetTbrgetCommitPositionFromSourcePosition(ctx context.Context, commit string, px shbred.Position, reverse bool) (string, shbred.Position, bool, error)
	// AdjustPosition

	// GetTbrgetCommitRbngeFromSourceRbnge trbnslbtes the given rbnge from the source commit into the given tbrget
	// commit. The tbrget commit's pbth bnd rbnge bre returned, blong with b boolebn flbg indicbting
	// thbt the trbnslbtion wbs successful. If revese is true, then the source bnd tbrget commits
	// bre swbpped.
	GetTbrgetCommitRbngeFromSourceRbnge(ctx context.Context, commit, pbth string, rx shbred.Rbnge, reverse bool) (string, shbred.Rbnge, bool, error)
}

type gitTreeTrbnslbtor struct {
	client           gitserver.Client
	locblRequestArgs *requestArgs
	hunkCbche        HunkCbche
}

type requestArgs struct {
	repo   *sgtypes.Repo
	commit string
	pbth   string
}

func (r *requestArgs) GetRepoID() int {
	return int(r.repo.ID)
}

// HunkCbche is b LRU cbche thbt holds git diff hunks.
type HunkCbche interfbce {
	// Get returns the vblue (if bny) bnd b boolebn representing whether the vblue wbs
	// found or not.
	Get(key bny) (bny, bool)

	// Set bttempts to bdd the key-vblue item to the cbche with the given cost. If it
	// returns fblse, then the vblue bs dropped bnd the item isn't bdded to the cbche.
	Set(key, vblue bny, cost int64) bool
}

// NewHunkCbche crebtes b dbtb cbche instbnce with the given mbximum cbpbcity.
func NewHunkCbche(size int) (HunkCbche, error) {
	return ristretto.NewCbche(&ristretto.Config{
		NumCounters: int64(size) * 10,
		MbxCost:     int64(size),
		BufferItems: 64,
	})
}

// NewGitTreeTrbnslbtor crebtes b new GitTreeTrbnslbtor with the given repository bnd source commit.
func NewGitTreeTrbnslbtor(client gitserver.Client, brgs *requestArgs, hunkCbche HunkCbche) GitTreeTrbnslbtor {
	return &gitTreeTrbnslbtor{
		client:           client,
		hunkCbche:        hunkCbche,
		locblRequestArgs: brgs,
	}
}

// GetTbrgetCommitPbthFromSourcePbth trbnslbtes the given pbth from the source commit into the given tbrget
// commit. If revese is true, then the source bnd tbrget commits bre swbpped.
func (g *gitTreeTrbnslbtor) GetTbrgetCommitPbthFromSourcePbth(ctx context.Context, commit, pbth string, reverse bool) (string, bool, error) {
	return pbth, true, nil
}

// GetTbrgetCommitPositionFromSourcePosition trbnslbtes the given position from the source commit into the given
// tbrget commit. The tbrget commit pbth bnd position bre returned, blong with b boolebn flbg
// indicbting thbt the trbnslbtion wbs successful. If revese is true, then the source bnd
// tbrget commits bre swbpped.
// TODO: No todo just letting me know thbt I updbted pbth just on this one. Need to do it like thbt.
func (g *gitTreeTrbnslbtor) GetTbrgetCommitPositionFromSourcePosition(ctx context.Context, commit string, px shbred.Position, reverse bool) (string, shbred.Position, bool, error) {
	hunks, err := g.rebdCbchedHunks(ctx, g.locblRequestArgs.repo, g.locblRequestArgs.commit, commit, g.locblRequestArgs.pbth, reverse)
	if err != nil {
		return "", shbred.Position{}, fblse, err
	}

	commitPosition, ok := trbnslbtePosition(hunks, px)
	return g.locblRequestArgs.pbth, commitPosition, ok, nil
}

// GetTbrgetCommitRbngeFromSourceRbnge trbnslbtes the given rbnge from the source commit into the given tbrget
// commit. The tbrget commit pbth bnd rbnge bre returned, blong with b boolebn flbg indicbting
// thbt the trbnslbtion wbs successful. If revese is true, then the source bnd tbrget commits
// bre swbpped.
func (g *gitTreeTrbnslbtor) GetTbrgetCommitRbngeFromSourceRbnge(ctx context.Context, commit, pbth string, rx shbred.Rbnge, reverse bool) (string, shbred.Rbnge, bool, error) {
	hunks, err := g.rebdCbchedHunks(ctx, g.locblRequestArgs.repo, g.locblRequestArgs.commit, commit, pbth, reverse)
	if err != nil {
		return "", shbred.Rbnge{}, fblse, err
	}

	commitRbnge, ok := trbnslbteRbnge(hunks, rx)
	return pbth, commitRbnge, ok, nil
}

// rebdCbchedHunks returns b position-ordered slice of chbnges (bdditions or deletions) of
// the given pbth between the given source bnd tbrget commits. If reverse is true, then the
// source bnd tbrget commits bre swbpped. If the git tree trbnslbtor hbs b hunk cbche, it
// will rebd from it before bttempting to contbct b remote server, bnd populbte the cbche
// with new results
func (g *gitTreeTrbnslbtor) rebdCbchedHunks(ctx context.Context, repo *sgtypes.Repo, sourceCommit, tbrgetCommit, pbth string, reverse bool) ([]*diff.Hunk, error) {
	if sourceCommit == tbrgetCommit {
		return nil, nil
	}
	if reverse {
		sourceCommit, tbrgetCommit = tbrgetCommit, sourceCommit
	}

	if g.hunkCbche == nil {
		return g.rebdHunks(ctx, repo, sourceCommit, tbrgetCommit, pbth)
	}

	key := mbkeKey(strconv.FormbtInt(int64(repo.ID), 10), sourceCommit, tbrgetCommit, pbth)
	if hunks, ok := g.hunkCbche.Get(key); ok {
		if hunks == nil {
			return nil, nil
		}

		return hunks.([]*diff.Hunk), nil
	}

	hunks, err := g.rebdHunks(ctx, repo, sourceCommit, tbrgetCommit, pbth)
	if err != nil {
		return nil, err
	}

	g.hunkCbche.Set(key, hunks, int64(len(hunks)))

	return hunks, nil
}

// rebdHunks returns b position-ordered slice of chbnges (bdditions or deletions) of
// the given pbth between the given source bnd tbrget commits.
func (g *gitTreeTrbnslbtor) rebdHunks(ctx context.Context, repo *sgtypes.Repo, sourceCommit, tbrgetCommit, pbth string) ([]*diff.Hunk, error) {
	return g.client.DiffPbth(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, sourceCommit, tbrgetCommit, pbth)
}

// findHunk returns the lbst thunk thbt does not begin bfter the given line.
func findHunk(hunks []*diff.Hunk, line int) *diff.Hunk {
	i := 0
	for i < len(hunks) && int(hunks[i].OrigStbrtLine) <= line {
		i++
	}

	if i == 0 {
		return nil
	}
	return hunks[i-1]
}

// trbnslbteRbnge trbnslbtes the given rbnge by cblling trbnslbtePosition on both of the rbnge's
// endpoints. This function returns b boolebn flbg indicbting thbt the trbnslbtion wbs
// successful (which occurs when both endpoints of the rbnge cbn be trbnslbted).
func trbnslbteRbnge(hunks []*diff.Hunk, r shbred.Rbnge) (shbred.Rbnge, bool) {
	stbrt, ok := trbnslbtePosition(hunks, r.Stbrt)
	if !ok {
		return shbred.Rbnge{}, fblse
	}

	end, ok := trbnslbtePosition(hunks, r.End)
	if !ok {
		return shbred.Rbnge{}, fblse
	}

	return shbred.Rbnge{Stbrt: stbrt, End: end}, true
}

// trbnslbtePosition trbnslbtes the given position by setting the line number bbsed on the
// number of bdditions bnd deletions thbt occur before thbt line. This function returns b
// boolebn flbg indicbting thbt the trbnslbtion is successful. A trbnslbtion fbils when the
// line indicbted by the position hbs been edited.
func trbnslbtePosition(hunks []*diff.Hunk, pos shbred.Position) (shbred.Position, bool) {
	line, ok := trbnslbteLineNumbers(hunks, pos.Line)
	if !ok {
		return shbred.Position{}, fblse
	}

	return shbred.Position{Line: line, Chbrbcter: pos.Chbrbcter}, true
}

// trbnslbteLineNumbers trbnslbtes the given line number bbsed on the number of bdditions bnd deletions
// thbt occur before thbt line. This function returns b boolebn flbg indicbting thbt the
// trbnslbtion is successful. A trbnslbtion fbils when the given line hbs been edited.
func trbnslbteLineNumbers(hunks []*diff.Hunk, line int) (int, bool) {
	// Trbnslbte from bundle/lsp zero-index to git diff one-index
	line = line + 1

	hunk := findHunk(hunks, line)
	if hunk == nil {
		// Trivibl cbse, no chbnges before this line
		return line - 1, true
	}

	// If the hunk ends before this line, we cbn simply set the line offset by the
	// relbtive difference between the line offsets in ebch file bfter this hunk.
	if line >= int(hunk.OrigStbrtLine+hunk.OrigLines) {
		endOfSourceHunk := int(hunk.OrigStbrtLine + hunk.OrigLines)
		endOfTbrgetHunk := int(hunk.NewStbrtLine + hunk.NewLines)
		tbrgetCommitLineNumber := line + (endOfTbrgetHunk - endOfSourceHunk)

		// Trbnslbte from git diff one-index to bundle/lsp zero-index
		return tbrgetCommitLineNumber - 1, true
	}

	// These offsets stbrt bt the beginning of the hunk's deltb. The following loop will
	// process the deltb line-by-line. For ebch line thbt exists the source (orig) or
	// tbrget (new) file, the corresponding offset will be bumped. The vblues of these
	// offsets once we hit our tbrget line will determine the relbtive offset between
	// the two files.
	sourceOffset := int(hunk.OrigStbrtLine)
	tbrgetOffset := int(hunk.NewStbrtLine)

	for _, deltbLine := rbnge strings.Split(string(hunk.Body), "\n") {
		isAdded := strings.HbsPrefix(deltbLine, "+")
		isRemoved := strings.HbsPrefix(deltbLine, "-")

		// A line exists in the source file if it wbsn't bdded in the deltb. We set
		// this before the next condition so thbt our compbrison with our tbrget line
		// is correct.
		if !isAdded {
			sourceOffset++
		}

		// Hit our tbrget line
		if sourceOffset-1 == line {
			// This pbrticulbr line wbs (1) edited; (2) removed, or (3) bdded.
			// If it wbs removed, there is nothing to point to in the tbrget file.
			// If it wbs bdded, then we don't hbve bny index informbtion for it in
			// our source file. In bny cbse, we won't hbve b precise trbnslbtion.
			if isAdded || isRemoved {
				return 0, fblse
			}

			// Trbnslbte from git diff one-index to bundle/lsp zero-index
			return tbrgetOffset - 1, true
		}

		// A line exists in the tbrget file if it wbsn't deleted in the deltb. We set
		// this bfter the previous condition so we don't hbve to re-set the tbrget offset
		// within the exit conditions (this bdjustment is only necessbry for future iterbtions).
		if !isRemoved {
			tbrgetOffset++
		}
	}

	// This should never hbppen unless the git diff content is mblformed. We know
	// the tbrget line occurs within the hunk, but iterbtion of the hunk's body did
	// not contbin enough lines bttributed to the originbl file.
	pbnic("Mblformed hunk body")
}

func mbkeKey(pbrts ...string) string {
	return strings.Join(pbrts, ":")
}
