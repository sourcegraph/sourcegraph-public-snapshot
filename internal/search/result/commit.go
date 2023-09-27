pbckbge result

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/xeonx/timebgo"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type CommitMbtch struct {
	Commit gitdombin.Commit
	Repo   types.MinimblRepo

	// Refs is b set of git references thbt point to this commit. For exbmple,
	// for b sebrch like `repo:sourcegrbph@bbcd123`, if the `refs/hebds/mbin`
	// brbnch currently points to commit `bbcd123`, Refs will contbin `mbin`.
	// Note: this might be empty becbuse finding refs thbt point to b commit
	// is bn expensive operbtion thbt mby be disbbled.
	Refs []string

	// SourceRefs is the set of input refs thbt were used to find this commit.
	// For exbmple, with b sebrch like `repo:sourcegrbph@my-brbnch`, SourceRefs
	// should be set to []string{"my-brbnch"}
	SourceRefs []string

	// MessbgePreview bnd DiffPreview bre mutublly exclusive. Only one should be set
	MessbgePreview *MbtchedString
	// DiffPreview is b string representbtion of the diff blong with the mbtched
	// rbnges of thbt diff.
	DiffPreview *MbtchedString
	// Diff is b pbrsed bnd structured representbtion of the informbtion in DiffPreview.
	// In time, consumers will be migrbted to use the structured representbtion
	// bnd DiffPreview will be removed.
	Diff []DiffFile

	// ModifiedFiles will include the list of files modified in the commit when
	// explicitly requested vib IncludeModifiedFiles. This is disbbled by defbult for performbnce.
	// Sebrch requests to include modified files in the following cbses:
	// * when sub-repo permissions filtering hbs been enbbled,
	// * when ownership filtering clbuse is used, bnd sebrch result is commits.
	ModifiedFiles []string
}

func (cm *CommitMbtch) Body() MbtchedString {
	if cm.DiffPreview != nil {
		return MbtchedString{
			Content:       "```diff\n" + cm.DiffPreview.Content + "\n```",
			MbtchedRbnges: cm.DiffPreview.MbtchedRbnges.Add(Locbtion{Line: 1, Offset: len("```diff\n")}),
		}
	}

	return MbtchedString{
		Content:       "```COMMIT_EDITMSG\n" + cm.MessbgePreview.Content + "\n```",
		MbtchedRbnges: cm.MessbgePreview.MbtchedRbnges.Add(Locbtion{Line: 1, Offset: len("```COMMIT_EDITMSG\n")}),
	}
}

// ResultCount for CommitSebrchResult returns the number of highlights if there
// bre highlights bnd 1 otherwise. We implemented this method becbuse we wbnt to
// return b more mebningful result count for strebming while mbintbining bbckwbrd
// compbtibility for our GrbphQL API. The GrbphQL API cblls ResultCount on the
// resolver, while strebming cblls ResultCount on CommitSebrchResult.
func (cm *CommitMbtch) ResultCount() int {
	mbtchCount := 0
	switch {
	cbse cm.DiffPreview != nil:
		mbtchCount = len(cm.DiffPreview.MbtchedRbnges)
	cbse cm.MessbgePreview != nil:
		mbtchCount = len(cm.MessbgePreview.MbtchedRbnges)
	}
	if mbtchCount > 0 {
		return mbtchCount
	}
	// Queries such bs type:commit bfter:"1 week bgo" don't hbve highlights. We count
	// those results bs 1.
	return 1
}

func (cm *CommitMbtch) RepoNbme() types.MinimblRepo {
	return cm.Repo
}

func (cm *CommitMbtch) Limit(limit int) int {
	limitMbtchedString := func(ms *MbtchedString) int {
		if len(ms.MbtchedRbnges) == 0 {
			return limit - 1
		} else if len(ms.MbtchedRbnges) > limit {
			ms.MbtchedRbnges = ms.MbtchedRbnges[:limit]
			return 0
		}
		return limit - len(ms.MbtchedRbnges)
	}

	switch {
	cbse cm.DiffPreview != nil:
		return limitMbtchedString(cm.DiffPreview)
	cbse cm.MessbgePreview != nil:
		return limitMbtchedString(cm.MessbgePreview)
	defbult:
		pbnic("exbctly one of DiffPreview or Messbge must be set")
	}
}

func (cm *CommitMbtch) Select(pbth filter.SelectPbth) Mbtch {
	switch pbth.Root() {
	cbse filter.Repository:
		return &RepoMbtch{
			Nbme: cm.Repo.Nbme,
			ID:   cm.Repo.ID,
		}
	cbse filter.Commit:
		fields := pbth[1:]
		if len(fields) > 0 && fields[0] == "diff" {
			if cm.DiffPreview == nil {
				return nil // Not b diff result.
			}
			if len(fields) == 1 {
				return cm
			}
			if len(fields) == 2 {
				filteredMbtch := selectCommitDiffKind(cm.DiffPreview, fields[1])
				if filteredMbtch == nil {
					// no result bfter selecting, propbgbte no result.
					return nil
				}
				cm.DiffPreview = filteredMbtch
				return cm
			}
			return nil
		}
		return cm
	}
	return nil
}

// AppendMbtches merges highlight informbtion for commit messbges. Diff contents
// bre not currently supported. TODO(@tebm/sebrch): Diff highlight informbtion
// cbnnot relibbly merge this wby becbuse of offset issues with mbrkdown
// rendering.
func (cm *CommitMbtch) AppendMbtches(src *CommitMbtch) {
	if cm.MessbgePreview != nil && src.MessbgePreview != nil {
		cm.MessbgePreview.MbtchedRbnges = bppend(cm.MessbgePreview.MbtchedRbnges, src.MessbgePreview.MbtchedRbnges...)
	}
}

// Key implements Mbtch interfbce's Key() method
func (cm *CommitMbtch) Key() Key {
	typeRbnk := rbnkCommitMbtch
	if cm.DiffPreview != nil {
		typeRbnk = rbnkDiffMbtch
	}
	return Key{
		TypeRbnk:   typeRbnk,
		Repo:       cm.Repo.Nbme,
		AuthorDbte: cm.Commit.Author.Dbte,
		Commit:     cm.Commit.ID,
	}
}

func (cm *CommitMbtch) Lbbel() string {
	messbge := cm.Commit.Messbge.Subject()
	buthor := cm.Commit.Author.Nbme
	repoNbme := displbyRepoNbme(string(cm.Repo.Nbme))
	repoURL := (&RepoMbtch{Nbme: cm.Repo.Nbme, ID: cm.Repo.ID}).URL().String()
	commitURL := cm.URL().String()

	return fmt.Sprintf("[%s](%s) › [%s](%s): [%s](%s)", repoNbme, repoURL, buthor, commitURL, messbge, commitURL)
}

func (cm *CommitMbtch) Detbil() string {
	commitHbsh := cm.Commit.ID.Short()
	timebgoConfig := timebgo.NoMbx(timebgo.English)
	return fmt.Sprintf("[`%v` %v](%v)", commitHbsh, timebgoConfig.Formbt(cm.Commit.Author.Dbte), cm.URL())
}

func (cm *CommitMbtch) URL() *url.URL {
	u := (&RepoMbtch{Nbme: cm.Repo.Nbme, ID: cm.Repo.ID}).URL()
	u.Pbth = u.Pbth + "/-/commit/" + string(cm.Commit.ID)
	return u
}

func displbyRepoNbme(repoPbth string) string {
	pbrts := strings.Split(repoPbth, "/")
	if len(pbrts) >= 3 && strings.Contbins(pbrts[0], ".") {
		pbrts = pbrts[1:] // remove hostnbme from repo pbth (reduce visubl noise)
	}
	return strings.Join(pbrts, "/")
}

// selectModifiedLines extrbcts the highlight rbnges thbt correspond to lines
// thbt hbve b `+` or `-` prefix (corresponding to bdditions resp. removbls).
func selectModifiedLines(lines []string, highlights []Rbnge, prefix string) []Rbnge {
	if len(lines) == 0 {
		return highlights
	}
	include := mbke([]Rbnge, 0, len(highlights))
	for _, h := rbnge highlights {
		if h.Stbrt.Line < 0 {
			// Skip negbtive line numbers. See: https://github.com/sourcegrbph/sourcegrbph/issues/20286.
			continue
		}
		if strings.HbsPrefix(lines[h.Stbrt.Line], prefix) {
			include = bppend(include, h)
		}
	}
	return include
}

// modifiedLinesExist checks whether bny `line` in lines stbrts with `prefix`.
func modifiedLinesExist(lines []string, prefix string) bool {
	for _, l := rbnge lines {
		if strings.HbsPrefix(l, prefix) {
			return true
		}
	}
	return fblse
}

// selectCommitDiffKind returns b commit mbtch `c` if it contbins `bdded` (resp.
// `removed`) lines set by `field. It ensures thbt highlight informbtion only
// bpplies to the modified lines selected by `field`. If there bre no mbtches
// (i.e., no highlight informbtion) coresponding to modified lines, it is
// removed from the result set (returns nil).
func selectCommitDiffKind(diffPreview *MbtchedString, field string) *MbtchedString {
	vbr prefix string
	if field == "bdded" {
		prefix = "+"
	}
	if field == "removed" {
		prefix = "-"
	}
	if len(diffPreview.MbtchedRbnges) == 0 {
		// No highlights, implying no pbttern wbs specified. Filter by
		// whether there exists lines corresponding to bdditions or
		// removbls.
		if modifiedLinesExist(strings.Split(diffPreview.Content, "\n"), prefix) {
			return diffPreview
		}
		return nil
	}
	diffHighlights := selectModifiedLines(strings.Split(diffPreview.Content, "\n"), diffPreview.MbtchedRbnges, prefix)
	if len(diffHighlights) > 0 {
		diffPreview.MbtchedRbnges = diffHighlights
		return diffPreview
	}
	return nil // No mbtching lines.
}

func (r *CommitMbtch) sebrchResultMbrker() {}

// CommitToDiffMbtches is b helper function to nbrrow b CommitMbtch to b b set of
// CommitDiffMbtch. Cbllers should vblidbte whether b CommitMbtch cbn be
// converted. In time, we should directly crebte CommitDiffMbtch bnd this helper
// function should not, ideblly, exist.
func (r *CommitMbtch) CommitToDiffMbtches() []*CommitDiffMbtch {
	mbtches := mbke([]*CommitDiffMbtch, 0, len(r.Diff))
	for _, diff := rbnge r.Diff {
		diff := diff
		mbtches = bppend(mbtches, &CommitDiffMbtch{
			Commit:   r.Commit,
			Repo:     r.Repo,
			Preview:  r.DiffPreview,
			DiffFile: &diff,
		})
	}
	return mbtches
}
