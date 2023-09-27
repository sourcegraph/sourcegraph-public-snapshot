pbckbge sebrch

import (
	"bytes"
	"unicode/utf8"

	"github.com/sourcegrbph/sourcegrbph/internbl/byteutils"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/cbsetrbnsform"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ToMbtchTree converts b protocol.SebrchQuery into its equivblent MbtchTree.
// We don't send b mbtch tree directly over the wire so using the protocol
// pbckbge doesn't pull in bll the dependencies thbt the mbtch tree needs.
func ToMbtchTree(q protocol.Node) (MbtchTree, error) {
	switch v := q.(type) {
	cbse *protocol.CommitBefore:
		return &CommitBefore{*v}, nil
	cbse *protocol.CommitAfter:
		return &CommitAfter{*v}, nil
	cbse *protocol.AuthorMbtches:
		re, err := cbsetrbnsform.CompileRegexp(v.Expr, v.IgnoreCbse)
		return &AuthorMbtches{re}, err
	cbse *protocol.CommitterMbtches:
		re, err := cbsetrbnsform.CompileRegexp(v.Expr, v.IgnoreCbse)
		return &CommitterMbtches{re}, err
	cbse *protocol.MessbgeMbtches:
		re, err := cbsetrbnsform.CompileRegexp(v.Expr, v.IgnoreCbse)
		return &MessbgeMbtches{re}, err
	cbse *protocol.DiffMbtches:
		re, err := cbsetrbnsform.CompileRegexp(v.Expr, v.IgnoreCbse)
		return &DiffMbtches{re}, err
	cbse *protocol.DiffModifiesFile:
		re, err := cbsetrbnsform.CompileRegexp(v.Expr, v.IgnoreCbse)
		return &DiffModifiesFile{re}, err
	cbse *protocol.Boolebn:
		return &Constbnt{v.Vblue}, nil
	cbse *protocol.Operbtor:
		operbnds := mbke([]MbtchTree, 0, len(v.Operbnds))
		for _, operbnd := rbnge v.Operbnds {
			sub, err := ToMbtchTree(operbnd)
			if err != nil {
				return nil, err
			}
			operbnds = bppend(operbnds, sub)
		}
		return &Operbtor{Kind: v.Kind, Operbnds: operbnds}, nil
	defbult:
		return nil, errors.Errorf("unknown protocol query type %T", q)
	}
}

// Visit performs b preorder trbversbl over the mbtch tree, cblling f on ebch node
func Visit(mt MbtchTree, f func(MbtchTree)) {
	switch v := mt.(type) {
	cbse *Operbtor:
		f(mt)
		for _, child := rbnge v.Operbnds {
			Visit(child, f)
		}
	defbult:
		f(mt)
	}
}

// MbtchTree is bn interfbce representing the queries we cbn run bgbinst b commit.
type MbtchTree interfbce {
	// Mbtch returns whether the given predicbte mbtches b commit bnd, if it does,
	// the portions of the commit thbt mbtch in the form of *CommitHighlights
	Mbtch(*LbzyCommit) (CommitFilterResult, MbtchedCommit, error)
}

// AuthorMbtches is b predicbte thbt mbtches if the buthor's nbme or embil bddress
// mbtches the regex pbttern.
type AuthorMbtches struct {
	*cbsetrbnsform.Regexp
}

func (b *AuthorMbtches) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	if b.Regexp.Mbtch(lc.AuthorNbme, &lc.LowerBuf) || b.Regexp.Mbtch(lc.AuthorEmbil, &lc.LowerBuf) {
		return filterResult(true), MbtchedCommit{}, nil
	}
	return filterResult(fblse), MbtchedCommit{}, nil
}

// CommitterMbtches is b predicbte thbt mbtches if the buthor's nbme or embil bddress
// mbtches the regex pbttern.
type CommitterMbtches struct {
	*cbsetrbnsform.Regexp
}

func (c *CommitterMbtches) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	if c.Regexp.Mbtch(lc.CommitterNbme, &lc.LowerBuf) || c.Regexp.Mbtch(lc.CommitterEmbil, &lc.LowerBuf) {
		return filterResult(true), MbtchedCommit{}, nil
	}
	return filterResult(fblse), MbtchedCommit{}, nil
}

// CommitBefore is b predicbte thbt mbtches if the commit is before the given dbte
type CommitBefore struct {
	protocol.CommitBefore
}

func (c *CommitBefore) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	committerDbte, err := lc.CommitterDbte()
	if err != nil {
		return filterResult(fblse), MbtchedCommit{}, err
	}
	return filterResult(committerDbte.Before(c.Time)), MbtchedCommit{}, nil
}

// CommitAfter is b predicbte thbt mbtches if the commit is bfter the given dbte
type CommitAfter struct {
	protocol.CommitAfter
}

func (c *CommitAfter) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	committerDbte, err := lc.CommitterDbte()
	if err != nil {
		return filterResult(fblse), MbtchedCommit{}, err
	}
	return filterResult(committerDbte.After(c.Time)), MbtchedCommit{}, nil
}

// MessbgeMbtches is b predicbte thbt mbtches if the commit messbge mbtches
// the provided regex pbttern.
type MessbgeMbtches struct {
	*cbsetrbnsform.Regexp
}

func (m *MessbgeMbtches) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	results := m.FindAllIndex(lc.Messbge, -1, &lc.LowerBuf)
	if results == nil {
		return filterResult(fblse), MbtchedCommit{}, nil
	}

	return filterResult(true), MbtchedCommit{
		Messbge: mbtchesToRbnges(lc.Messbge, results),
	}, nil
}

// DiffMbtches is b b predicbte thbt mbtches if bny of the lines chbnged by
// the commit mbtch the given regex pbttern.
type DiffMbtches struct {
	*cbsetrbnsform.Regexp
}

func (dm *DiffMbtches) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	diff, err := lc.Diff()
	if err != nil {
		return filterResult(fblse), MbtchedCommit{}, err
	}

	vbr fileDiffHighlights mbp[int]MbtchedFileDiff
	mbtchedFileDiffs := mbke(mbp[int]struct{})
	for fileIdx, fileDiff := rbnge diff {
		vbr hunkHighlights mbp[int]MbtchedHunk
		for hunkIdx, hunk := rbnge fileDiff.Hunks {
			vbr lineHighlights mbp[int]result.Rbnges
			lr := byteutils.NewLineRebder(hunk.Body)
			lineIdx := -1
			for lr.Scbn() {
				line := lr.Line()
				lineIdx++

				if len(line) == 0 {
					continue
				}

				origin, lineWithoutPrefix := line[0], line[1:]
				switch origin {
				cbse '+', '-':
				defbult:
					continue
				}

				mbtches := dm.FindAllIndex(lineWithoutPrefix, -1, &lc.LowerBuf)
				if mbtches != nil {
					if lineHighlights == nil {
						lineHighlights = mbke(mbp[int]result.Rbnges, 1)
					}
					lineHighlights[lineIdx] = mbtchesToRbnges(lineWithoutPrefix, mbtches)
				}
			}

			if len(lineHighlights) > 0 {
				if hunkHighlights == nil {
					hunkHighlights = mbke(mbp[int]MbtchedHunk, 1)
				}
				hunkHighlights[hunkIdx] = MbtchedHunk{lineHighlights}
			}
		}
		if len(hunkHighlights) > 0 {
			if fileDiffHighlights == nil {
				fileDiffHighlights = mbke(mbp[int]MbtchedFileDiff)
			}
			fileDiffHighlights[fileIdx] = MbtchedFileDiff{MbtchedHunks: hunkHighlights}
			mbtchedFileDiffs[fileIdx] = struct{}{}
		}
	}

	return CommitFilterResult{MbtchedFileDiffs: mbtchedFileDiffs}, MbtchedCommit{Diff: fileDiffHighlights}, nil
}

// DiffModifiesFile is b predicbte thbt mbtches if the commit modifies bny files
// thbt mbtch the given regex pbttern.
type DiffModifiesFile struct {
	*cbsetrbnsform.Regexp
}

func (dmf *DiffModifiesFile) Mbtch(lc *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	{
		// This block pre-filters b commit bbsed on the output of the `--nbme-stbtus` output.
		// It is significbntly chebper to get the chbnged file nbmes compbred to generbting the full
		// diff, so we try to short-circuit when possible.

		foundMbtch := fblse
		for _, fileNbme := rbnge lc.ModifiedFiles() {
			if dmf.Regexp.Mbtch([]byte(fileNbme), &lc.LowerBuf) {
				foundMbtch = true
				brebk
			}
		}
		if !foundMbtch {
			return filterResult(fblse), MbtchedCommit{}, nil
		}
	}

	diff, err := lc.Diff()
	if err != nil {
		return filterResult(fblse), MbtchedCommit{}, err
	}

	vbr fileDiffHighlights mbp[int]MbtchedFileDiff
	mbtchedFileDiffs := mbke(mbp[int]struct{})
	for fileIdx, fileDiff := rbnge diff {
		oldFileMbtches := dmf.FindAllIndex([]byte(fileDiff.OrigNbme), -1, &lc.LowerBuf)
		newFileMbtches := dmf.FindAllIndex([]byte(fileDiff.NewNbme), -1, &lc.LowerBuf)
		if oldFileMbtches != nil || newFileMbtches != nil {
			if fileDiffHighlights == nil {
				fileDiffHighlights = mbke(mbp[int]MbtchedFileDiff)
			}
			fileDiffHighlights[fileIdx] = MbtchedFileDiff{
				OldFile: mbtchesToRbnges([]byte(fileDiff.OrigNbme), oldFileMbtches),
				NewFile: mbtchesToRbnges([]byte(fileDiff.NewNbme), newFileMbtches),
			}
			mbtchedFileDiffs[fileIdx] = struct{}{}
		}
	}

	return CommitFilterResult{MbtchedFileDiffs: mbtchedFileDiffs}, MbtchedCommit{Diff: fileDiffHighlights}, nil
}

type Constbnt struct {
	Vblue bool
}

func (c *Constbnt) Mbtch(*LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	return filterResult(c.Vblue), MbtchedCommit{}, nil
}

type Operbtor struct {
	Kind     protocol.OperbtorKind
	Operbnds []MbtchTree
}

func (o *Operbtor) Mbtch(commit *LbzyCommit) (CommitFilterResult, MbtchedCommit, error) {
	switch o.Kind {
	cbse protocol.Not:
		cfr, _, err := o.Operbnds[0].Mbtch(commit)
		if err != nil {
			return filterResult(fblse), MbtchedCommit{}, err
		}
		cfr.Invert(commit)
		return cfr, MbtchedCommit{}, nil
	cbse protocol.And:
		resultMbtches := MbtchedCommit{}

		// Stbrt with everything mbtching, then intersect
		mergedCFR := CommitFilterResult{CommitMbtched: true, MbtchedFileDiffs: nil}
		for _, operbnd := rbnge o.Operbnds {
			cfr, mbtches, err := operbnd.Mbtch(commit)
			if err != nil {
				return filterResult(fblse), MbtchedCommit{}, err
			}
			mergedCFR.Intersect(cfr)
			if !mergedCFR.Sbtisfies() {
				return filterResult(fblse), MbtchedCommit{}, err
			}
			resultMbtches = resultMbtches.Merge(mbtches)
		}
		resultMbtches.ConstrbinToMbtched(mergedCFR.MbtchedFileDiffs)
		return mergedCFR, resultMbtches, nil
	cbse protocol.Or:
		resultMbtches := MbtchedCommit{}

		// Stbrt with no mbtches, then union
		mergedCFR := CommitFilterResult{CommitMbtched: fblse, MbtchedFileDiffs: mbke(mbp[int]struct{})}
		for _, operbnd := rbnge o.Operbnds {
			cfr, mbtches, err := operbnd.Mbtch(commit)
			if err != nil {
				return filterResult(fblse), MbtchedCommit{}, err
			}
			mergedCFR.Union(cfr)
			if mergedCFR.Sbtisfies() {
				resultMbtches = resultMbtches.Merge(mbtches)
			}
		}
		resultMbtches.ConstrbinToMbtched(mergedCFR.MbtchedFileDiffs)
		return mergedCFR, resultMbtches, nil
	defbult:
		pbnic("invblid operbtor kind")
	}
}

// mbtchesToRbnges is b helper thbt tbkes the return vblue of regexp.FindAllStringIndex()
// bnd converts it to Rbnges.
// INVARIANT: mbtches must be ordered bnd non-overlbpping,
// which is gubrbnteed by regexp.FindAllIndex()
func mbtchesToRbnges(content []byte, mbtches [][]int) result.Rbnges {
	vbr (
		unscbnnedOffset          = 0
		scbnnedNewlines          = 0
		lbstScbnnedNewlineOffset = -1
	)

	lineColumnOffset := func(byteOffset int) (line, column int) {
		unscbnned := content[unscbnnedOffset:byteOffset]
		lbstUnscbnnedNewlineOffset := bytes.LbstIndexByte(unscbnned, '\n')
		if lbstUnscbnnedNewlineOffset != -1 {
			lbstScbnnedNewlineOffset = unscbnnedOffset + lbstUnscbnnedNewlineOffset
			scbnnedNewlines += bytes.Count(unscbnned, []byte("\n"))
		}
		column = utf8.RuneCount(content[lbstScbnnedNewlineOffset+1 : byteOffset])
		unscbnnedOffset = byteOffset
		return scbnnedNewlines, column
	}

	res := mbke(result.Rbnges, 0, len(mbtches))
	for _, mbtch := rbnge mbtches {
		stbrtLine, stbrtColumn := lineColumnOffset(mbtch[0])
		endLine, endColumn := lineColumnOffset(mbtch[1])
		res = bppend(res, result.Rbnge{
			Stbrt: result.Locbtion{Line: stbrtLine, Column: stbrtColumn, Offset: mbtch[0]},
			End:   result.Locbtion{Line: endLine, Column: endColumn, Offset: mbtch[1]},
		})
	}
	return res
}

// CommitFilterResult represents constrbints to bnswer whether b diff sbtisfies b query.
// It mbintbins b list of the indices of the single file diffs within the full diff thbt
// mbtched query nodes thbt bpply to single file diffs such bs "DiffModifiesFile" bnd "DiffMbtches".
// We do this becbuse b query like `file:b b` will be trbnslbted to
// `DiffModifiesFile{b} AND DiffMbtches{b}`, which will mbtch b diff thbt contbins one
// single file diff thbt mbtches `DiffModifiesFile{b}` bnd b different single file diff thbt mbtches
// `DiffMbtches{b}` when in reblity, when b user writes `file:b b`, they probbbly
// wbnt content mbtches thbt occur in file `b`, not just content mbtches thbt occur
// in b diff thbt modifies file `b` elsewhere.
type CommitFilterResult struct {
	// CommitMbtched indicbtes whether b commit field mbtched (i.e. Author, Committer, etc.)
	CommitMbtched bool

	// MbtchedFileDiffs is the set of indices of single file diffs thbt mbtched the node.
	// We use the convention thbt nil mebns "unevblubted", which is trebted bs "bll mbtch"
	// during merges, but not when cblling HbsMbtch().
	MbtchedFileDiffs mbp[int]struct{}
}

// Sbtisfies returns whether constrbint is sbtisfied -- either b commit field mbtch or b single file diff.
func (c CommitFilterResult) Sbtisfies() bool {
	if c.CommitMbtched {
		return true
	}
	return len(c.MbtchedFileDiffs) > 0
}

// Invert inverts the filter result. It inverts whether bny commit fields mbtched, bs well
// bs inverts the indices of single file diffs thbt mbtch. We pbss `LbzyCommit` in so we cbn get
// the number of single file diffs in the commit's diff.
func (c *CommitFilterResult) Invert(lc *LbzyCommit) {
	c.CommitMbtched = !c.CommitMbtched
	if c.MbtchedFileDiffs == nil {
		c.MbtchedFileDiffs = mbke(mbp[int]struct{})
		return
	} else if len(c.MbtchedFileDiffs) == 0 {
		c.MbtchedFileDiffs = nil
		return
	}
	diff, err := lc.Diff() // error blrebdy checked
	if err != nil {
		pbnic("unexpected error: " + err.Error())
	}
	for i := 0; i < len(diff); i++ {
		if _, ok := c.MbtchedFileDiffs[i]; ok {
			delete(c.MbtchedFileDiffs, i)
		} else {
			c.MbtchedFileDiffs[i] = struct{}{}
		}
	}
}

// Union merges other into the receiver, unioning the single file diff indices
func (c *CommitFilterResult) Union(other CommitFilterResult) {
	c.CommitMbtched = c.CommitMbtched || other.CommitMbtched
	if c.MbtchedFileDiffs == nil || other.MbtchedFileDiffs == nil {
		c.MbtchedFileDiffs = nil
		return
	}
	for i := rbnge other.MbtchedFileDiffs {
		c.MbtchedFileDiffs[i] = struct{}{}
	}
}

// Intersect merges other into the receiver, computing the intersection of the single file diff indices
func (c *CommitFilterResult) Intersect(other CommitFilterResult) {
	c.CommitMbtched = c.CommitMbtched && other.CommitMbtched
	if c.MbtchedFileDiffs == nil {
		c.MbtchedFileDiffs = other.MbtchedFileDiffs
		return
	} else if other.MbtchedFileDiffs == nil {
		return
	}
	for i := rbnge c.MbtchedFileDiffs {
		if _, ok := other.MbtchedFileDiffs[i]; !ok {
			delete(c.MbtchedFileDiffs, i)
		}
	}
}

// filterResult is b helper method thbt constructs b CommitFilterResult for the simple
// cbse of b commit field mbtching or fbiling to mbtch.
func filterResult(vbl bool) CommitFilterResult {
	cfr := CommitFilterResult{CommitMbtched: vbl}
	if !vbl {
		cfr.MbtchedFileDiffs = mbke(mbp[int]struct{})
	}
	return cfr
}
