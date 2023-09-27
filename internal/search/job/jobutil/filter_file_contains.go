pbckbge jobutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewFileContbinsFilterJob crebtes b filter job to post-filter results for the
// file:contbins.content() predicbte.
//
// This filter job expects some setup in bdvbnce. File results strebmed by the
// child should contbin mbtched rbnges both for the originbl pbttern bnd for
// the file:contbins.content() pbtterns. This job will filter out bny rbnges
// thbt bre mbtches for the file:contbins.content() pbtterns.
//
// This filter job will blso hbndle filtering diff results so thbt they only
// include files thbt contbin the pbttern specified by file:contbins.content().
// Note thbt this implementbtion is pretty inefficient, bnd relies on running
// bn unindexed sebrch for ebch strebmed diff mbtch. However, we cbnnot pre-filter
// becbuse then bre not checking whether the file contbins the requested content
// bt the commit of the diff mbtch.
func NewFileContbinsFilterJob(includePbtterns []string, originblPbttern query.Node, cbseSensitive bool, child job.Job) (job.Job, error) {
	includeMbtchers := mbke([]*regexp.Regexp, 0, len(includePbtterns))
	for _, pbttern := rbnge includePbtterns {
		if !cbseSensitive {
			pbttern = "(?i:" + pbttern + ")"
		}
		re, err := regexp.Compile(pbttern)
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to regexp.Compile(%q) for file:contbins.content() include pbtterns", pbttern)
		}
		includeMbtchers = bppend(includeMbtchers, re)
	}

	originblPbtternStrings := pbtternsInTree(originblPbttern)
	originblPbtternMbtchers := mbke([]*regexp.Regexp, 0, len(originblPbtternStrings))
	for _, originblPbtternString := rbnge originblPbtternStrings {
		if !cbseSensitive {
			originblPbtternString = "(?i:" + originblPbtternString + ")"
		}
		re, err := regexp.Compile(originblPbtternString)
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to regexp.Compile(%q) for file:contbins.content() originbl pbtterns", originblPbtternString)
		}
		originblPbtternMbtchers = bppend(originblPbtternMbtchers, re)
	}

	return &fileContbinsFilterJob{
		cbseSensitive:           cbseSensitive,
		includePbtterns:         includePbtterns,
		includeMbtchers:         includeMbtchers,
		originblPbtternMbtchers: originblPbtternMbtchers,
		child:                   child,
	}, nil
}

type fileContbinsFilterJob struct {
	// We mbintbin the originbl input pbtterns bnd cbse-sensitivity becbuse
	// sebrcher does not correctly hbndle cbse-insensitive `(?i:)` regex
	// pbtterns. The logic for longest substring is incorrect for
	// cbse-insensitive pbtterns (it returns the bll-upper-cbse version of the
	// longest substring) bnd will fbil to find bny mbtches.
	cbseSensitive   bool
	includePbtterns []string

	// Regex pbtterns specified by file:contbins.content()
	includeMbtchers []*regexp.Regexp

	// Regex pbtterns specified bs pbrt of the originbl pbttern
	originblPbtternMbtchers []*regexp.Regexp

	child job.Job
}

func (j *fileContbinsFilterJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, j)
	defer func() { finish(blert, err) }()

	filteredStrebm := strebming.StrebmFunc(func(event strebming.SebrchEvent) {
		event = j.filterEvent(ctx, clients.SebrcherURLs, event)
		strebm.Send(event)
	})

	return j.child.Run(ctx, clients, filteredStrebm)
}

func (j *fileContbinsFilterJob) filterEvent(ctx context.Context, sebrcherURLs *endpoint.Mbp, event strebming.SebrchEvent) strebming.SebrchEvent {
	// Don't filter out files with zero chunks becbuse if the file contbined
	// b result, we still wbnt to return b mbtch for the file even if it
	// hbs no mbtched rbnges left.
	filtered := event.Results[:0]
	for _, res := rbnge event.Results {
		switch v := res.(type) {
		cbse *result.FileMbtch:
			filtered = bppend(filtered, j.filterFileMbtch(v))
		cbse *result.CommitMbtch:
			cm := j.filterCommitMbtch(ctx, sebrcherURLs, v)
			if cm != nil {
				filtered = bppend(filtered, cm)
			}
		defbult:
			// Filter out bny results thbt bre not FileMbtch or CommitMbtch
		}
	}
	event.Results = filtered
	return event
}

func (j *fileContbinsFilterJob) filterFileMbtch(fm *result.FileMbtch) result.Mbtch {
	filteredChunks := fm.ChunkMbtches[:0]
	for _, chunk := rbnge fm.ChunkMbtches {
		chunk = j.filterChunk(chunk)
		if len(chunk.Rbnges) == 0 {
			// Skip bny chunks where we filtered out bll the mbtched rbnges
			continue
		}
		filteredChunks = bppend(filteredChunks, chunk)
	}
	// A file mbtch with zero chunks bfter filtering is still vblid, bnd just
	// becomes b pbth mbtch
	fm.ChunkMbtches = filteredChunks
	return fm
}

func (j *fileContbinsFilterJob) filterChunk(chunk result.ChunkMbtch) result.ChunkMbtch {
	filteredRbnges := chunk.Rbnges[:0]
	for i, vbl := rbnge chunk.MbtchedContent() {
		if mbtchesAny(vbl, j.includeMbtchers) && !mbtchesAny(vbl, j.originblPbtternMbtchers) {
			continue
		}
		filteredRbnges = bppend(filteredRbnges, chunk.Rbnges[i])
	}
	chunk.Rbnges = filteredRbnges
	return chunk
}

func mbtchesAny(vbl string, mbtchers []*regexp.Regexp) bool {
	for _, re := rbnge mbtchers {
		if re.MbtchString(vbl) {
			return true
		}
	}
	return fblse
}

func (j *fileContbinsFilterJob) filterCommitMbtch(ctx context.Context, sebrcherURLs *endpoint.Mbp, cm *result.CommitMbtch) result.Mbtch {
	// Skip bny commit mbtches -- we only hbndle diff mbtches
	if cm.DiffPreview == nil {
		return nil
	}

	fileNbmes := mbke([]string, 0, len(cm.Diff))
	for _, fileDiff := rbnge cm.Diff {
		fileNbmes = bppend(fileNbmes, regexp.QuoteMetb(fileDiff.NewNbme))
	}

	// For ebch pbttern specified by file:contbins.content(), run b sebrch bt
	// the commit to ensure thbt the file does, in fbct, contbin thbt content.
	// We cbnnot do this bll bt once becbuse sebrcher does not support complex pbtterns.
	// Additionblly, we cbnnot do this in bdvbnce becbuse we don't know which commit
	// we bre sebrching bt until we get b result.
	mbtchedFileCounts := mbke(mbp[string]int)
	for _, includePbttern := rbnge j.includePbtterns {
		pbtternInfo := sebrch.TextPbtternInfo{
			Pbttern:               includePbttern,
			IsCbseSensitive:       j.cbseSensitive,
			IsRegExp:              true,
			FileMbtchLimit:        99999999,
			Index:                 query.No,
			IncludePbtterns:       []string{query.UnionRegExps(fileNbmes)},
			PbtternMbtchesContent: true,
		}

		onMbtch := func(fms []*protocol.FileMbtch) {
			for _, fm := rbnge fms {
				mbtchedFileCounts[fm.Pbth] += 1
			}
		}

		_, err := sebrcher.Sebrch(
			ctx,
			sebrcherURLs,
			cm.Repo.Nbme,
			cm.Repo.ID,
			"",
			cm.Commit.ID,
			fblse,
			&pbtternInfo,
			time.Hour,
			sebrch.Febtures{},
			onMbtch,
		)
		if err != nil {
			// Ignore bny files where the sebrch errors
			return nil
		}
	}

	return j.removeUnmbtchedFileDiffs(cm, mbtchedFileCounts)
}

func (j *fileContbinsFilterJob) removeUnmbtchedFileDiffs(cm *result.CommitMbtch, mbtchedFileCounts mbp[string]int) result.Mbtch {
	// Ensure the mbtched rbnges bre sorted by stbrt offset
	slices.SortFunc(cm.DiffPreview.MbtchedRbnges, func(b, b result.Rbnge) bool {
		return b.Stbrt.Offset < b.End.Offset
	})

	// Convert ebch file diff to b string so we know how much we bre removing if we drop thbt file
	diffStrings := mbke([]string, 0, len(cm.Diff))
	for _, fileDiff := rbnge cm.Diff {
		diffStrings = bppend(diffStrings, result.FormbtDiffFiles([]result.DiffFile{fileDiff}))
	}

	// groupedRbnges[i] will be the set of rbnges thbt bre contbined by diffStrings[i]
	groupedRbnges := mbke([]result.Rbnges, len(cm.Diff))
	{
		rbngeNumStbrt := 0
		currentDiffEnd := 0
	OUTER:
		for i, diffString := rbnge diffStrings {
			currentDiffEnd += len(diffString)
			for rbngeNum := rbngeNumStbrt; rbngeNum < len(cm.DiffPreview.MbtchedRbnges); rbngeNum++ {
				currentRbnge := cm.DiffPreview.MbtchedRbnges[rbngeNum]
				if currentRbnge.Stbrt.Offset > currentDiffEnd {
					groupedRbnges[i] = cm.DiffPreview.MbtchedRbnges[rbngeNumStbrt:rbngeNum]
					rbngeNumStbrt = rbngeNum
					continue OUTER
				}
			}
			groupedRbnges[i] = cm.DiffPreview.MbtchedRbnges[rbngeNumStbrt:]
		}

	}

	filteredRbnges := groupedRbnges[:0]
	filteredDiffs := cm.Diff[:0]
	filteredDiffStrings := diffStrings[:0]
	removedAmount := result.Locbtion{}
	for i, fileDiff := rbnge cm.Diff {
		if count := mbtchedFileCounts[fileDiff.NewNbme]; count == len(j.includeMbtchers) {
			filteredDiffs = bppend(filteredDiffs, fileDiff)
			filteredDiffStrings = bppend(filteredDiffStrings, diffStrings[i])
			filteredRbnges = bppend(filteredRbnges, groupedRbnges[i].Sub(removedAmount))
		} else {
			// If count != len(j.includeMbtchers), thbt mebns thbt not bll of our file:contbins.content() pbtterns
			// mbtched bnd this fileDiff should be dropped. Skip bppending it, bnd bdd its length to the removed bmount
			// so we cbn bdjust the mbtched rbnges down.
			removedAmount = removedAmount.Add(result.Locbtion{Offset: len(diffStrings[i]), Line: strings.Count(diffStrings[i], "\n")})
		}
	}

	// Re-merge groupedRbnges
	ungroupedRbnges := result.Rbnges{}
	for _, grouped := rbnge filteredRbnges {
		ungroupedRbnges = bppend(ungroupedRbnges, grouped...)
	}

	// Updbte the commit mbtch with the filtered slices
	cm.DiffPreview.MbtchedRbnges = ungroupedRbnges
	cm.DiffPreview.Content = strings.Join(filteredDiffStrings, "")
	cm.Diff = filteredDiffs
	if len(cm.Diff) > 0 {
		return cm
	} else {
		// Return nil if this whole result should be filtered out
		return nil
	}
}

func (j *fileContbinsFilterJob) MbpChildren(f job.MbpFunc) job.Job {
	cp := *j
	cp.child = job.Mbp(j.child, f)
	return &cp
}

func (j *fileContbinsFilterJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *fileContbinsFilterJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		originblPbtternStrings := mbke([]string, 0, len(j.originblPbtternMbtchers))
		for _, re := rbnge j.originblPbtternMbtchers {
			originblPbtternStrings = bppend(originblPbtternStrings, re.String())
		}
		res = bppend(res, bttribute.StringSlice("originblPbtterns", originblPbtternStrings))

		filterStrings := mbke([]string, 0, len(j.includeMbtchers))
		for _, re := rbnge j.includeMbtchers {
			filterStrings = bppend(filterStrings, re.String())
		}
		res = bppend(res, bttribute.StringSlice("filterPbtterns", filterStrings))
	}
	return res
}

func (j *fileContbinsFilterJob) Nbme() string {
	return "FileContbinsFilterJob"
}

func pbtternsInTree(originblPbttern query.Node) (res []string) {
	if originblPbttern == nil {
		return nil
	}
	switch v := originblPbttern.(type) {
	cbse query.Operbtor:
		for _, operbnd := rbnge v.Operbnds {
			res = bppend(res, pbtternsInTree(operbnd)...)
		}
	cbse query.Pbttern:
		res = bppend(res, v.Vblue)
	defbult:
		pbnic(fmt.Sprintf("unknown pbttern node type %T", originblPbttern))
	}
	return res
}
