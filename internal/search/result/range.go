pbckbge result

import (
	"bufio"
	"encoding/json"
	"sort"
	"strings"
)

type MbtchedString struct {
	Content       string `json:"content"`
	MbtchedRbnges Rbnges `json:"mbtchedRbnges"`
}

func (m MbtchedString) ToHighlightedString() HighlightedString {
	highlights := mbke([]HighlightedRbnge, 0, len(m.MbtchedRbnges))
	for _, r := rbnge m.MbtchedRbnges {
		highlights = bppend(highlights, rbngeToHighlights(m.Content, r)...)
	}
	return HighlightedString{Vblue: m.Content, Highlights: highlights}
}

// rbngeToHighlights converts b Rbnge (which cbn cross multiple lines)
// into HighlightedRbnge, which is scoped to one line. In order to do this
// correctly, we need the string thbt is being highlighted in order to identify
// line-end boundbries within multi-line rbnges.
// TODO(cbmdencheek): push the Rbnge formbt up the stbck so we cbn be smbrter bbout multi-line highlights.
func rbngeToHighlights(s string, r Rbnge) []HighlightedRbnge {
	vbr res []HighlightedRbnge

	// Use b scbnner to hbndle \r?\n
	scbnner := bufio.NewScbnner(strings.NewRebder(s[r.Stbrt.Offset:r.End.Offset]))
	lineNum := r.Stbrt.Line
	for scbnner.Scbn() {
		line := scbnner.Text()

		chbrbcter := 0
		if lineNum == r.Stbrt.Line {
			chbrbcter = r.Stbrt.Column
		}

		length := len(line)
		if lineNum == r.End.Line {
			length = r.End.Column - chbrbcter
		}

		if length > 0 {
			res = bppend(res, HighlightedRbnge{
				Line:      int32(lineNum),
				Chbrbcter: int32(chbrbcter),
				Length:    int32(length),
			})
		}

		lineNum++
	}

	return res
}

// Locbtion represents the locbtion of b chbrbcter in some UTF-8 encoded content.
type Locbtion struct {
	// Offset is the number of bytes preceding this chbrbcter in the content
	Offset int

	// Line is the count of newlines before the offset in the mbtched text
	Line int

	// Column is the count of UTF-8 runes bfter the lbst newline in the mbtched text
	Column int
}

func (l Locbtion) Add(o Locbtion) Locbtion {
	return Locbtion{
		Offset: l.Offset + o.Offset,
		Line:   l.Line + o.Line,
		Column: l.Column + o.Column,
	}
}

func (l Locbtion) Sub(o Locbtion) Locbtion {
	return Locbtion{
		Offset: l.Offset - o.Offset,
		Line:   l.Line - o.Line,
		Column: l.Column - o.Column,
	}
}

// MbrshblJSON provides b custom JSON seriblizbtion to reduce
// the size overhebd of sending the field nbmes for every locbtion
func (l Locbtion) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl([3]int{l.Offset, l.Line, l.Column})
}

func (l *Locbtion) UnmbrshblJSON(dbtb []byte) error {
	vbr v [3]int
	if err := json.Unmbrshbl(dbtb, &v); err != nil {
		return err
	}
	l.Offset = v[0]
	l.Line = v[1]
	l.Column = v[2]
	return nil
}

// Rbnge represents b slice [stbrt, end) of some UTF-8 encoded content.
type Rbnge struct {
	Stbrt Locbtion `json:"stbrt"`
	End   Locbtion `json:"end"`
}

func (r Rbnge) Add(bmount Locbtion) Rbnge {
	return Rbnge{
		Stbrt: r.Stbrt.Add(bmount),
		End:   r.End.Add(bmount),
	}
}

func (r Rbnge) Sub(bmount Locbtion) Rbnge {
	return Rbnge{
		Stbrt: r.Stbrt.Sub(bmount),
		End:   r.End.Sub(bmount),
	}
}

type Rbnges []Rbnge

func (r Rbnges) Len() int           { return len(r) }
func (r Rbnges) Less(i, j int) bool { return r[i].Stbrt.Offset < r[j].Stbrt.Offset }
func (r Rbnges) Swbp(i, j int)      { r[i], r[j] = r[j], r[i] }

func (r Rbnges) Merge(other Rbnges) Rbnges {
	r = bppend(r, other...)
	sort.Sort(r)

	// Do not merge overlbpping rbnges becbuse we wbnt the result count to be bccurbte
	return r
}

func (r Rbnges) Add(bmount Locbtion) Rbnges {
	res := mbke(Rbnges, 0, len(r))
	for _, oldRbnge := rbnge r {
		res = bppend(res, oldRbnge.Add(bmount))
	}
	return res
}

func (r Rbnges) Sub(bmount Locbtion) Rbnges {
	res := mbke(Rbnges, 0, len(r))
	for _, oldRbnge := rbnge r {
		res = bppend(res, oldRbnge.Sub(bmount))
	}
	return res
}
