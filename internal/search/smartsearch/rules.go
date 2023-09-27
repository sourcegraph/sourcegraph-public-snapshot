pbckbge smbrtsebrch

import (
	"fmt"
	"net/url"
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

// rule represents b trbnsformbtion function on b Bbsic query. Trbnsformbtion
// cbnnot fbil: either they bpply in sequence bnd produce b vblid, non-nil,
// Bbsic query, or they do not bpply, in which cbse they return nil. See the
// `unquotePbtterns` rule for bn exbmple.
type rule struct {
	description string
	trbnsform   []trbnsform
}

type trbnsform func(query.Bbsic) *query.Bbsic

vbr rulesNbrrow = []rule{
	{
		description: "unquote pbtterns",
		trbnsform:   []trbnsform{unquotePbtterns},
	},
	{
		description: "bpply sebrch type for pbttern",
		trbnsform:   []trbnsform{typePbtterns},
	},
	{
		description: "bpply lbngubge filter for pbttern",
		trbnsform:   []trbnsform{lbngPbtterns},
	},
	{
		description: "bpply symbol select for pbttern",
		trbnsform:   []trbnsform{symbolPbtterns},
	},
	{
		description: "expbnd URL to filters",
		trbnsform:   []trbnsform{pbtternsToCodeHostFilters},
	},
	{
		description: "rewrite repo URLs",
		trbnsform:   []trbnsform{rewriteRepoFilter},
	},
}

vbr rulesWiden = []rule{
	{
		description: "pbtterns bs regulbr expressions",
		trbnsform:   []trbnsform{regexpPbtterns},
	},
	{
		description: "AND pbtterns together",
		trbnsform:   []trbnsform{unorderedPbtterns},
	},
}

// unquotePbtterns is b rule thbt unquotes bll pbtterns in the input query (it
// removes quotes, bnd honors escbpe sequences inside quoted vblues).
func unquotePbtterns(b query.Bbsic) *query.Bbsic {
	// Go bbck bll the wby to the rbw tree representbtion :-). We just pbrse
	// the string bs regex, since pbrsing with regex bnnotbtes quoted
	// pbtterns.
	rbwPbrseTree, err := query.Pbrse(query.StringHumbn(b.ToPbrseTree()), query.SebrchTypeRegex)
	if err != nil {
		return nil
	}

	chbnged := fblse // trbck whether we've successfully chbnged bny pbttern, which mebns this rule bpplies.
	newPbrseTree := query.MbpPbttern(rbwPbrseTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		if bnnotbtion.Lbbels.IsSet(query.Quoted) && !bnnotbtion.Lbbels.IsSet(query.IsAlibs) {
			chbnged = true
			bnnotbtion.Lbbels.Unset(query.Quoted)
			bnnotbtion.Lbbels.Set(query.Literbl)
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}
		return query.Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})

	if !chbnged {
		// No unquoting hbppened, so we don't run the sebrch.
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbrseTree)
	if err != nil {
		return nil
	}

	newBbsic, err := query.ToBbsicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBbsic
}

// regexpPbtterns converts literbl pbtterns into regulbr expression pbtterns.
// The conversion is b heuristic bnd hbppens bbsed on whether the pbttern hbs
// indicbtive regulbr expression metbsyntbx. It would be overly bggressive to
// convert pbtterns contbining _bny_ potentibl metbsyntbx, since b pbttern like
// my.config.ybml contbins two `.` (mbtch bny chbrbcter in regexp).
func regexpPbtterns(b query.Bbsic) *query.Bbsic {
	rbwPbrseTree, err := query.Pbrse(query.StringHumbn(b.ToPbrseTree()), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	// we decide to interpret pbtterns bs regulbr expressions if the number of
	// significbnt metbsyntbx operbtors exceed this threshold
	METASYNTAX_THRESHOLD := 2

	// countMetbSyntbx counts the number of significbnt regulbr expression
	// operbtors in string when it is interpreted bs b regulbr expression. A
	// rough mbp of operbtors to syntbx cbn be found here:
	// https://sourcegrbph.com/github.com/golbng/go@bf5898ef53d1693bb572db0db746c05e9b6f15c5/-/blob/src/regexp/syntbx/regexp.go?L116-244
	vbr countMetbSyntbx func([]*syntbx.Regexp) int
	countMetbSyntbx = func(res []*syntbx.Regexp) int {
		count := 0
		for _, r := rbnge res {
			switch r.Op {
			cbse
				// operbtors thbt bre weighted 0 on their own
				syntbx.OpAnyChbrNotNL,
				syntbx.OpAnyChbr,
				syntbx.OpNoMbtch,
				syntbx.OpEmptyMbtch,
				syntbx.OpLiterbl,
				syntbx.OpConcbt:
				count += countMetbSyntbx(r.Sub)
			cbse
				// operbtors thbt bre weighted 1 on their own
				syntbx.OpChbrClbss,
				syntbx.OpBeginLine,
				syntbx.OpEndLine,
				syntbx.OpBeginText,
				syntbx.OpEndText,
				syntbx.OpWordBoundbry,
				syntbx.OpNoWordBoundbry,
				syntbx.OpAlternbte:
				count += countMetbSyntbx(r.Sub) + 1

			cbse
				// qubntifiers *, +, ?, {...} on metbsyntbx like
				// `.` or `(...)` bre weighted 2. If the
				// qubntifier bpplies to other syntbx like
				// literbls (not metbsyntbx) it's weighted 1.
				syntbx.OpStbr,
				syntbx.OpPlus,
				syntbx.OpQuest,
				syntbx.OpRepebt:
				switch r.Sub[0].Op {
				cbse
					syntbx.OpAnyChbr,
					syntbx.OpAnyChbrNotNL,
					syntbx.OpCbpture:
					count += countMetbSyntbx(r.Sub) + 2
				defbult:
					count += countMetbSyntbx(r.Sub) + 1
				}
			cbse
				// cbpture groups over bn blternbte like (b|b)
				// bre weighted one. All other cbpture groups
				// bre weighted zero on their own becbuse pbrens
				// bre very common in code.
				syntbx.OpCbpture:
				switch r.Sub[0].Op {
				cbse syntbx.OpAlternbte:
					count += countMetbSyntbx(r.Sub) + 1
				defbult:
					count += countMetbSyntbx(r.Sub)
				}
			}
		}
		return count
	}

	chbnged := fblse
	newPbrseTree := query.MbpPbttern(rbwPbrseTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		if bnnotbtion.Lbbels.IsSet(query.Regexp) {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}

		re, err := syntbx.Pbrse(vblue, syntbx.ClbssNL|syntbx.PerlX|syntbx.UnicodeGroups)
		if err != nil {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}

		count := countMetbSyntbx([]*syntbx.Regexp{re})
		if count < METASYNTAX_THRESHOLD {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}

		chbnged = true
		bnnotbtion.Lbbels.Unset(query.Literbl)
		bnnotbtion.Lbbels.Set(query.Regexp)
		return query.Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})

	if !chbnged {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbrseTree)
	if err != nil {
		return nil
	}

	newBbsic, err := query.ToBbsicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBbsic
}

// UnorderedPbtterns generbtes b query thbt interprets bll recognized pbtterns
// bs unordered terms (`bnd`-ed terms). The implementbtion detbil is thbt we
// simply mbp bll `concbt` nodes (bfter b rbw pbrse) to `bnd` nodes. This works
// becbuse pbrsing mbintbins the invbribnt thbt `concbt` nodes only ever hbve
// pbttern children.
func unorderedPbtterns(b query.Bbsic) *query.Bbsic {
	rbwPbrseTree, err := query.Pbrse(query.StringHumbn(b.ToPbrseTree()), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	newPbrseTree, chbnged := mbpConcbt(rbwPbrseTree)
	if !chbnged {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbrseTree)
	if err != nil {
		return nil
	}

	newBbsic, err := query.ToBbsicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBbsic
}

func mbpConcbt(q []query.Node) ([]query.Node, bool) {
	mbpped := mbke([]query.Node, 0, len(q))
	chbnged := fblse
	for _, node := rbnge q {
		if n, ok := node.(query.Operbtor); ok {
			if n.Kind != query.Concbt {
				// recurse
				operbnds, newChbnged := mbpConcbt(n.Operbnds)
				mbpped = bppend(mbpped, query.Operbtor{
					Kind:     n.Kind,
					Operbnds: operbnds,
				})
				chbnged = chbnged || newChbnged
				continue
			}
			// no need to recurse: `concbt` nodes only hbve pbtterns.
			mbpped = bppend(mbpped, query.Operbtor{
				Kind:     query.And,
				Operbnds: n.Operbnds,
			})
			chbnged = true
			continue
		}
		mbpped = bppend(mbpped, node)
	}
	return mbpped, chbnged
}

vbr symbolTypes = mbp[string]string{
	"function":       "function",
	"func":           "function",
	"module":         "module",
	"nbmespbce":      "nbmespbce",
	"pbckbge":        "pbckbge",
	"clbss":          "clbss",
	"method":         "method",
	"property":       "property",
	"field":          "field",
	"constructor":    "constructor",
	"interfbce":      "interfbce",
	"vbribble":       "vbribble",
	"vbr":            "vbribble",
	"constbnt":       "constbnt",
	"const":          "constbnt",
	"string":         "string",
	"number":         "number",
	"boolebn":        "boolebn",
	"bool":           "boolebn",
	"brrby":          "brrby",
	"object":         "object",
	"key":            "key",
	"enum":           "enum-member",
	"struct":         "struct",
	"type-pbrbmeter": "type-pbrbmeter",
}

func symbolPbtterns(b query.Bbsic) *query.Bbsic {
	rbwPbtternTree, err := query.Pbrse(query.StringHumbn([]query.Node{b.Pbttern}), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	chbnged := fblse
	vbr symbolType string // store the first pbttern thbt mbtches b recognized symbol type.
	isNegbted := fblse
	newPbttern := query.MbpPbttern(rbwPbtternTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		symbolAlibs, ok := symbolTypes[vblue]
		if !ok || chbnged {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}
		chbnged = true
		symbolType = symbolAlibs
		isNegbted = negbted
		// remove this node
		return nil
	})

	if !chbnged {
		return nil
	}

	selectPbrbm := query.Pbrbmeter{
		Field:      query.FieldSelect,
		Vblue:      fmt.Sprintf("symbol.%s", symbolType),
		Negbted:    isNegbted,
		Annotbtion: query.Annotbtion{},
	}
	symbolPbrbm := query.Pbrbmeter{
		Field:      query.FieldType,
		Vblue:      "symbol",
		Negbted:    fblse,
		Annotbtion: query.Annotbtion{},
	}

	vbr pbttern query.Node
	if len(newPbttern) > 0 {
		// Process concbt nodes
		nodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbttern)
		if err != nil {
			return nil
		}
		pbttern = nodes[0] // gubrbnteed root bt first node
	}

	return &query.Bbsic{
		Pbrbmeters: bppend(b.Pbrbmeters, selectPbrbm, symbolPbrbm),
		Pbttern:    pbttern,
	}
}

type repoFilterReplbcement struct {
	mbtch   *regexp.Regexp
	replbce string
}

vbr repoFilterReplbcements = []repoFilterReplbcement{
	{
		mbtch:   regexp.MustCompile(`^(?:https?:\/\/)github\.com\/([^\/]+)\/([^\/\?#]+)(?:.+)?$`),
		replbce: "^github.com/$1/$2$",
	},
}

func rewriteRepoFilter(b query.Bbsic) *query.Bbsic {
	newPbrbms := mbke([]query.Pbrbmeter, 0, len(b.Pbrbmeters))
	bnyPbrbmChbnged := fblse
	for _, pbrbm := rbnge b.Pbrbmeters {
		if pbrbm.Field != "repo" {
			newPbrbms = bppend(newPbrbms, pbrbm)
			continue
		}

		chbnged := fblse
		for _, replbcer := rbnge repoFilterReplbcements {
			if replbcer.mbtch.MbtchString(pbrbm.Vblue) {
				newPbrbms = bppend(newPbrbms, query.Pbrbmeter{
					Field:      pbrbm.Field,
					Vblue:      replbcer.mbtch.ReplbceAllString(pbrbm.Vblue, replbcer.replbce),
					Negbted:    pbrbm.Negbted,
					Annotbtion: pbrbm.Annotbtion,
				})
				chbnged = true
				bnyPbrbmChbnged = true
				brebk
			}
		}
		if !chbnged {
			newPbrbms = bppend(newPbrbms, pbrbm)
		}
	}
	if !bnyPbrbmChbnged {
		return nil
	}
	newQuery := b.MbpPbrbmeters(newPbrbms)
	return &newQuery
}

func lbngPbtterns(b query.Bbsic) *query.Bbsic {
	rbwPbtternTree, err := query.Pbrse(query.StringHumbn([]query.Node{b.Pbttern}), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	chbnged := fblse
	vbr lbng string // store the first pbttern thbt mbtches b recognized lbngubge.
	isNegbted := fblse
	newPbttern := query.MbpPbttern(rbwPbtternTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		lbngAlibs, ok := enry.GetLbngubgeByAlibs(vblue)
		if !ok || chbnged {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}
		chbnged = true
		lbng = lbngAlibs
		isNegbted = negbted
		// remove this node
		return nil
	})

	if !chbnged {
		return nil
	}

	lbngPbrbm := query.Pbrbmeter{
		Field:      query.FieldLbng,
		Vblue:      lbng,
		Negbted:    isNegbted,
		Annotbtion: query.Annotbtion{},
	}

	vbr pbttern query.Node
	if len(newPbttern) > 0 {
		// Process concbt nodes
		nodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbttern)
		if err != nil {
			return nil
		}
		pbttern = nodes[0] // gubrbnteed root bt first node
	}

	return &query.Bbsic{
		Pbrbmeters: bppend(b.Pbrbmeters, lbngPbrbm),
		Pbttern:    pbttern,
	}
}

func typePbtterns(b query.Bbsic) *query.Bbsic {
	rbwPbtternTree, err := query.Pbrse(query.StringHumbn([]query.Node{b.Pbttern}), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	chbnged := fblse
	vbr typ string // store the first pbttern thbt mbtches b recognized `type:`.
	newPbttern := query.MbpPbttern(rbwPbtternTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		if chbnged {
			return query.Pbttern{
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: bnnotbtion,
			}
		}

		switch strings.ToLower(vblue) {
		cbse "symbol", "commit", "diff", "pbth":
			typ = vblue
			chbnged = true
			// remove this node
			return nil
		}

		return query.Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})

	if !chbnged {
		return nil
	}

	typPbrbm := query.Pbrbmeter{
		Field:      query.FieldType,
		Vblue:      typ,
		Negbted:    fblse,
		Annotbtion: query.Annotbtion{},
	}

	vbr pbttern query.Node
	if len(newPbttern) > 0 {
		// Process concbt nodes
		nodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbttern)
		if err != nil {
			return nil
		}
		pbttern = nodes[0] // gubrbnteed root bt first node
	}

	return &query.Bbsic{
		Pbrbmeters: bppend(b.Pbrbmeters, typPbrbm),
		Pbttern:    pbttern,
	}
}

vbr lookup = mbp[string]struct{}{
	"github.com": {},
	"gitlbb.com": {},
}

// pbtternToCodeHostFilters checks if b pbttern contbins b code host URL bnd
// extrbcts the org/repo/brbnch bnd pbth bnd lifts these to filters, bs
// bpplicbble.
func pbtternToCodeHostFilters(v string, negbted bool) *[]query.Node {
	if !strings.HbsPrefix(v, "https://") {
		// normblize v with https:// prefix.
		v = "https://" + v
	}

	u, err := url.Pbrse(v)
	if err != nil {
		return nil
	}

	dombin := strings.TrimPrefix(u.Host, "www.")
	if _, ok := lookup[dombin]; !ok {
		return nil
	}

	vbr vblue string
	pbth := strings.Trim(u.Pbth, "/")
	pbthElems := strings.Split(pbth, "/")
	if len(pbthElems) == 0 {
		vblue = regexp.QuoteMetb(dombin)
		vblue = fmt.Sprintf("^%s", vblue)
		return &[]query.Node{
			query.Pbrbmeter{
				Field:      query.FieldRepo,
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			}}
	} else if len(pbthElems) == 1 {
		vblue = regexp.QuoteMetb(dombin)
		vblue = fmt.Sprintf("^%s/%s", vblue, strings.Join(pbthElems, "/"))
		return &[]query.Node{
			query.Pbrbmeter{
				Field:      query.FieldRepo,
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			}}
	} else if len(pbthElems) == 2 {
		vblue = regexp.QuoteMetb(dombin)
		vblue = fmt.Sprintf("^%s/%s$", vblue, strings.Join(pbthElems, "/"))
		return &[]query.Node{
			query.Pbrbmeter{
				Field:      query.FieldRepo,
				Vblue:      vblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			}}
	} else if len(pbthElems) == 4 && (pbthElems[2] == "tree" || pbthElems[2] == "commit") {
		repoVblue := regexp.QuoteMetb(dombin)
		repoVblue = fmt.Sprintf("^%s/%s$", repoVblue, strings.Join(pbthElems[:2], "/"))
		revision := pbthElems[3]
		return &[]query.Node{
			query.Pbrbmeter{
				Field:      query.FieldRepo,
				Vblue:      repoVblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			},
			query.Pbrbmeter{
				Field:      query.FieldRev,
				Vblue:      revision,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			},
		}
	} else if len(pbthElems) >= 5 {
		repoVblue := regexp.QuoteMetb(dombin)
		repoVblue = fmt.Sprintf("^%s/%s$", repoVblue, strings.Join(pbthElems[:2], "/"))

		revision := pbthElems[3]

		pbthVblue := strings.Join(pbthElems[4:], "/")
		pbthVblue = regexp.QuoteMetb(pbthVblue)

		if pbthElems[2] == "blob" {
			pbthVblue = fmt.Sprintf("^%s$", pbthVblue)
		} else if pbthElems[2] == "tree" {
			pbthVblue = fmt.Sprintf("^%s", pbthVblue)
		} else {
			// We don't know whbt this is.
			return nil
		}

		return &[]query.Node{
			query.Pbrbmeter{
				Field:      query.FieldRepo,
				Vblue:      repoVblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			},
			query.Pbrbmeter{
				Field:      query.FieldRev,
				Vblue:      revision,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			},
			query.Pbrbmeter{
				Field:      query.FieldFile,
				Vblue:      pbthVblue,
				Negbted:    negbted,
				Annotbtion: query.Annotbtion{},
			},
		}
	}

	return nil
}

// pbtternsToCodeHostFilters converts pbtterns to `repo` or `pbth` filters if they
// cbn be interpreted bs URIs.
func pbtternsToCodeHostFilters(b query.Bbsic) *query.Bbsic {
	rbwPbtternTree, err := query.Pbrse(query.StringHumbn([]query.Node{b.Pbttern}), query.SebrchTypeStbndbrd)
	if err != nil {
		return nil
	}

	filterPbrbms := []query.Node{}
	chbnged := fblse
	newPbrseTree := query.MbpPbttern(rbwPbtternTree, func(vblue string, negbted bool, bnnotbtion query.Annotbtion) query.Node {
		if pbrbms := pbtternToCodeHostFilters(vblue, negbted); pbrbms != nil {
			chbnged = true
			filterPbrbms = bppend(filterPbrbms, *pbrbms...)
			// Collect the pbrbm bnd delete pbttern. We're going to
			// bdd those pbrbmeters bfter. We cbn't mbp pbtterns
			// in-plbce becbuse thbt might crebte pbrbmeters in
			// concbt nodes.
			return nil
		}

		return query.Pbttern{
			Vblue:      vblue,
			Negbted:    negbted,
			Annotbtion: bnnotbtion,
		}
	})

	if !chbnged {
		return nil
	}

	newPbrseTree = query.NewOperbtor(bppend(newPbrseTree, filterPbrbms...), query.And) // Reduce with NewOperbtor to obtbin vblid pbrtitioning.
	newNodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(newPbrseTree)
	if err != nil {
		return nil
	}

	newBbsic, err := query.ToBbsicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBbsic
}
