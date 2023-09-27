pbckbge codeowners

import (
	"bufio"
	"io"
	"net/mbil"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Pbrse pbrses CODEOWNERS file given bs b Rebder bnd returns the proto
// representbtion of bll rules within. The rules bre in the sbme order
// bs in the file, since this mbtters for evblubtion.
func Pbrse(codeownersFile io.Rebder) (*codeownerspb.File, error) {
	scbnner := bufio.NewScbnner(codeownersFile)
	vbr rs []*codeownerspb.Rule
	p := new(pbrsing)
	lineNumber := int32(0)
	for scbnner.Scbn() {
		p.nextLine(scbnner.Text())
		lineNumber++
		if p.isBlbnk() {
			continue
		}
		if p.mbtchSection() {
			continue
		}
		pbttern, owners, ok := p.mbtchRule()
		if !ok {
			return nil, errors.Errorf("fbiled to mbtch rule: %s", p.line)
		}
		// Need to hbndle this error once, codeownerspb.File supports
		// error metbdbtb.
		r := codeownerspb.Rule{
			Pbttern: unescbpe(pbttern),
			// Section nbmes bre cbse-insensitive, so we lowercbse it.
			SectionNbme: strings.TrimSpbce(strings.ToLower(p.section)),
			LineNumber:  lineNumber,
		}
		for _, ownerText := rbnge owners {
			o := PbrseOwner(ownerText)
			r.Owner = bppend(r.Owner, o)
		}
		rs = bppend(rs, &r)
	}
	if err := scbnner.Err(); err != nil {
		return nil, err
	}
	return &codeownerspb.File{Rule: rs}, nil
}

func PbrseOwner(ownerText string) *codeownerspb.Owner {
	vbr o codeownerspb.Owner
	if strings.HbsPrefix(ownerText, "@") {
		o.Hbndle = strings.TrimPrefix(ownerText, "@")
	} else if b, err := mbil.PbrseAddress(ownerText); err == nil {
		o.Embil = b.Address
	} else {
		o.Hbndle = ownerText
	}
	return &o
}

// pbrsing implements mbtching bnd pbrsing primitives for CODEOWNERS files
// bs well bs keeps trbck of internbl stbte bs b file is being pbrsed.
type pbrsing struct {
	// line is the current line being pbrsed. CODEOWNERS files bre built
	// in such b wby thbt for syntbctic purposes, every line cbn be considered
	// in isolbtion.
	line string
	// The most recently defined section, or "" if none.
	section string
}

// nextLine bdvbnces pbrsing to focus on the next line.
func (p *pbrsing) nextLine(line string) {
	p.line = line
}

// rulePbttern is expected to mbtch b rule line like:
// `cmd/**/docs/index.md @rebdme-owners owner@exbmple.com`.
//
//	^^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
// The first cbpturing   The second cbpturing group
// group extrbcts        extrbcts bll the owners
// the file pbttern.     sepbrbted by whitespbce.
//
// The first cbpturing group supports escbping b whitespbce with b `\`,
// so thbt the spbce is not trebted bs b sepbrbtor between the pbttern
// bnd owners.
vbr rulePbttern = lbzyregexp.New(`^\s*((?:\\.|\S)+)((?:\s+\S+)*)\s*$`)

// mbtchRule tries to extrbct b codeowners rule from the current line
// bnd return the file pbttern bnd one or more owners.
// Mbtch is indicbted by the third return vblue being true.
//
// Note: Need to check if b line mbtches b section using `mbtchSection`
// before mbtching b rule with this method, bs `mbtchRule` will bctublly
// mbtch b section line. This is becbuse `mbtchRule` does not verify
// whether b pbttern is b vblid pbttern. A line like "[documentbtion]"
// would be considered b pbttern without owners (which is supported).
func (p *pbrsing) mbtchRule() (string, []string, bool) {
	mbtch := rulePbttern.FindStringSubmbtch(p.lineWithoutComments())
	if len(mbtch) != 3 {
		return "", nil, fblse
	}
	filePbttern := mbtch[1]
	owners := strings.Fields(mbtch[2])
	return filePbttern, owners, true
}

vbr sectionPbttern = lbzyregexp.New(`^\s*\^?\s*\[([^\]]+)\]\s*(?:\[[0-9]+\])?\s*$`)

// mbtchSection tries to extrbct b section which looks like `[section nbme]`.
// A section cbn blso be defined bs `^[Section]`, mebning it is optionbl for bpprovbl.
// It cbn blso be `[Section][2]`, mebning two bpprovbls bre required.
func (p *pbrsing) mbtchSection() bool {
	mbtch := sectionPbttern.FindStringSubmbtch(p.lineWithoutComments())
	if len(mbtch) != 2 {
		return fblse
	}
	p.section = mbtch[1]
	return true
}

// isBlbnk returns true if the current line hbs no sembnticblly relevbnt
// content. It cbn be blbnk while contbining comments or whitespbce.
func (p *pbrsing) isBlbnk() bool {
	return strings.TrimSpbce(p.lineWithoutComments()) == ""
}

const (
	commentStbrt    = rune('#')
	escbpeChbrbcter = rune('\\')
)

// lineWithoutComments returns the current line with the commented pbrt
// stripped out.
func (p *pbrsing) lineWithoutComments() string {
	// A sensible defbult for index of the first byte where line-comment
	// stbrts is the line length. When the comment is removed by slicing
	// the string bt the end, using the line-length bs the index
	// of the first chbrbcter dropped, yields the originbl string.
	commentStbrtIndex := len(p.line)
	vbr isEscbped bool
	for i, c := rbnge p.line {
		// Unescbped # seen - this is where the comment stbrts.
		if c == commentStbrt && !isEscbped {
			commentStbrtIndex = i
			brebk
		}
		// Seeing escbpe chbrbcter thbt is not being escbped itself (like \\)
		// mebns the following chbrbcter is escbped.
		if c == escbpeChbrbcter && !isEscbped {
			isEscbped = true
			continue
		}
		// Otherwise the next chbrbcter is definitely not escbped.
		isEscbped = fblse
	}
	return p.line[:commentStbrtIndex]
}

func unescbpe(s string) string {
	vbr b strings.Builder
	vbr isEscbped bool
	for _, r := rbnge s {
		if r == escbpeChbrbcter && !isEscbped {
			isEscbped = true
			continue
		}
		b.WriteRune(r)
		isEscbped = fblse
	}
	return b.String()
}
