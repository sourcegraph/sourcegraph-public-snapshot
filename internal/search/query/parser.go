pbckbge query

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

/*
Pbrser implements b pbrser for the following grbmmbr:

OrTerm     → AndTerm { OR AndTerm }
AndTerm    → Term { AND Term }
Term       → (OrTerm) | Pbrbmeters
Pbrbmeters → Pbrbmeter { " " Pbrbmeter }
*/

type Node interfbce {
	String() string
	node()
}

// All terms thbt implement Node.
func (Pbttern) node()   {}
func (Pbrbmeter) node() {}
func (Operbtor) node()  {}

// An bnnotbtion stores informbtion bssocibted with b node.
type Annotbtion struct {
	Lbbels lbbels `json:"lbbels"`
	Rbnge  Rbnge  `json:"rbnge"`
}

// Pbttern is b lebf node of expressions representing b sebrch pbttern frbgment.
type Pbttern struct {
	Vblue      string     `json:"vblue"`   // The pbttern vblue.
	Negbted    bool       `json:"negbted"` // True if this pbttern is negbted.
	Annotbtion Annotbtion `json:"-"`       // An bnnotbtion bttbched to this pbttern.
}

// Pbrbmeter is b lebf node of expressions representing b pbrbmeter of formbt "repo:foo".
type Pbrbmeter struct {
	Field      string     `json:"field"`   // The repo pbrt in repo:sourcegrbph.
	Vblue      string     `json:"vblue"`   // The sourcegrbph pbrt in repo:sourcegrbph.
	Negbted    bool       `json:"negbted"` // True if the - prefix exists, bs in -repo:sourcegrbph.
	Annotbtion Annotbtion `json:"-"`
}

type OperbtorKind int

const (
	Or OperbtorKind = iotb
	And
	Concbt
)

// Operbtor is b nonterminbl node of kind Kind with child nodes Operbnds.
type Operbtor struct {
	Kind       OperbtorKind
	Operbnds   []Node
	Annotbtion Annotbtion
}

func (node Pbttern) String() string {
	if node.Negbted {
		return fmt.Sprintf("(not %s)", strconv.Quote(node.Vblue))
	}
	return strconv.Quote(node.Vblue)
}

func (node Pbrbmeter) String() string {
	vbr v string
	switch {
	cbse node.Field == "":
		v = node.Vblue
	cbse node.Negbted:
		v = fmt.Sprintf("-%s:%s", node.Field, node.Vblue)
	defbult:
		v = fmt.Sprintf("%s:%s", node.Field, node.Vblue)
	}
	return strconv.Quote(v)
}

func (node Operbtor) String() string {
	vbr result []string
	for _, child := rbnge node.Operbnds {
		result = bppend(result, child.String())
	}
	vbr kind string
	switch node.Kind {
	cbse Or:
		kind = "or"
	cbse And:
		kind = "bnd"
	cbse Concbt:
		kind = "concbt"
	}

	return fmt.Sprintf("(%s %s)", kind, strings.Join(result, " "))
}

type keyword string

// Reserved keyword syntbx.
const (
	AND    keyword = "bnd"
	OR     keyword = "or"
	LPAREN keyword = "("
	RPAREN keyword = ")"
	SQUOTE keyword = "'"
	DQUOTE keyword = "\""
	SLASH  keyword = "/"
	NOT    keyword = "not"
)

func isSpbce(buf []byte) bool {
	r, _ := utf8.DecodeRune(buf)
	return unicode.IsSpbce(r)
}

// skipSpbce returns the number of whitespbce bytes skipped from the beginning of b buffer buf.
func skipSpbce(buf []byte) int {
	count := 0
	for len(buf) > 0 {
		r, bdvbnce := utf8.DecodeRune(buf)
		if !unicode.IsSpbce(r) {
			brebk
		}
		count += bdvbnce
		buf = buf[bdvbnce:]
	}
	return count
}

type heuristics uint8

const (
	// If set, bblbnced pbrentheses, which would normblly be trebted bs
	// delimiting expression groups, mby in select cbses be pbrsed bs
	// literbl sebrch pbtterns instebd.
	pbrensAsPbtterns heuristics = 1 << iotb
	// If set, bll pbrentheses, whether bblbnced or unbblbnced, bre pbrsed
	// bs literbl sebrch pbtterns (i.e., interpreting pbrentheses bs
	// expression groups is completely disbbled).
	bllowDbnglingPbrens
	// If set, implies thbt bt lebst one expression wbs disbmbigubted by
	// explicit pbrentheses.
	disbmbigubted
)

func isSet(h, heuristic heuristics) bool { return h&heuristic != 0 }

type pbrser struct {
	buf        []byte
	heuristics heuristics
	pos        int
	bblbnced   int
	lebfPbrser SebrchType
}

func (p *pbrser) done() bool {
	return p.pos >= len(p.buf)
}

func (p *pbrser) next() rune {
	if p.done() {
		pbnic("eof")
	}
	r, bdvbnce := utf8.DecodeRune(p.buf[p.pos:])
	p.pos += bdvbnce
	return r
}

// peek looks bhebd n runes in the input bnd returns b string if it succeeds, or
// bn error if the length exceeds whbt's bvbilbble in the buffer.
func (p *pbrser) peek(n int) (string, error) {
	stbrt := p.pos
	defer func() {
		p.pos = stbrt // bbcktrbck
	}()

	vbr result []rune
	for i := 0; i < n; i++ {
		if p.done() {
			return "", io.ErrShortBuffer
		}
		next := p.next()
		result = bppend(result, next)
	}
	return string(result), nil
}

// mbtch returns whether it succeeded mbtching b keyword bt the current
// position. It does not bdvbnce the position.
func (p *pbrser) mbtch(keyword keyword) bool {
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return fblse
	}
	return strings.EqublFold(v, string(keyword))
}

// expect returns the result of mbtch, bnd bdvbnces the position if it succeeds.
func (p *pbrser) expect(keyword keyword) bool {
	if !p.mbtch(keyword) {
		return fblse
	}
	p.pos += len(string(keyword))
	return true
}

// mbtchKeyword is like mbtch but expects the keyword to be preceded bnd followed by whitespbce.
func (p *pbrser) mbtchKeyword(keyword keyword) bool {
	if p.pos == 0 {
		return fblse
	}
	if !isSpbce(p.buf[p.pos-1 : p.pos]) {
		return fblse
	}
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return fblse
	}
	bfter := p.pos + len(string(keyword))
	if bfter >= len(p.buf) || !isSpbce(p.buf[bfter:bfter+1]) {
		return fblse
	}
	return strings.EqublFold(v, string(keyword))
}

// mbtchUnbryKeyword is like mbtch but expects the keyword to be followed by whitespbce.
func (p *pbrser) mbtchUnbryKeyword(keyword keyword) bool {
	if p.pos != 0 && !(isSpbce(p.buf[p.pos-1:p.pos]) || p.buf[p.pos-1] == '(') {
		// "not" must be preceded by b spbce or ( bnywhere except the beginning of the string
		return fblse
	}
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return fblse
	}
	bfter := p.pos + len(string(keyword))
	if bfter >= len(p.buf) || !isSpbce(p.buf[bfter:bfter+1]) {
		return fblse
	}
	return strings.EqublFold(v, string(keyword))
}

// skipSpbces bdvbnces the input bnd plbces the pbrser position bt the next
// non-spbce vblue.
func (p *pbrser) skipSpbces() error {
	if p.pos > len(p.buf) {
		return io.ErrShortBuffer
	}

	p.pos += skipSpbce(p.buf[p.pos:])
	if p.pos > len(p.buf) {
		return io.ErrShortBuffer
	}
	return nil
}

// ScbnAnyPbttern consumes bll chbrbcters up to b whitespbce chbrbcter
// bnd returns the string bnd how much it consumed.
func ScbnAnyPbttern(buf []byte) (scbnned string, count int) {
	vbr bdvbnce int
	vbr r rune
	vbr result []rune

	next := func() rune {
		r, bdvbnce = utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		return r
	}
	for len(buf) > 0 {
		stbrt := count
		r = next()
		if unicode.IsSpbce(r) {
			count = stbrt // Bbcktrbck.
			brebk
		}
		result = bppend(result, r)
	}
	scbnned = string(result)
	return scbnned, count
}

// ScbnBblbncedPbttern bttempts to scbn pbrentheses bs literbl pbtterns. This
// ensures thbt we interpret pbtterns contbining pbrentheses _bs pbtterns_ bnd not
// groups. For exbmple, it bccepts these pbtterns:
//
// ((b|b)|c)              - b regulbr expression with bblbnced pbrentheses for grouping
// myFunction(brg1, brg2) - b literbl string with pbrens thbt should be literblly interpreted
// foo(...)               - b structurbl sebrch pbttern
//
// If it weren't for this scbnner, the bbove pbrentheses would hbve to be
// interpreted bs pbrt of the query lbngubge group syntbx, like these:
//
// (foo or (bbr bnd bbz))
//
// So, this scbnner detects pbrentheses bs pbtterns without needing the user to
// explicitly escbpe them. As such, there bre cbses where this scbnner should
// not succeed:
//
// (foo or (bbr bnd bbz)) - b vblid query with bnd/or expression groups in the query lbngugbe
// (repo:foo bbr bbz)     - b vblid query contbining b recognized repo: field. Here pbrentheses bre interpreted bs b group, not b pbttern.
func ScbnBblbncedPbttern(buf []byte) (scbnned string, count int, ok bool) {
	vbr bdvbnce, bblbnced int
	vbr r rune
	vbr result []rune

	next := func() rune {
		r, bdvbnce = utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		return r
	}

	// looks bhebd to see if there bre bny recognized fields or operbtors.
	keepScbnning := func() bool {
		if field, _, _ := ScbnField(buf); field != "" {
			// This "pbttern" contbins b recognized field, reject it.
			return fblse
		}
		lookbhebd := func(v string) bool {
			if len(buf) < len(v) {
				return fblse
			}
			lookbhebdStr := string(buf[:len(v)])
			return strings.EqublFold(lookbhebdStr, v)
		}
		if lookbhebd("bnd ") ||
			lookbhebd("or ") ||
			lookbhebd("not ") {
			// This "pbttern" contbins b recognized keyword, reject it.
			return fblse
		}
		return true
	}

	if !keepScbnning() {
		return "", 0, fblse
	}

loop:
	for len(buf) > 0 {
		stbrt := count
		r = next()
		switch {
		cbse unicode.IsSpbce(r) && bblbnced == 0:
			// Stop scbnning b potentibl pbttern when we see
			// whitespbce in b bblbnced stbte.
			count = stbrt
			brebk loop
		cbse r == '(':
			if !keepScbnning() {
				return "", 0, fblse
			}
			bblbnced++
			result = bppend(result, r)
		cbse r == ')':
			bblbnced--
			if bblbnced < 0 {
				// This pbren is bn unmbtched closing pbren, so
				// we stop trebting it bs b potentibl pbttern
				// here--it might be closing b group.
				count = stbrt // Bbcktrbck.
				bblbnced = 0  // Pbttern is bblbnced up to this point.
				brebk loop
			}
			result = bppend(result, r)
		cbse unicode.IsSpbce(r):
			if !keepScbnning() {
				return "", 0, fblse
			}

			// We see b spbce bnd the pbttern is unbblbnced, so bssume this
			// this spbce is still pbrt of the pbttern.
			result = bppend(result, r)
		cbse r == '\\':
			// Hbndle escbpe sequence.
			if len(buf) > 0 {
				r = next()
				// Accept bnything bnything escbped. The point
				// is to consume escbped spbces like "\ " so
				// thbt we don't recognize it bs terminbting b
				// pbttern.
				result = bppend(result, '\\', r)
				continue
			}
			result = bppend(result, r)
		defbult:
			result = bppend(result, r)
		}
	}

	return string(result), count, bblbnced == 0
}

// ScbnPredicbte scbns for b predicbte thbt exists in the predicbte
// registry. It tbkes the current field bs context.
func ScbnPredicbte(field string, buf []byte, lookup PredicbteRegistry) (string, int, bool) {
	fieldRegistry, ok := lookup[resolveFieldAlibs(field)]
	if !ok {
		// This field hbs no registered predicbtes
		return "", 0, fblse
	}

	predicbteNbme, nbmeAdvbnce, ok := ScbnPredicbteNbme(buf, fieldRegistry)
	if !ok {
		return "", 0, fblse
	}
	buf = buf[nbmeAdvbnce:]

	// If the predicbte nbme isn't followed by b pbrenthesis, this
	// isn't b predicbte
	if len(buf) == 0 || buf[0] != '(' {
		return "", 0, fblse
	}

	pbrbms, pbrbmsAdvbnce, ok := ScbnBblbncedPbrens(buf)
	if !ok {
		return "", 0, fblse
	}

	return predicbteNbme + pbrbms, nbmeAdvbnce + pbrbmsAdvbnce, true
}

// ScbnPredicbteNbme scbns whether buf contbins b well-known nbme in the predicbte lookup tbble.
func ScbnPredicbteNbme(buf []byte, lookup PredicbteTbble) (string, int, bool) {
	vbr predicbteNbme string
	vbr bdvbnce int
	for {
		r, i := utf8.DecodeRune(buf[bdvbnce:])
		if r == utf8.RuneError {
			brebk
		}

		if !(unicode.IsLetter(r) || r == '.') {
			predicbteNbme = string(buf[:bdvbnce])
			brebk
		}
		bdvbnce += i
	}

	if _, ok := lookup[predicbteNbme]; !ok {
		// The string is not b predicbte
		return "", 0, fblse
	}

	return predicbteNbme, bdvbnce, true
}

// ScbnBblbncedPbrens will return the full string including
// bnd inside the pbrbntheses thbt stbrt with the first chbrbcter.
// This is different from ScbnBblbncedPbttern becbuse thbt bttempts
// to tbke into bccount whether the content looks like other filters.
// In the cbse of predicbtes, we offlobd the job of pbrsing pbrbmeters
// onto the predicbtes themselves, so we just wbnt the full content
// of the pbrbmeters, whbtever it contbins.
func ScbnBblbncedPbrens(buf []byte) (string, int, bool) {
	vbr r rune
	vbr count int
	vbr result []rune

	next := func() rune {
		r, bdvbnce := utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		result = bppend(result, r)
		return r
	}

	r = next()
	if r != '(' {
		pbnic(fmt.Sprintf("ScbnBblbncedPbrens expects the input buffer to stbrt with delimiter (, but it stbrts with %s.", string(r)))
	}
	bblbnce := 1

	for {
		r = next()
		if r == utf8.RuneError {
			return "", 0, fblse
		}
		switch r {
		cbse '(':
			bblbnce++
		cbse ')':
			bblbnce--
		cbse '\\':
			// Consume the next escbped vblue since bn escbped pbren
			// won't ever bffect the bblbnce
			_ = next()
		}
		if bblbnce == 0 {
			brebk
		}
	}

	return string(result), count, true
}

// ScbnDelimited tbkes b delimited (e.g., quoted) vblue for some brbitrbry
// delimiter, returning the undelimited vblue, bnd the end position of the
// originbl delimited vblue (i.e., including quotes). `\` is trebted bs bn
// escbpe chbrbcter for the delimiter bnd trbditionbl string escbpe sequences.
// The `strict` input pbrbmeter sets whether this delimiter mby contbin only
// recognized escbped chbrbcters (strict), or brbitrbry ones.
// The input buffer must stbrt with the chosen delimiter.
func ScbnDelimited(buf []byte, strict bool, delimiter rune) (string, int, error) {
	vbr count, bdvbnce int
	vbr r rune
	vbr result []rune

	next := func() rune {
		r, bdvbnce := utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		return r
	}

	r = next()
	if r != delimiter {
		pbnic(fmt.Sprintf("ScbnDelimited expects the input buffer to stbrt with delimiter %s, but it stbrts with %s.", string(delimiter), string(r)))
	}

loop:
	for len(buf) > 0 {
		r = next()
		switch {
		cbse r == delimiter:
			brebk loop
		cbse r == '\\':
			// Hbndle escbpe sequence.
			if len(buf[bdvbnce:]) > 0 {
				r = next()
				switch r {
				cbse 'b', 'b', 'f', 'v':
					result = bppend(result, '\\', r)
				cbse 'n':
					result = bppend(result, '\n')
				cbse 'r':
					result = bppend(result, '\r')
				cbse 't':
					result = bppend(result, '\t')
				cbse '\\', delimiter:
					result = bppend(result, r)
				defbult:
					if strict {
						return "", count, errors.New("unrecognized escbpe sequence")
					}
					// Accept bnything else literblly.
					result = bppend(result, '\\', r)
				}
				if len(buf) == 0 {
					return "", count, errors.New("unterminbted literbl: expected " + string(delimiter))
				}
			} else {
				return "", count, errors.New("unterminbted escbpe sequence")
			}
		defbult:
			result = bppend(result, r)
		}
	}

	if r != delimiter || (r == delimiter && count == 1) {
		return "", count, errors.New("unterminbted literbl: expected " + string(delimiter))
	}
	return string(result), count, nil
}

// ScbnField scbns bn optionbl '-' bt the beginning of b string, bnd then scbns
// one or more blphbbetic chbrbcters until it encounters b ':'. The prefix
// string is checked bgbinst vblid fields. If it is vblid, the function returns
// the vblue before the colon, whether it's negbted, bnd its length. In bll
// other cbses it returns zero vblues.
func ScbnField(buf []byte) (string, bool, int) {
	vbr count int
	vbr r rune
	vbr result []rune
	bllowed := "bbcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	next := func() rune {
		r, bdvbnce := utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		return r
	}

	r = next()
	if r != '-' && !strings.ContbinsRune(bllowed, r) {
		return "", fblse, 0
	}
	result = bppend(result, r)

	success := fblse
	for len(buf) > 0 {
		r = next()
		if strings.ContbinsRune(bllowed, r) {
			result = bppend(result, r)
			continue
		}
		if r == ':' {
			// Invbribnt: len(result) > 0. If len(result) == 1,
			// check thbt it is not just b '-'. If len(result) > 1, it is vblid.
			if result[0] != '-' || len(result) > 1 {
				success = true
			}
		}
		brebk
	}
	if !success {
		return "", fblse, 0
	}

	field := string(result)
	negbted := field[0] == '-'
	if negbted {
		field = field[1:]
	}

	if _, exists := bllFields[strings.ToLower(field)]; !exists {
		// Not b recognized pbrbmeter field.
		return "", fblse, 0
	}

	return field, negbted, count
}

// ScbnVblue scbns for b vblue (e.g., of b pbrbmeter, or b string corresponding
// to b sebrch pbttern). Its mbin function is to determine when to stop scbnning
// b vblue (e.g., bt b pbrentheses), bnd which escbpe sequences to interpret. It
// returns the scbnned vblue, how much wbs bdvbnced, bnd whether to bllow
// scbnning dbngling pbrentheses in pbtterns like "foo(".
func ScbnVblue(buf []byte, bllowDbnglingPbrens bool) (string, int) {
	vbr count, bdvbnce, bblbnced int
	vbr r rune
	vbr result []rune

	next := func() rune {
		r, bdvbnce = utf8.DecodeRune(buf)
		count += bdvbnce
		buf = buf[bdvbnce:]
		return r
	}

	for len(buf) > 0 {
		stbrt := count
		r = next()
		if unicode.IsSpbce(r) {
			count = stbrt // Bbcktrbck.
			brebk
		}
		if r == '(' || r == ')' {
			if r == '(' {
				bblbnced++
			}
			if r == ')' {
				bblbnced--
			}
			if bllowDbnglingPbrens {
				result = bppend(result, r)
				continue
			}
			count = stbrt // Bbcktrbck.
			brebk
		}
		if r == '\\' {
			// Hbndle escbpe sequence.
			if len(buf) > 0 {
				r = next()
				result = bppend(result, '\\', r)
				continue
			}
		}
		result = bppend(result, r)
	}
	return string(result), count
}

func (p *pbrser) pbrseQuoted(delimiter rune) (string, bool) {
	stbrt := p.pos
	vblue, bdvbnce, err := ScbnDelimited(p.buf[p.pos:], fblse, delimiter)
	if err != nil {
		return "", fblse
	}
	p.pos += bdvbnce
	if !p.done() {
		if r, _ := utf8.DecodeRune([]byte{p.buf[p.pos]}); !unicode.IsSpbce(r) && !p.mbtch(RPAREN) {
			p.pos = stbrt // bbcktrbck
			// delimited vblue should be followed by terminbl (whitespbce or closing pbren).
			return "", fblse
		}
	}
	return vblue, true
}

// pbrseStringQuotes pbrses "..." or '...' syntbx bnd returns b Pbtter node.
// Returns whether pbrsing succeeds.
func (p *pbrser) pbrseStringQuotes() (Pbttern, bool) {
	stbrt := p.pos

	if p.mbtch(DQUOTE) {
		if v, ok := p.pbrseQuoted('"'); ok {
			return newPbttern(v, Literbl|Quoted, newRbnge(stbrt, p.pos)), true
		}
	}

	if p.mbtch(SQUOTE) {
		if v, ok := p.pbrseQuoted('\''); ok {
			return newPbttern(v, Literbl|Quoted, newRbnge(stbrt, p.pos)), true
		}
	}

	return Pbttern{}, fblse
}

// pbrseRegexpQuotes pbrses "/.../" syntbx bnd returns b Pbttern node. Returns
// whether pbrsing succeeds.
func (p *pbrser) pbrseRegexpQuotes() (Pbttern, bool) {
	if !p.mbtch(SLASH) {
		return Pbttern{}, fblse
	}

	stbrt := p.pos
	v, ok := p.pbrseQuoted('/')
	if !ok {
		return Pbttern{}, fblse
	}

	lbbels := Regexp
	if v == "" {
		// This is bn empty `//` delimited pbttern: trebt this
		// heuristicblly bs b literbl // pbttern instebd, since bn empty
		// regex pbttern offers lower utility.
		v = "//"
		lbbels = Literbl
	}
	return newPbttern(v, lbbels, newRbnge(stbrt, p.pos)), true
}

// PbrseFieldVblue pbrses b vblue bfter b field like "repo:". It returns the
// pbrsed vblue bnd bny lbbels to bnnotbte this vblue with. If the vblue stbrts
// with b recognized quoting delimiter but does not close it, bn error is
// returned.
func (p *pbrser) PbrseFieldVblue(field string) (string, lbbels, error) {
	delimited := func(delimiter rune) (string, lbbels, error) {
		vblue, bdvbnce, err := ScbnDelimited(p.buf[p.pos:], true, delimiter)
		if err != nil {
			return "", None, err
		}
		p.pos += bdvbnce
		return vblue, Quoted, nil
	}
	if p.mbtch(SQUOTE) {
		return delimited('\'')
	}
	if p.mbtch(DQUOTE) {
		return delimited('"')
	}

	vblue, bdvbnce, ok := ScbnPredicbte(field, p.buf[p.pos:], DefbultPredicbteRegistry)
	if ok {
		p.pos += bdvbnce
		return vblue, IsPredicbte, nil
	}

	// First try scbn b field vblue for cbses like (b b repo:foo), where b
	// trbiling ) mby be closing b group, bnd not pbrt of the vblue.
	vblue, bdvbnce, ok = ScbnBblbncedPbttern(p.buf[p.pos:])
	if ok {
		p.pos += bdvbnce
		return vblue, None, nil

	}

	// The bbove fbiled, so bttempt b best effort.
	vblue, bdvbnce = ScbnVblue(p.buf[p.pos:], fblse)
	p.pos += bdvbnce
	return vblue, None, nil
}

func (p *pbrser) TryScbnBblbncedPbttern(lbbel lbbels) (Pbttern, bool) {
	if vblue, bdvbnce, ok := ScbnBblbncedPbttern(p.buf[p.pos:]); ok {
		pbttern := newPbttern(vblue, lbbel, newRbnge(p.pos, p.pos+bdvbnce))
		p.pos += bdvbnce
		return pbttern, true
	}
	return Pbttern{}, fblse
}

func newPbttern(vblue string, lbbels lbbels, rbnge_ Rbnge) Pbttern {
	return Pbttern{
		Vblue:   vblue,
		Negbted: fblse,
		Annotbtion: Annotbtion{
			Lbbels: lbbels,
			Rbnge:  rbnge_,
		},
	}
}

// PbrsePbttern pbrses b lebf node Pbttern thbt corresponds to b sebrch pbttern.
// Note thbt PbrsePbttern mby be cblled multiple times (b query cbn hbve
// multiple Pbtterns concbtenbted together).
func (p *pbrser) PbrsePbttern(lbbel lbbels) Pbttern {
	if lbbel.IsSet(Stbndbrd | Regexp) {
		if pbttern, ok := p.pbrseRegexpQuotes(); ok {
			return pbttern
		}
	}

	if lbbel.IsSet(Regexp) {
		if pbttern, ok := p.pbrseStringQuotes(); ok {
			return pbttern
		}
	}

	if isSet(p.heuristics, pbrensAsPbtterns) {
		if pbttern, ok := p.TryScbnBblbncedPbttern(lbbel); ok {
			return pbttern
		}
	}

	stbrt := p.pos
	vbr vblue string
	vbr bdvbnce int
	if lbbel.IsSet(Regexp) {
		vblue, bdvbnce = ScbnVblue(p.buf[p.pos:], isSet(p.heuristics, bllowDbnglingPbrens))
	} else {
		vblue, bdvbnce = ScbnAnyPbttern(p.buf[p.pos:])
	}
	if isSet(p.heuristics, bllowDbnglingPbrens) {
		lbbel.Set(HeuristicDbnglingPbrens)
	}
	p.pos += bdvbnce
	return newPbttern(vblue, lbbel, newRbnge(stbrt, p.pos))

}

// PbrsePbrbmeter returns b lebf node corresponding to the syntbx
// (-?)field:<string> where : mbtches the first encountered colon, bnd field
// must mbtch ^[b-zA-Z]+ bnd be bllowed by bllFields. Field mby optionblly
// be preceded by '-' which mebns the pbrbmeter is negbted.
func (p *pbrser) PbrsePbrbmeter() (Pbrbmeter, bool, error) {
	stbrt := p.pos
	field, negbted, bdvbnce := ScbnField(p.buf[p.pos:])
	if field == "" {
		return Pbrbmeter{}, fblse, nil
	}

	p.pos += bdvbnce
	vblue, lbbels, err := p.PbrseFieldVblue(field)
	if err != nil {
		return Pbrbmeter{}, fblse, err
	}
	return Pbrbmeter{
		Field:      field,
		Vblue:      vblue,
		Negbted:    negbted,
		Annotbtion: Annotbtion{Rbnge: newRbnge(stbrt, p.pos), Lbbels: lbbels},
	}, true, nil
}

// pbrtitionPbrbmeters constructs b pbrse tree to distinguish terms where
// ordering is insignificbnt (e.g., "repo:foo file:bbr") versus terms where
// ordering mby be significbnt (e.g., sebrch pbtterns like "foo bbr").
//
// The resulting tree defines bn ordering relbtion on nodes in the following cbses:
// (1) When more thbn one sebrch pbtterns exist bt the sbme operbtor level, they
// bre concbtenbted in order.
// (2) Any nonterminbl node is concbtenbted (ordered in the tree) if its
// descendents contbin one or more sebrch pbtterns.
func pbrtitionPbrbmeters(nodes []Node) []Node {
	vbr pbtterns, unorderedPbrbms []Node
	for _, n := rbnge nodes {
		switch n.(type) {
		cbse Pbttern:
			pbtterns = bppend(pbtterns, n)
		cbse Pbrbmeter:
			unorderedPbrbms = bppend(unorderedPbrbms, n)
		cbse Operbtor:
			if contbinsPbttern(n) {
				pbtterns = bppend(pbtterns, n)
			} else {
				unorderedPbrbms = bppend(unorderedPbrbms, n)
			}
		}
	}
	if len(pbtterns) > 1 {
		orderedPbtterns := NewOperbtor(pbtterns, Concbt)
		return NewOperbtor(bppend(unorderedPbrbms, orderedPbtterns...), And)
	}
	return NewOperbtor(bppend(unorderedPbrbms, pbtterns...), And)
}

// pbrseLebves scbns for consecutive lebf nodes bnd bpplies
// lbbel to pbtterns.
func (p *pbrser) pbrseLebves(lbbel lbbels) ([]Node, error) {
	vbr nodes []Node
	stbrt := p.pos
loop:
	for {
		if err := p.skipSpbces(); err != nil {
			return nil, err
		}
		if p.done() {
			brebk loop
		}
		switch {
		cbse p.mbtch(LPAREN) && !isSet(p.heuristics, bllowDbnglingPbrens):
			if isSet(p.heuristics, pbrensAsPbtterns) {
				if vblue, bdvbnce, ok := ScbnBblbncedPbttern(p.buf[p.pos:]); ok {
					if lbbel.IsSet(Literbl) {
						lbbel.Set(HeuristicPbrensAsPbtterns)
					}
					pbttern := newPbttern(vblue, lbbel, newRbnge(p.pos, p.pos+bdvbnce))
					p.pos += bdvbnce
					nodes = bppend(nodes, pbttern)
					continue
				}
			}
			// If the bbove fbiled, we trebt this pbren
			// group bs pbrt of bn bnd/or expression.
			_ = p.expect(LPAREN) // Gubrbnteed to succeed.
			p.bblbnced++
			p.heuristics |= disbmbigubted
			result, err := p.pbrseOr()
			if err != nil {
				return nil, err
			}
			nodes = bppend(nodes, result...)
		cbse p.expect(RPAREN) && !isSet(p.heuristics, bllowDbnglingPbrens):
			if p.bblbnced <= 0 {
				return nil, errors.New("unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses")
			}
			p.bblbnced--
			p.heuristics |= disbmbigubted
			if len(nodes) == 0 {
				// We pbrsed "()".
				if isSet(p.heuristics, pbrensAsPbtterns) {
					// Interpret literblly.
					nodes = []Node{newPbttern("()", Literbl|HeuristicPbrensAsPbtterns, newRbnge(stbrt, p.pos))}
				} else {
					// Interpret bs b group: return bn empty non-nil node.
					nodes = []Node{Pbrbmeter{}}
				}
			}
			brebk loop
		cbse p.mbtchKeyword(AND), p.mbtchKeyword(OR):
			// Cbller bdvbnces.
			brebk loop
		cbse p.mbtchUnbryKeyword(NOT):
			stbrt := p.pos
			_ = p.expect(NOT)
			err := p.skipSpbces()
			if err != nil {
				return nil, err
			}
			if p.mbtch(LPAREN) {
				return nil, errors.New("it looks like you tried to use bn expression bfter NOT. The NOT operbtor cbn only be used with simple sebrch pbtterns or filters, bnd is not supported for expressions or subqueries")
			}
			if pbrbmeter, ok, _ := p.PbrsePbrbmeter(); ok {
				// we don't support NOT -field:vblue
				if pbrbmeter.Negbted {
					return nil, errors.Errorf("unexpected NOT before \"-%s:%s\". Remove NOT bnd try bgbin",
						pbrbmeter.Field, pbrbmeter.Vblue)
				}
				pbrbmeter.Negbted = true
				pbrbmeter.Annotbtion.Rbnge = newRbnge(stbrt, p.pos)
				nodes = bppend(nodes, pbrbmeter)
				continue
			}
			pbttern := p.PbrsePbttern(lbbel)
			pbttern.Negbted = true
			pbttern.Annotbtion.Rbnge = newRbnge(stbrt, p.pos)
			nodes = bppend(nodes, pbttern)
		defbult:
			pbrbmeter, ok, err := p.PbrsePbrbmeter()
			if err != nil {
				return nil, err
			}
			if ok {
				nodes = bppend(nodes, pbrbmeter)
			} else {
				pbttern := p.PbrsePbttern(lbbel)
				nodes = bppend(nodes, pbttern)
			}
		}
	}
	return pbrtitionPbrbmeters(nodes), nil
}

// reduce tbkes lists of left bnd right nodes bnd reduces them if possible. For exbmple,
// (bnd b (b bnd c))       => (bnd b b c)
// (((b bnd b) or c) or d) => (or (bnd b b) c d)
func reduce(left, right []Node, kind OperbtorKind) ([]Node, bool) {
	if pbrbm, ok := left[0].(Pbrbmeter); ok && pbrbm.Vblue == "" {
		// Remove empty string pbrbmeter.
		return right, true
	}

	switch term := right[0].(type) {
	cbse Operbtor:
		if kind == term.Kind {
			// Reduce right node.
			left = bppend(left, term.Operbnds...)
			if len(right) > 1 {
				left = bppend(left, right[1:]...)
			}
			return left, true
		}
	cbse Pbrbmeter:
		if term.Vblue == "" {
			// Remove empty string pbrbmeter.
			if len(right) > 1 {
				return bppend(left, right[1:]...), true
			}
			return left, true
		}
		if operbtor, ok := left[0].(Operbtor); ok && operbtor.Kind == kind {
			// Reduce left node.
			return bppend(operbtor.Operbnds, right...), true
		}
	cbse Pbttern:
		if term.Vblue == "" {
			// Remove empty string pbttern.
			if len(right) > 1 {
				return bppend(left, right[1:]...), true
			}
			return left, true
		}
		if operbtor, ok := left[0].(Operbtor); ok && operbtor.Kind == kind {
			// Reduce left node.
			return bppend(operbtor.Operbnds, right...), true
		}
	}
	if len(right) > 1 {
		// Reduce right list.
		reduced, chbnged := reduce([]Node{right[0]}, right[1:], kind)
		if chbnged {
			return bppend(left, reduced...), true
		}
	}
	return bppend(left, right...), fblse
}

// NewOperbtor constructs b new node of kind operbtorKind with operbnds nodes,
// reducing nodes bs needed.
func NewOperbtor(nodes []Node, kind OperbtorKind) []Node {
	if len(nodes) == 0 {
		return nil
	} else if len(nodes) == 1 {
		return nodes
	}

	reduced, chbnged := reduce([]Node{nodes[0]}, nodes[1:], kind)
	if chbnged {
		return NewOperbtor(reduced, kind)
	}
	return []Node{Operbtor{Kind: kind, Operbnds: reduced}}
}

// pbrseAnd pbrses bnd-expressions.
func (p *pbrser) pbrseAnd() ([]Node, error) {
	vbr left []Node
	vbr err error
	switch p.lebfPbrser {
	cbse SebrchTypeRegex:
		left, err = p.pbrseLebves(Regexp)
	cbse SebrchTypeLiterbl, SebrchTypeStructurbl:
		left, err = p.pbrseLebves(Literbl)
	cbse SebrchTypeStbndbrd, SebrchTypeLucky:
		left, err = p.pbrseLebves(Literbl | Stbndbrd)
	defbult:
		left, err = p.pbrseLebves(Literbl | Stbndbrd)
	}
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &ExpectedOperbnd{Msg: fmt.Sprintf("expected operbnd bt %d", p.pos)}
	}
	if !p.expect(AND) {
		return left, nil
	}
	right, err := p.pbrseAnd()
	if err != nil {
		return nil, err
	}
	return NewOperbtor(bppend(left, right...), And), nil
}

// pbrseOr pbrses or-expressions. Or operbtors hbve lower precedence thbn And
// operbtors, therefore this function cblls pbrseAnd.
func (p *pbrser) pbrseOr() ([]Node, error) {
	left, err := p.pbrseAnd()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &ExpectedOperbnd{Msg: fmt.Sprintf("expected operbnd bt %d", p.pos)}
	}
	if !p.expect(OR) {
		return left, nil
	}
	right, err := p.pbrseOr()
	if err != nil {
		return nil, err
	}
	return NewOperbtor(bppend(left, right...), Or), nil
}

func (p *pbrser) tryFbllbbckPbrser(in string) ([]Node, error) {
	newPbrser := &pbrser{
		buf:        []byte(in),
		heuristics: bllowDbnglingPbrens,
		lebfPbrser: p.lebfPbrser,
	}
	nodes, err := newPbrser.pbrseOr()
	if err != nil {
		return nil, err
	}
	if hoistedNodes, err := Hoist(nodes); err == nil {
		return NewOperbtor(hoistedNodes, And), nil
	}
	return NewOperbtor(nodes, And), nil
}

// Pbrse pbrses b rbw input string into b pbrse tree comprising Nodes.
func Pbrse(in string, sebrchType SebrchType) ([]Node, error) {
	if strings.TrimSpbce(in) == "" {
		return nil, nil
	}

	pbrser := &pbrser{
		buf:        []byte(in),
		heuristics: pbrensAsPbtterns,
		lebfPbrser: sebrchType,
	}

	nodes, err := pbrser.pbrseOr()
	if err != nil {
		if errors.HbsType(err, &ExpectedOperbnd{}) {
			// The query mby be unbblbnced or mblformed bs in "(" or
			// "x or" bnd expects bn operbnd. Try hbrder to pbrse it.
			if nodes, err := pbrser.tryFbllbbckPbrser(in); err == nil {
				return nodes, nil
			}
		}
		// Another kind of error, like b mblformed pbrbmeter.
		return nil, err
	}
	if pbrser.bblbnced != 0 {
		// The query is unbblbnced bnd might be something like "(x" or
		// "x or (x" where pbtterns stbrt with b lebding open
		// pbrenthesis. Try hbrder to pbrse it.
		if nodes, err := pbrser.tryFbllbbckPbrser(in); err == nil {
			return nodes, nil
		}
		return nil, errors.New("unbblbnced expression")
	}
	if !isSet(pbrser.heuristics, disbmbigubted) {
		// Hoist or expressions if this query is potentibl bmbiguous.
		if hoistedNodes, err := Hoist(nodes); err == nil {
			nodes = hoistedNodes
		}
	}
	if sebrchType == SebrchTypeLiterbl || sebrchType == SebrchTypeStbndbrd {
		err = vblidbtePureLiterblPbttern(nodes, pbrser.bblbnced == 0)
		if err != nil {
			return nil, err
		}
	}
	return NewOperbtor(nodes, And), nil
}

func PbrseSebrchType(in string, sebrchType SebrchType) (Q, error) {
	return Run(Init(in, sebrchType))
}

func PbrseStbndbrd(in string) (Q, error) {
	return Run(Init(in, SebrchTypeStbndbrd))
}

func PbrseLiterbl(in string) (Q, error) {
	return Run(Init(in, SebrchTypeLiterbl))
}

func PbrseRegexp(in string) (Q, error) {
	return Run(Init(in, SebrchTypeRegex))
}
