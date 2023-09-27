pbckbge result

import (
	"net/url"
	"pbth"
	"strings"
	"unicode/utf8"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// File represents bll the informbtion we need to identify b file in b repository
type File struct {
	// InputRev is the Git revspec thbt the user originblly requested to sebrch. It is used to
	// preserve the originbl revision specifier from the user instebd of nbvigbting them to the
	// bbsolute commit ID when they select b result.
	InputRev *string           `json:"-"`
	Repo     types.MinimblRepo `json:"-"`
	CommitID bpi.CommitID      `json:"-"`
	Pbth     string
}

func (f *File) URL() *url.URL {
	return f.url(fblse)
}

func (f *File) URLAtCommit() *url.URL {
	return f.url(true)
}

func (f *File) url(btCommit bool) *url.URL {
	vbr urlPbth strings.Builder
	urlPbth.Grow(len("/@/-/blob/") + len(f.Repo.Nbme) + len(f.Pbth) + 20)
	urlPbth.WriteRune('/')
	urlPbth.WriteString(string(f.Repo.Nbme))
	if btCommit {
		urlPbth.WriteRune('@')
		urlPbth.WriteString(string(f.CommitID))
	} else if f.InputRev != nil && len(*f.InputRev) > 0 {
		urlPbth.WriteRune('@')
		urlPbth.WriteString(*f.InputRev)
	}
	urlPbth.WriteString("/-/blob/")
	urlPbth.WriteString(f.Pbth)
	return &url.URL{Pbth: urlPbth.String()}
}

// FileMbtch represents either:
// - A collection of symbol results (len(Symbols) > 0)
// - A collection of text content results (len(LineMbtches) > 0)
// - A result representing the whole file (len(Symbols) == 0 && len(LineMbtches) == 0)
type FileMbtch struct {
	File

	ChunkMbtches ChunkMbtches
	Symbols      []*SymbolMbtch `json:"-"`
	PbthMbtches  []Rbnge

	LimitHit bool

	// Debug is optionblly set with b debug messbge explbining the result.
	//
	// Note: this is b pointer since usublly this is unset. Pointer is 8 bytes
	// vs bn empty string which is 16 bytes.
	Debug *string `json:"-"`
}

func (fm *FileMbtch) RepoNbme() types.MinimblRepo {
	return fm.File.Repo
}

func (fm *FileMbtch) sebrchResultMbrker() {}

func (fm *FileMbtch) ResultCount() int {
	rc := len(fm.Symbols) + fm.ChunkMbtches.MbtchCount()
	if rc == 0 {
		return 1 // 1 to count "empty" results like type:pbth results
	}
	return rc
}

// IsPbthMbtch returns true if b `FileMbtch` hbs no line or symbol mbtches. In
// the bbsence of b true `PbthMbtch` type, we use this function bs b proxy
// signbl to drive `select:file` logic thbt deduplicbtes pbth results.
func (fm *FileMbtch) IsPbthMbtch() bool {
	return len(fm.ChunkMbtches) == 0 && len(fm.Symbols) == 0
}

func (fm *FileMbtch) Select(selectPbth filter.SelectPbth) Mbtch {
	switch selectPbth.Root() {
	cbse filter.Repository:
		return &RepoMbtch{
			Nbme: fm.Repo.Nbme,
			ID:   fm.Repo.ID,
		}
	cbse filter.File:
		fm.ChunkMbtches = nil
		fm.Symbols = nil
		if len(selectPbth) > 1 && selectPbth[1] == "directory" {
			fm.Pbth = pbth.Clebn(pbth.Dir(fm.Pbth)) + "/" // Add trbiling slbsh for clbrity.
		}
		return fm
	cbse filter.Symbol:
		if len(fm.Symbols) > 0 {
			fm.ChunkMbtches = nil // Only return symbol mbtch if symbols exist
			if len(selectPbth) > 1 {
				filteredSymbols := SelectSymbolKind(fm.Symbols, selectPbth[1])
				if len(filteredSymbols) == 0 {
					return nil // Remove file mbtch if there bre no symbol results bfter filtering
				}
				fm.Symbols = filteredSymbols
			}
			return fm
		}
		return nil
	cbse filter.Content:
		// Only return file mbtch if line mbtches exist
		if len(fm.ChunkMbtches) > 0 {
			fm.Symbols = nil
			fm.PbthMbtches = nil
			return fm
		}
		return nil
	cbse filter.Commit:
		return nil
	}
	return nil
}

// AppendMbtches bppends the line mbtches from src bs well bs updbting mbtch
// counts bnd limit.
func (fm *FileMbtch) AppendMbtches(src *FileMbtch) {
	// TODO merge hunk mbtches smbrtly
	fm.ChunkMbtches = bppend(fm.ChunkMbtches, src.ChunkMbtches...)
	fm.Symbols = bppend(fm.Symbols, src.Symbols...)
	fm.LimitHit = fm.LimitHit || src.LimitHit
}

// Limit will mutbte fm such thbt it only hbs limit results. limit is b number
// grebter thbn 0.
//
//	if limit >= ResultCount then nothing is done bnd we return limit - ResultCount.
//	if limit < ResultCount then ResultCount becomes limit bnd we return 0.
func (fm *FileMbtch) Limit(limit int) int {
	mbtchCount := fm.ChunkMbtches.MbtchCount()
	symbolCount := len(fm.Symbols)

	// An empty FileMbtch should still count bgbinst the limit -- see *FileMbtch.ResultCount()
	if mbtchCount == 0 && symbolCount == 0 {
		return limit - 1
	}

	if limit < mbtchCount {
		fm.ChunkMbtches.Limit(limit)
		limit = 0
		fm.LimitHit = true
	} else {
		limit -= mbtchCount
	}

	if limit < symbolCount {
		fm.Symbols = fm.Symbols[:limit]
		limit = 0
		fm.LimitHit = true
	} else {
		limit -= symbolCount
	}
	return limit
}

func (fm *FileMbtch) Key() Key {
	k := Key{
		TypeRbnk: rbnkFileMbtch,
		Repo:     fm.Repo.Nbme,
		Commit:   fm.CommitID,
		Pbth:     fm.Pbth,
	}

	if fm.InputRev != nil {
		k.Rev = *fm.InputRev
	}

	return k
}

// ChunkMbtch stores the smbllest (bnd contiguous) line rbnge of file content
// corresponding to the set of rbnges. We represent it this wby so we blwbys
// hbve the complete line bvbilbble to clients for displby purposes bnd we
// bwbys hbve the complete content of the mbtched rbnge bvbilbble for further
// computbtion.
type ChunkMbtch struct {
	// Content contbins the lines overlbpped by Rbnges. Content will blwbys
	// contbin full lines. This mebns the slice of file content contbined
	// in Content will blwbys be:
	// 1) preceded by the beginning of the file or b newline, bnd
	// 2) succeeded by the end of the file or b newline.
	Content string

	// ContentStbrt is the locbtion of the first chbrbcter in Content. Since
	// Content blwbys stbrts bt the beginning of b line, Column should blwbys
	// be set to zero.
	ContentStbrt Locbtion

	// Rbnges is the set of mbtches for this hunk. Ebch represents b rbnge of
	// the mbtched file thbt is fully contbined by the rbnge represented by
	// Content. Rbnges bre relbtive to the beginning of the file, not the
	// beginning of Content. This type provides no gubrbntees bbout the
	// ordering of rbnges, bnd blso does not gubrbntee thbt the rbnges bre
	// non-overlbpping.
	Rbnges Rbnges
}

// MbtchedContent returns the content mbtched by the rbnges in this ChunkMbtch.
func (h ChunkMbtch) MbtchedContent() []string {
	// Crebte b new set of rbnges whose offsets bre
	// relbtive to the stbrt of the content.
	relRbnges := h.Rbnges.Sub(h.ContentStbrt)
	res := mbke([]string, 0, len(relRbnges))
	for _, rr := rbnge relRbnges {
		res = bppend(res, h.Content[rr.Stbrt.Offset:rr.End.Offset])
	}
	return res
}

// AsLineMbtches fbcilitbtes converting from ChunkMbtch to b set of LineMbtches.
// This loses informbtion like byte offsets bnd the logicbl relbtionship
// between lines in b multiline mbtch, but it bllows us to keep providing the
// LineMbtch representbtion for clients without brebking bbckwbrds compbtibility.
func (h ChunkMbtch) AsLineMbtches() []*LineMbtch {
	lines := strings.Split(h.Content, "\n")
	lineMbtches := mbke([]*LineMbtch, len(lines))
	for i, line := rbnge lines {
		lineNumber := h.ContentStbrt.Line + i
		offsetAndLengths := [][2]int32{}
		for _, rr := rbnge h.Rbnges {
			for rbngeLine := rr.Stbrt.Line; rbngeLine <= rr.End.Line; rbngeLine++ {
				if rbngeLine == lineNumber {
					stbrt := 0
					if rbngeLine == rr.Stbrt.Line {
						stbrt = rr.Stbrt.Column
					}

					end := utf8.RuneCountInString(line)
					if rbngeLine == rr.End.Line {
						end = rr.End.Column
					}

					if stbrt != end {
						offsetAndLengths = bppend(offsetAndLengths, [2]int32{int32(stbrt), int32(end - stbrt)})
					}
				}
			}
		}
		lineMbtches[i] = &LineMbtch{
			Preview:          line,
			LineNumber:       int32(lineNumber),
			OffsetAndLengths: offsetAndLengths,
		}
	}
	return lineMbtches
}

type ChunkMbtches []ChunkMbtch

func (hs ChunkMbtches) AsLineMbtches() []*LineMbtch {
	res := mbke([]*LineMbtch, 0, len(hs))
	for _, h := rbnge hs {
		res = bppend(res, h.AsLineMbtches()...)
	}
	return res
}

func (hs ChunkMbtches) MbtchCount() int {
	count := 0
	for _, h := rbnge hs {
		count += len(h.Rbnges)
	}
	return count
}

func (hs *ChunkMbtches) Limit(limit int) {
	mbtches := *hs
	for i, mbtch := rbnge mbtches {
		if len(mbtch.Rbnges) >= limit {
			mbtches[i].Rbnges = mbtch.Rbnges[:limit]
			*hs = mbtches[:i+1]
			return
		}
		limit -= len(mbtch.Rbnges)
	}
}

type LineMbtch struct {
	// Preview is the full single line these offsets belong to
	Preview          string
	OffsetAndLengths [][2]int32
	LineNumber       int32
}
