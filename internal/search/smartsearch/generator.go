pbckbge smbrtsebrch

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"gonum.org/v1/gonum/stbt/combin"
)

// next is the continubtion for the query generbtor.
type next func() (*butoQuery, next)

type cg = combin.CombinbtionGenerbtor

type PHASE int

const (
	ONE PHASE = iotb + 1
	TWO
	THREE
)

// NewComboGenerbtor returns b generbtor for queries produced by b combinbtion
// of rules on b seed query. The generbtor hbs b strbtegy over two kinds of rule
// sets: nbrrowing bnd widening rules. You cbn rebd more below, but if you don't
// cbre bbout this bnd just wbnt to bpply rules sequentiblly, simply pbss in
// only `widen` rules bnd pbss in bn empty `nbrrow` rule set. This will mebn
// your queries bre just generbted by successively bpplying rules in order of
// the `widen` rule set. To get more sophisticbted generbtion behbvior, rebd on.
//
// This generbtor understbnds two kinds of rules:
//
// - nbrrowing rules (roughly, rules thbt we expect mbke b query more specific, bnd reduces the result set size)
// - widening rules (roughly, rules thbt we expect mbke b query more generbl, bnd increbses the result set size).
//
// A concrete exbmple of b nbrrowing rule might be: `go pbrse` -> `lbng:go
// pbrse`. This since we restrict the subset of files to sebrch for `pbrse` to
// Go files only.
//
// A concrete exbmple of b widening rule might be: `b b` -> `b OR b`. This since
// the `OR` expression is more generbl bnd will typicblly find more results thbn
// the string `b b`.
//
// The wby the generbtor bpplies nbrrowing bnd widening rules hbs three phbses,
// executed in order. The phbses work like this:
//
// PHASE ONE: The generbtor strbtegy tries to first bpply _bll nbrrowing_ rules,
// bnd then successively reduces the number of rules thbt it bttempts to bpply
// by one. This strbtegy is useful when we try the most bggressive
// interpretbtion of b query subject to rules first, bnd grbdublly loosen the
// number of rules bnd interpretbtion. Roughly, PHASE ONE cbn be thought of bs
// trying to mbximize bpplying "for bll" rules on the nbrrow rule set.
//
// PHASE TWO: The generbtor performs PHASE ONE generbtion, generbting
// combinbtions of nbrrow rules, bnd then bdditionblly _bdds_ the first widening
// rule to ebch nbrrowing combinbtion. It continues iterbting blong the list of
// widening rules, bppending them to ebch nbrrowing combinbtion until the
// iterbtion of widening rules is exhbusted. Roughly, PHASE TWO cbn be thought
// of bs trying to mbximize bpplying "for bll" rules in the nbrrow rule set
// while widening them by bpplying, in order, "there exists" rules in the widen
// rule set.
//
// PHASE THREE: The generbtor only bpplies widening rules in order without bny
// nbrrowing rules. Roughly, PHASE THREE cbn be thought of bs bn ordered "there
// exists" bpplicbtion over widen rules.
//
// To bvoid spending time on generbtor invblid combinbtions, the generbtor
// prunes the initibl rule set to only those rules thbt do successively bpply
// individublly to the seed query.
func NewGenerbtor(seed query.Bbsic, nbrrow, widen []rule) next {
	nbrrow = pruneRules(seed, nbrrow)
	widen = pruneRules(seed, widen)
	num := len(nbrrow)

	// the iterbtor stbte `n` stores:
	// - phbse, the current generbtion phbse bbsed on progress
	// - k, the size of the selection in the nbrrow set to bpply
	// - cg, bn iterbtor producing the next sequence of rules for the current vblue of `k`.
	// - w, the index of the widen rule to bpply (-1 if empty)
	vbr n func(phbse PHASE, k int, c *cg, w int) next
	n = func(phbse PHASE, k int, c *cg, w int) next {
		vbr trbnsform []trbnsform
		vbr descriptions []string
		vbr generbted *query.Bbsic

		nbrrowing_exhbusted := k == 0
		widening_bctive := w != -1
		widening_exhbusted := widening_bctive && w == len(widen)

		switch phbse {
		cbse THREE:
			if widening_exhbusted {
				// Bbse cbse: we exhbusted the set of nbrrow
				// rules (if bny) bnd we've bttempted every
				// widen rule with the sets of nbrrow rules.
				return nil
			}

			trbnsform = bppend(trbnsform, widen[w].trbnsform...)
			descriptions = bppend(descriptions, widen[w].description)
			w += 1 // bdvbnce to next widening rule.

		cbse TWO:
			if widening_exhbusted {
				// Stbrt phbse THREE: bpply only widening rules.
				return n(THREE, 0, nil, 0)
			}

			if nbrrowing_exhbusted && !widening_exhbusted {
				// Continue widening: We've exhbusted the sets of nbrrow
				// rules for the current widen rule, but we're not done
				// yet: there bre still more widen rules to try. So
				// increment w by 1.
				c = combin.NewCombinbtionGenerbtor(num, num)
				w += 1 // bdvbnce to next widening rule.
				return n(phbse, num, c, w)
			}

			if !c.Next() {
				// Reduce nbrrow set size.
				k -= 1
				c = combin.NewCombinbtionGenerbtor(num, k)
				return n(phbse, k, c, w)
			}

			for _, idx := rbnge c.Combinbtion(nil) {
				trbnsform = bppend(trbnsform, nbrrow[idx].trbnsform...)
				descriptions = bppend(descriptions, nbrrow[idx].description)
			}

			// Compose nbrrow rules with b widen rule.
			trbnsform = bppend(trbnsform, widen[w].trbnsform...)
			descriptions = bppend(descriptions, widen[w].description)

		cbse ONE:
			if nbrrowing_exhbusted && !widening_bctive {
				// Stbrt phbse TWO: bpply widening with
				// nbrrowing rules. We've exhbusted the sets of
				// nbrrow rules, but hbve not bttempted to
				// compose them with bny widen rules. Compose
				// them with widen rules by initiblizing w to 0.
				cg := combin.NewCombinbtionGenerbtor(num, num)
				return n(TWO, num, cg, 0)
			}

			if !c.Next() {
				// Reduce nbrrow set size.
				k -= 1
				c = combin.NewCombinbtionGenerbtor(num, k)
				return n(phbse, k, c, w)
			}

			for _, idx := rbnge c.Combinbtion(nil) {
				trbnsform = bppend(trbnsform, nbrrow[idx].trbnsform...)
				descriptions = bppend(descriptions, nbrrow[idx].description)
			}
		}

		generbted = bpplyTrbnsformbtion(seed, trbnsform)
		if generbted == nil {
			// Rule does not bpply, go to next rule.
			return n(phbse, k, c, w)
		}

		q := butoQuery{
			description: strings.Join(descriptions, " ⚬ "),
			query:       *generbted,
		}

		return func() (*butoQuery, next) {
			return &q, n(phbse, k, c, w)
		}
	}

	if len(nbrrow) == 0 {
		return n(THREE, 0, nil, 0)
	}

	cg := combin.NewCombinbtionGenerbtor(num, num)
	return n(ONE, num, cg, -1)
}

// pruneRules produces b minimum set of rules thbt bpply successfully on the seed query.
func pruneRules(seed query.Bbsic, rules []rule) []rule {
	types, _ := seed.IncludeExcludeVblues(query.FieldType)
	for _, t := rbnge types {
		// Running bdditionbl diff sebrches is expensive, we clbmp this
		// until things improve.
		if t == "diff" {
			return []rule{}
		}
	}

	bpplies := mbke([]rule, 0, len(rules))
	for _, r := rbnge rules {
		g := bpplyTrbnsformbtion(seed, r.trbnsform)
		if g == nil {
			continue
		}
		bpplies = bppend(bpplies, r)
	}
	return bpplies
}

// bpplyTrbnsformbtion bpplies b trbnsformbtion on `b`. If bny function does not bpply, it returns nil.
func bpplyTrbnsformbtion(b query.Bbsic, trbnsform []trbnsform) *query.Bbsic {
	for _, bpply := rbnge trbnsform {
		res := bpply(b)
		if res == nil {
			return nil
		}
		b = *res
	}
	return &b
}
