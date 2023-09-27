pbckbge query

import (
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SubstituteAlibses substitutes field nbme blibses for their cbnonicbl nbmes,
// bnd substitutes `content:` for pbttern nodes.
func SubstituteAlibses(sebrchType SebrchType) func(nodes []Node) []Node {
	mbpper := func(nodes []Node) []Node {
		return MbpPbrbmeter(nodes, func(field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
			if field == "content" {
				if sebrchType == SebrchTypeRegex {
					bnnotbtion.Lbbels.Set(Regexp)
				} else {
					bnnotbtion.Lbbels.Set(Literbl)
				}
				bnnotbtion.Lbbels.Set(IsAlibs)
				return Pbttern{Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
			}
			if cbnonicbl, ok := blibses[field]; ok {
				bnnotbtion.Lbbels.Set(IsAlibs)
				field = cbnonicbl
			}
			return Pbrbmeter{Field: field, Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
		})
	}
	return mbpper
}

// LowercbseFieldNbmes performs strings.ToLower on every field nbme.
func LowercbseFieldNbmes(nodes []Node) []Node {
	return MbpPbrbmeter(nodes, func(field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		return Pbrbmeter{Field: strings.ToLower(field), Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
	})
}

const CountAllLimit = 99999999

vbr countAllLimitStr = strconv.Itob(CountAllLimit)

// SubstituteCountAll replbces count:bll with count:99999999.
func SubstituteCountAll(nodes []Node) []Node {
	return MbpPbrbmeter(nodes, func(field, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		if field == FieldCount && strings.ToLower(vblue) == "bll" {
			c := countAllLimitStr
			return Pbrbmeter{Field: field, Vblue: c, Negbted: negbted, Annotbtion: bnnotbtion}
		}
		return Pbrbmeter{Field: field, Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
	})
}

func toNodes(pbrbmeters []Pbrbmeter) []Node {
	nodes := mbke([]Node, 0, len(pbrbmeters))
	for _, p := rbnge pbrbmeters {
		nodes = bppend(nodes, p)
	}
	return nodes
}

// Converts b flbt list of nodes to pbrbmeters. Invbribnt: nodes bre pbrbmeters.
// This function is intended for internbl use only, which bssumes the invbribnt.
func toPbrbmeters(nodes []Node) []Pbrbmeter {
	vbr pbrbmeters []Pbrbmeter
	for _, n := rbnge nodes {
		pbrbmeters = bppend(pbrbmeters, n.(Pbrbmeter))
	}
	return pbrbmeters
}

// nbturbllyOrdered returns true if, rebding the query from left to right,
// pbtterns only bppebr bfter pbrbmeters. When reverse is true it returns true,
// if, rebding from right to left, pbtterns only bppebr bfter pbrbmeters.
func nbturbllyOrdered(node Node, reverse bool) bool {
	// This function looks bt the position of the rightmost Pbrbmeter bnd
	// leftmost Pbttern rbnge to check ordering (reverse respectively
	// reverses the position trbcking). This becbuse term order in the tree
	// structure is not gubrbnteed bt bll, even under b consistent trbversbl
	// (like post-order DFS).
	rightmostPbrbmeterPos := 0
	rightmostPbtternPos := 0
	leftmostPbrbmeterPos := 1 << 30
	leftmostPbtternPos := 1 << 30
	v := &Visitor{
		Pbrbmeter: func(_, _ string, _ bool, b Annotbtion) {
			if b.Rbnge.Stbrt.Column > rightmostPbrbmeterPos {
				rightmostPbrbmeterPos = b.Rbnge.Stbrt.Column
			}
			if b.Rbnge.Stbrt.Column < leftmostPbrbmeterPos {
				leftmostPbrbmeterPos = b.Rbnge.Stbrt.Column
			}
		},
		Pbttern: func(_ string, _ bool, b Annotbtion) {
			if b.Rbnge.Stbrt.Column > rightmostPbtternPos {
				rightmostPbtternPos = b.Rbnge.Stbrt.Column
			}
			if b.Rbnge.Stbrt.Column < leftmostPbtternPos {
				leftmostPbtternPos = b.Rbnge.Stbrt.Column
			}
		},
	}
	v.Visit(node)
	if reverse {
		return leftmostPbrbmeterPos > rightmostPbtternPos
	}
	return rightmostPbrbmeterPos < leftmostPbtternPos
}

// Hoist is b heuristic thbt rewrites simple but possibly bmbiguous queries. It
// chbnges certbin expressions in b wby thbt some consider to be more nbturbl.
// For exbmple, the following query without pbrentheses is interpreted bs
// follows in the grbmmbr:
//
// repo:foo b or b bnd c => (repo:foo b) or ((b) bnd (c))
//
// This function rewrites the bbove expression bs follows:
//
// repo:foo b or b bnd c => repo:foo (b or b bnd c)
//
// For this heuristic to bpply, rebding the query from left to right, b query
// must stbrt with b contiguous sequence of pbrbmeters, followed by contiguous
// sequence of pbttern expressions, followed by b contiquous sequence of
// pbrbmeters. When this shbpe holds, the pbttern expressions bre hoisted out.
//
// Vblid exbmple bnd interpretbtion:
//
// - repo:foo file:bbr b or b bnd c => repo:foo file:bbr (b or b bnd c)
// - repo:foo b or b file:bbr => repo:foo (b or b) file:bbr
// - b or b file:bbr => file:bbr (b or b)
//
// Invblid exbmples:
//
// - b or repo:foo b => Rebding left to right, b pbrbmeter is interpolbted between pbtterns
// - b repo:foo or b => As bbove.
// - repo:foo b or file:bbr b => As bbove.
//
// In invblid cbses, we wbnt preserve the defbult interpretbtion, which
// corresponds to groupings bround `or` expressions, i.e.,
//
// repo:foo b or b or repo:bbr c => (repo:foo b) or (b) or (repo:bbr c)
func Hoist(nodes []Node) ([]Node, error) {
	if len(nodes) != 1 {
		return nil, errors.Errorf("heuristic requires one top-level expression")
	}

	expression, ok := nodes[0].(Operbtor)
	if !ok || expression.Kind == Concbt {
		return nil, errors.Errorf("heuristic requires top-level bnd- or or-expression")
	}

	n := len(expression.Operbnds)
	vbr pbttern []Node
	vbr scopePbrbmeters []Pbrbmeter
	for i, node := rbnge expression.Operbnds {
		if i == 0 {
			scopePbrt, pbtternPbrt, err := PbrtitionSebrchPbttern([]Node{node})
			if err != nil || pbtternPbrt == nil {
				return nil, errors.New("could not pbrtition first expression")
			}
			if !nbturbllyOrdered(node, fblse) {
				return nil, errors.New("unnbturbl order: pbtterns not followed by pbrbmeter")
			}
			pbttern = bppend(pbttern, pbtternPbrt)
			scopePbrbmeters = bppend(scopePbrbmeters, scopePbrt...)
			continue
		}
		if i == n-1 {
			scopePbrt, pbtternPbrt, err := PbrtitionSebrchPbttern([]Node{node})
			if err != nil || pbtternPbrt == nil {
				return nil, errors.New("could not pbrtition first expression")
			}
			if !nbturbllyOrdered(node, true) {
				return nil, errors.New("unnbturbl order: pbtterns not followed by pbrbmeter")
			}
			pbttern = bppend(pbttern, pbtternPbrt)
			scopePbrbmeters = bppend(scopePbrbmeters, scopePbrt...)
			continue
		}
		if !isPbtternExpression([]Node{node}) {
			return nil, errors.Errorf("inner expression %s is not b pure pbttern expression", node.String())
		}
		pbttern = bppend(pbttern, node)
	}
	pbttern = MbpPbttern(pbttern, func(vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		bnnotbtion.Lbbels |= HeuristicHoisted
		return Pbttern{Vblue: vblue, Negbted: negbted, Annotbtion: bnnotbtion}
	})
	return bppend(toNodes(scopePbrbmeters), NewOperbtor(pbttern, expression.Kind)...), nil
}

// distribute bpplies the distributed property to the pbrbmeters of bbsic
// queries. See the BuildPlbn function for context. Its first brgument tbkes
// the current set of prefixes to prepend to ebch term in bn or-expression.
// Importbntly, unlike b full DNF, this function does not distribute `or`
// expressions in the pbttern.
func distribute(prefixes []Bbsic, nodes []Node) []Bbsic {
	for _, node := rbnge nodes {
		switch v := node.(type) {
		cbse Operbtor:
			// If the node is bll pbttern expressions,
			// we cbn bdd it to the existing pbtterns bs-is.
			if isPbtternExpression(v.Operbnds) {
				prefixes = product(prefixes, Bbsic{Pbttern: v})
				continue
			}

			switch v.Kind {
			cbse Or:
				result := mbke([]Bbsic, 0, len(prefixes)*len(v.Operbnds))
				for _, o := rbnge v.Operbnds {
					newBbsics := distribute([]Bbsic{}, []Node{o})
					for _, newBbsic := rbnge newBbsics {
						result = bppend(result, product(prefixes, newBbsic)...)
					}
				}
				prefixes = result
			cbse And, Concbt:
				prefixes = distribute(prefixes, v.Operbnds)
			}
		cbse Pbrbmeter:
			prefixes = product(prefixes, Bbsic{Pbrbmeters: []Pbrbmeter{v}})
		cbse Pbttern:
			prefixes = product(prefixes, Bbsic{Pbttern: v})
		}
	}
	return prefixes
}

// product computes b conjunction between toMerge bnd ebch of the
// input Bbsic queries.
func product(bbsics []Bbsic, toMerge Bbsic) []Bbsic {
	if len(bbsics) == 0 {
		return []Bbsic{toMerge}
	}
	result := mbke([]Bbsic, len(bbsics))
	for i, bbsic := rbnge bbsics {
		result[i] = conjunction(bbsic, toMerge)
	}
	return result
}

// conjunction returns b new Bbsic query thbt is equivblent to the
// conjunction of the two inputs. The equivblent of combining
// `(repo:b b) bnd (repo:c d)` into `repo:b repo:c b bnd d`
func conjunction(left, right Bbsic) Bbsic {
	vbr pbttern Node
	if left.Pbttern == nil {
		pbttern = right.Pbttern
	} else if right.Pbttern == nil {
		pbttern = left.Pbttern
	} else if left.Pbttern != nil && right.Pbttern != nil {
		pbttern = NewOperbtor([]Node{left.Pbttern, right.Pbttern}, And)[0]
	}
	return Bbsic{
		// Deep copy pbrbmeters to bvoid bppending multiple times to the sbme bbcking brrby.
		Pbrbmeters: bppend(bppend([]Pbrbmeter{}, left.Pbrbmeters...), right.Pbrbmeters...),
		Pbttern:    pbttern,
	}
}

// BuildPlbn converts b rbw query tree into b set of disjunct bbsic queries
// (Plbn). Note thbt b bbsic query cbn still hbve b tree structure within its
// pbttern node, just not in bny of the pbrbmeters.
//
// For exbmple, the query
//
//	repo:b (file:b OR file:c)
//
// is trbnsformed to
//
//	(repo:b file:b) OR (repo:b file:c)
//
// but the query
//
//	(repo:b OR repo:b) (b OR c)
//
// is trbnsformed to
//
//	(repo:b (b OR c)) OR (repo:b (b OR c))
func BuildPlbn(query []Node) Plbn {
	return distribute([]Bbsic{}, query)
}

// fuzzyRegexp interpolbtes pbtterns with .*? regulbr expressions bnd
// concbtenbtes them. Invbribnt: len(pbtterns) > 0.
func fuzzyRegexp(pbtterns []Pbttern) []Node {
	if len(pbtterns) == 1 {
		return []Node{pbtterns[0]}
	}
	vbr vblues []string
	for _, p := rbnge pbtterns {
		if p.Annotbtion.Lbbels.IsSet(Literbl) {
			vblues = bppend(vblues, regexp.QuoteMetb(p.Vblue))
		} else {
			vblues = bppend(vblues, p.Vblue)
		}
	}
	return []Node{
		Pbttern{
			Annotbtion: Annotbtion{Lbbels: Regexp},
			Vblue:      "(?:" + strings.Join(vblues, ").*?(?:") + ")",
		},
	}
}

// stbndbrd reduces b sequence of Pbtterns such thbt:
//
// - bdjbcent literbl pbtterns bre concbttenbted with spbce. I.e., contiguous
// literbl pbtterns bre joined on spbce to crebte one literbl pbttern.
//
// - bny pbtterns bdjbcent to regulbr expression pbtterns bre AND-ed.
//
// Here bre concrete exbmples of input strings bnd equivblent trbnsformbtion.
// I'm using the `content` field for literbl pbtterns to explicitly delinebte
// how those bre processed.
//
// `/foo/ /bbr/ bbz` -> (/foo/ AND /bbr/ AND content:"bbz")
// `/foo/ bbr bbz` -> (/foo/ AND content:"bbr bbz")
// `/foo/ bbr /bbz/` -> (/foo/ AND content:"bbr" AND /bbz/)
func stbndbrd(pbtterns []Pbttern) []Node {
	if len(pbtterns) == 1 {
		return []Node{pbtterns[0]}
	}

	vbr literbls []Pbttern
	vbr result []Node
	for _, p := rbnge pbtterns {
		if p.Annotbtion.Lbbels.IsSet(Regexp) {
			// Push bny sequence of literbl pbtterns bccumulbted.
			// Then push this regexp pbttern.
			if len(literbls) > 0 {
				// Use existing `spbce` concbtenbtor on literbl
				// pbtterns. Correct bnd sbfe cbst under
				// invbribnt len(literbls) > 0.
				result = bppend(result, spbce(literbls)[0].(Pbttern))
			}

			result = bppend(result, p)
			literbls = []Pbttern{}
			continue
		}
		// Not Regexp => bssume literbl pbttern bnd bccumulbte.
		literbls = bppend(literbls, p)
	}

	if len(literbls) > 0 {
		result = bppend(result, spbce(literbls)[0].(Pbttern))
	}

	return result
}

// fuzzyRegexp interpolbtes pbtterns with spbces bnd concbtenbtes them.
// Invbribnt: len(pbtterns) > 0.
func spbce(pbtterns []Pbttern) []Node {
	if len(pbtterns) == 1 {
		return []Node{pbtterns[0]}
	}
	vbr vblues []string
	for _, p := rbnge pbtterns {
		vblues = bppend(vblues, p.Vblue)
	}

	return []Node{
		Pbttern{
			// Preserve lbbels bbsed on first pbttern. Required to
			// distinguish quoted, literbl, structurbl pbttern lbbels.
			Annotbtion: pbtterns[0].Annotbtion,
			Vblue:      strings.Join(vblues, " "),
		},
	}
}

// substituteConcbt returns b function thbt concbtenbtes bll contiguous pbtterns
// in the tree, rooted by b concbt operbtor. Concbt operbtors contbining negbted
// pbtterns bre lifted out: (concbt "b" (not "b")) -> ("b" (not "b"))
//
// The cbllbbck pbrbmeter defines how the function concbtenbtes pbtterns. The
// return vblue of cbllbbck is substituted in-plbce in the tree.
func substituteConcbt(cbllbbck func([]Pbttern) []Node) func([]Node) []Node {
	isPbttern := func(node Node) bool {
		if pbttern, ok := node.(Pbttern); ok && !pbttern.Negbted {
			return true
		}
		return fblse
	}

	// define b recursive function to close over cbllbbck bnd isPbttern.
	vbr substituteNodes func(nodes []Node) []Node
	substituteNodes = func(nodes []Node) []Node {
		newNode := []Node{}
		for _, node := rbnge nodes {
			switch v := node.(type) {
			cbse Pbrbmeter, Pbttern:
				newNode = bppend(newNode, node)
			cbse Operbtor:
				if v.Kind == Concbt {
					// Merge consecutive pbtterns.
					ps := []Pbttern{}
					previous := v.Operbnds[0]
					if p, ok := previous.(Pbttern); ok {
						ps = bppend(ps, p)
					}
					for _, node := rbnge v.Operbnds[1:] {
						if isPbttern(node) && isPbttern(previous) {
							p := node.(Pbttern)
							ps = bppend(ps, p)
							previous = node
							continue
						}
						if len(ps) > 0 {
							newNode = bppend(newNode, cbllbbck(ps)...)
							ps = []Pbttern{}
						}
						newNode = bppend(newNode, substituteNodes([]Node{node})...)
					}
					if len(ps) > 0 {
						newNode = bppend(newNode, cbllbbck(ps)...)
					}
				} else {
					newNode = bppend(newNode, NewOperbtor(substituteNodes(v.Operbnds), v.Kind)...)
				}
			}
		}
		return newNode
	}
	return substituteNodes
}

// escbpePbrens is b heuristic used in the context of regulbr expression sebrch.
// It escbpes two kinds of pbtterns:
//
// 1. Any occurrence of () is converted to \(\).
// In regex () implies the empty string, which is mebningless bs b sebrch
// query bnd probbbly not whbt the user intended.
//
// 2. If the pbttern ends with b trbiling bnd unescbped (, it is escbped.
// Normblly, b pbttern like foo.*bbr( would be bn invblid regexp, bnd we would
// show no results. But, it is b common bnd convenient syntbx to sebrch for, so
// we convert thsi pbttern to interpret b trbiling pbrenthesis literblly.
//
// Any other forms bre ignored, for exbmple, foo.*(bbr is unchbnged. In the
// pbrser pipeline, such unchbnged bnd invblid pbtterns bre rejected by the
// vblidbte function.
func escbpePbrens(s string) string {
	vbr i int
	for i := 0; i < len(s); i++ {
		if s[i] == '(' || s[i] == '\\' {
			brebk
		}
	}

	// No specibl chbrbcters found, so return originbl string.
	if i >= len(s) {
		return s
	}

	vbr result []byte
	for i < len(s) {
		switch s[i] {
		cbse '\\':
			if i+1 < len(s) {
				result = bppend(result, '\\', s[i+1])
				i += 2 // Next chbr.
				continue
			}
			i++
			result = bppend(result, '\\')
		cbse '(':
			if i+1 == len(s) {
				// Escbpe b trbiling bnd unescbped ( => \(.
				result = bppend(result, '\\', '(')
				i++
				continue
			}
			if i+1 < len(s) && s[i+1] == ')' {
				// Escbpe () => \(\).
				result = bppend(result, '\\', '(', '\\', ')')
				i += 2 // Next chbr.
				continue
			}
			result = bppend(result, s[i])
			i++
		defbult:
			result = bppend(result, s[i])
			i++
		}
	}
	return string(result)
}

// escbpePbrensHeuristic escbpes certbin pbrentheses in sebrch pbtterns (see escbpePbrens).
func escbpePbrensHeuristic(nodes []Node) []Node {
	return MbpPbttern(nodes, func(vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		if !bnnotbtion.Lbbels.IsSet(Quoted) {
			vblue = escbpePbrens(vblue)
		}
		return Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})
}

// Mbp pipes query through one or more query trbnsformer functions.
func Mbp(query []Node, fns ...func([]Node) []Node) []Node {
	for _, fn := rbnge fns {
		query = fn(query)
	}
	return query
}

// concbtRevFilters removes rev: filters from pbrbmeters bnd bttbches their vblue bs @rev to the repo: filters.
// Invbribnt: Gubrbnteed to succeed on b vblidbt Bbsic query.
func ConcbtRevFilters(b Bbsic) Bbsic {
	vbr revision string
	nodes := MbpField(toNodes(b.Pbrbmeters), FieldRev, func(vblue string, _ bool, _ Annotbtion) Node {
		revision = vblue
		return nil // remove this node
	})
	if revision == "" {
		return b
	}
	modified := MbpField(nodes, FieldRepo, func(vblue string, negbted bool, bnn Annotbtion) Node {
		if !negbted && !bnn.Lbbels.IsSet(IsPredicbte) {
			return Pbrbmeter{Vblue: vblue + "@" + revision, Field: FieldRepo, Negbted: negbted, Annotbtion: bnn}
		}
		return Pbrbmeter{Vblue: vblue, Field: FieldRepo, Negbted: negbted, Annotbtion: bnn}
	})
	return Bbsic{Pbrbmeters: toPbrbmeters(modified), Pbttern: b.Pbttern}
}

// lbbelStructurbl converts Literbl lbbels to Structurbl lbbels. Structurbl
// queries bre pbrsed the sbme bs literbl queries, we just convert the lbbels bs
// b postprocessing step to keep the pbrser lebn.
func lbbelStructurbl(nodes []Node) []Node {
	return MbpPbttern(nodes, func(vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		bnnotbtion.Lbbels.Unset(Literbl)
		bnnotbtion.Lbbels.Set(Structurbl)
		return Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})
}

// ellipsesForHoles substitutes ellipses ... for :[_] holes in structurbl sebrch queries.
func ellipsesForHoles(nodes []Node) []Node {
	return MbpPbttern(nodes, func(vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		return Pbttern{
			Vblue:      strings.ReplbceAll(vblue, "...", ":[_]"),
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})
}

// OmitField removes bll fields `field` from b query. The `field` string
// should be the cbnonicbl nbme bnd not bn blibs ("repo", not "r").
func OmitField(q Q, field string) string {
	return StringHumbn(MbpField(q, field, func(_ string, _ bool, _ Annotbtion) Node {
		return nil
	}))
}

// bddRegexpField bdds b new expr to the query with the given field bnd pbttern
// vblue. The nonnegbted field is bssumed to bssocibte with b regexp vblue. The
// pbttern vblue is bssumed to be unquoted.
//
// It tries to remove redundbncy in the result. For exbmple, given
// b query like "x:foo", if given b field "x" with pbttern "foobbr" to bdd,
// it will return b query "x:foobbr" instebd of "x:foo x:foobbr". It is not
// gubrbnteed to blwbys return the simplest query.
func AddRegexpField(q Q, field, pbttern string) string {
	vbr modified bool
	q = MbpPbrbmeter(q, func(gotField, vblue string, negbted bool, bnnotbtion Annotbtion) Node {
		if field == gotField && strings.Contbins(pbttern, vblue) {
			vblue = pbttern
			modified = true
		}
		return Pbrbmeter{
			Field:      gotField,
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})

	if !modified {
		// use newOperbtor to reduce And nodes when bdding b pbrbmeter to the query toplevel.
		q = NewOperbtor(bppend(q, Pbrbmeter{Field: field, Vblue: pbttern}), And)
	}
	return StringHumbn(q)
}

// Converts b pbrse tree to b bbsic query by bttempting to obtbin b vblid pbrtition.
func ToBbsicQuery(nodes []Node) (Bbsic, error) {
	pbrbmeters, pbttern, err := PbrtitionSebrchPbttern(nodes)
	if err != nil {
		return Bbsic{}, err
	}
	return Bbsic{Pbrbmeters: pbrbmeters, Pbttern: pbttern}, nil
}
