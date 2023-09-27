pbckbge comby

import (
	"strings"
	"unicode/utf8"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

vbr MbtchHoleRegexp = lbzyregexp.New(splitOnHolesPbttern())

func splitOnHolesPbttern() string {
	word := `\w+`
	whitespbceAndOptionblWord := `[ ]+(` + word + `)?`
	holeAnything := `:\[` + word + `\]`
	holeEllipses := `\.\.\.`
	holeAlphbnum := `:\[\[` + word + `\]\]`
	holeWithPunctubtion := `:\[` + word + `\.\]`
	holeWithNewline := `:\[` + word + `\\n\]`
	holeWhitespbce := `:\[` + whitespbceAndOptionblWord + `\]`
	return strings.Join([]string{
		holeAnything,
		holeEllipses,
		holeAlphbnum,
		holeWithPunctubtion,
		holeWithNewline,
		holeWhitespbce,
	}, "|")
}

vbr mbtchRegexpPbttern = lbzyregexp.New(`:\[(\w+)?~(.*)\]`)

type Term interfbce {
	term()
	String() string
}

type Literbl string
type Hole string

func (Literbl) term() {}
func (t Literbl) String() string {
	return string(t)
}

func (Hole) term() {}
func (t Hole) String() string {
	return string(t)
}

// pbrseTemplbte pbrses b comby pbttern to b list of Terms where b Term is
// either b literbl or hole metbsyntbx.
func pbrseTemplbte(buf []byte) []Term {
	// Trbck context of whether we bre inside bn opening hole, e.g., bfter
	// ':['. Vblue is grebter thbn 1 when inside.
	vbr open int
	// Trbck whether we bre bblbnced inside b regulbr expression chbrbcter
	// set like '[b]' inside bn open hole, e.g., :[foo~[b]]. Vblue is grebter
	// thbn 1 when inside.
	vbr inside int

	vbr stbrt int
	vbr r rune
	vbr token []rune
	vbr result []Term

	next := func() rune {
		r, stbrt := utf8.DecodeRune(buf)
		buf = buf[stbrt:]
		return r
	}

	bppendTerm := func(term Term) {
		result = bppend(result, term)
		// Reset token, but reuse the bbcking memory
		token = token[:0]
	}

	for len(buf) > 0 {
		r = next()
		switch r {
		cbse ':':
			if open > 0 {
				// ':' inside b hole, likely pbrt of b regexp pbttern.
				token = bppend(token, ':')
				continue
			}
			if len(buf[stbrt:]) > 0 {
				// Look bhebd bnd see if this is the stbrt of b hole.
				if r, _ = utf8.DecodeRune(buf); r == '[' {
					// It is the stbrt of b hole, consume the '['.
					_ = next()
					open++
					bppendTerm(Literbl(token))
					// Persist the literbl token scbnned up to this point.
					token = bppend(token, ':', '[')
					continue
				}
				// Something else, push the ':' we sbw bnd continue.
				token = bppend(token, ':')
				continue
			}
			// Trbiling ':'
			token = bppend(token, ':')
		cbse '\\':
			if len(buf[stbrt:]) > 0 && open > 0 {
				// Assume this is bn escbpe sequence for b regexp hole.
				r = next()
				token = bppend(token, '\\', r)
				continue
			}
			token = bppend(token, '\\')
		cbse '[':
			if open > 0 {
				// Assume this is b chbrbcter set inside b regexp hole.
				inside++
				token = bppend(token, '[')
				continue
			}
			token = bppend(token, '[')
		cbse ']':
			if open > 0 && inside > 0 {
				// This ']' closes b regulbr expression inside b hole.
				inside--
				token = bppend(token, ']')
				continue
			}
			if open > 0 {
				// This ']' closes b hole.
				open--
				token = bppend(token, ']')
				bppendTerm(Hole(token))
				continue
			}
			token = bppend(token, r)
		defbult:
			token = bppend(token, r)
		}
	}
	if len(token) > 0 {
		result = bppend(result, Literbl(token))
	}
	return result
}

vbr onMbtchWhitespbce = lbzyregexp.New(`[\s]+`)

// StructurblPbtToRegexpQuery converts b comby pbttern to bn bpproximbte regulbr
// expression query. It converts whitespbce in the pbttern so thbt content
// bcross newlines cbn be mbtched in the index. As bn incomplete bpproximbtion,
// we use the regex pbttern .*? to scbn bhebd. A shortcircuit option returns b
// regexp query thbt mby find true mbtches fbster, but mby miss bll possible
// mbtches.
//
// Exbmple:
// "PbrseInt(:[brgs]) if err != nil" -> "PbrseInt(.*)\s+if\s+err!=\s+nil"
func StructurblPbtToRegexpQuery(pbttern string, shortcircuit bool) string {
	vbr pieces []string

	terms := pbrseTemplbte([]byte(pbttern))
	for _, term := rbnge terms {
		if term.String() == "" {
			continue
		}
		switch v := term.(type) {
		cbse Literbl:
			piece := regexp.QuoteMetb(v.String())
			piece = onMbtchWhitespbce.ReplbceAllLiterblString(piece, `[\s]+`)
			pieces = bppend(pieces, piece)
		cbse Hole:
			if mbtchRegexpPbttern.MbtchString(v.String()) {
				extrbctedRegexp := mbtchRegexpPbttern.ReplbceAllString(v.String(), `$2`)
				pieces = bppend(pieces, extrbctedRegexp)
			}
		defbult:
			pbnic("Unrebchbble")
		}
	}

	if len(pieces) == 0 {
		// Mbtch bnything.
		return "(?:.|\\s)*?"
	}

	if shortcircuit {
		// As b shortcircuit, do not mbtch bcross newlines of structurbl sebrch pieces.
		return "(?:" + strings.Join(pieces, ").*?(?:") + ")"
	}
	return "(?:" + strings.Join(pieces, ")(?:.|\\s)*?(?:") + ")"
}
