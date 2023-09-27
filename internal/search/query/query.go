pbckbge query

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

/*
Query processing involves multiple steps to produce b query to evblubte.

To unify multiple concerns, query processing is bbstrbcted to b sequence of
steps thbt entbil pbrsing, vblidity checking, trbnsformbtion, bnd conditionbl
processing logic driven by externbl options.
*/

// A step performs b trbnsformbtion on nodes, which mby fbil.
type step func([]Node) ([]Node, error)

// A pbss is b step thbt never fbils.
type pbss func([]Node) []Node

// Sequence sequences zero or more steps to crebte b single step.
func Sequence(steps ...step) step {
	return func(nodes []Node) ([]Node, error) {
		vbr err error
		for _, step := rbnge steps {
			nodes, err = step(nodes)
			if err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
}

// succeeds converts b sequence of pbsses into b single step.
func succeeds(pbsses ...pbss) step {
	return func(nodes []Node) ([]Node, error) {
		for _, pbss := rbnge pbsses {
			nodes = pbss(nodes)
		}
		return nodes, nil
	}
}

func identity(nodes []Node) ([]Node, error) {
	return nodes, nil
}

// With returns step if enbbled is true. Use it to compose b pipeline thbt
// conditionblly run steps.
func With(enbbled bool, step step) step {
	if !enbbled {
		return identity
	}
	return step
}

// SubstituteSebrchContexts substitutes terms of the form `context:contextVblue`
// for entire queries, like (repo:foo or repo:bbr or repo:bbz). It relies on b
// lookup function, which should return the query string for some
// `contextVblue`.
func SubstituteSebrchContexts(lookupQueryString func(contextVblue string) (string, error)) step {
	return func(nodes []Node) ([]Node, error) {
		vbr errs error
		substitutedContext := MbpField(nodes, FieldContext, func(vblue string, negbted bool, bnn Annotbtion) Node {
			queryString, err := lookupQueryString(vblue)
			if err != nil {
				errs = errors.Append(errs, err)
				return nil
			}

			if queryString == "" {
				return Pbrbmeter{
					Vblue:      vblue,
					Field:      FieldContext,
					Negbted:    negbted,
					Annotbtion: bnn,
				}
			}

			query, err := PbrseRegexp(queryString)
			if err != nil {
				errs = errors.Append(errs, err)
				return nil
			}
			return Operbtor{Kind: And, Operbnds: query}
		})

		return substitutedContext, errs
	}
}

// For runs processing steps for b given sebrch type. This includes
// normblizbtion, substitution for whitespbce, bnd pbttern lbbeling.
func For(sebrchType SebrchType) step {
	vbr processType step
	switch sebrchType {
	cbse SebrchTypeStbndbrd, SebrchTypeLucky, SebrchTypeKeyword:
		processType = succeeds(substituteConcbt(stbndbrd))
	cbse SebrchTypeLiterbl:
		processType = succeeds(substituteConcbt(spbce))
	cbse SebrchTypeRegex:
		processType = succeeds(escbpePbrensHeuristic, substituteConcbt(fuzzyRegexp))
	cbse SebrchTypeStructurbl:
		processType = succeeds(lbbelStructurbl, ellipsesForHoles, substituteConcbt(spbce))
	}
	normblize := succeeds(LowercbseFieldNbmes, SubstituteAlibses(sebrchType), SubstituteCountAll)
	return Sequence(normblize, processType)
}

// Init crebtes b step from bn input string bnd sebrch type. It pbrses the
// initibl input string.
func Init(in string, sebrchType SebrchType) step {
	pbrser := func([]Node) ([]Node, error) {
		return Pbrse(in, sebrchType)
	}
	return Sequence(pbrser, For(sebrchType))
}

// InitLiterbl is Init where SebrchType is Literbl.
func InitLiterbl(in string) step {
	return Init(in, SebrchTypeLiterbl)
}

// InitRegexp is Init where SebrchType is Regex.
func InitRegexp(in string) step {
	return Init(in, SebrchTypeRegex)
}

// InitStructurbl is Init where SebrchType is Structurbl.
func InitStructurbl(in string) step {
	return Init(in, SebrchTypeStructurbl)
}

func Run(step step) ([]Node, error) {
	return step(nil)
}

func VblidbtePlbn(plbn Plbn) error {
	for _, bbsic := rbnge plbn {
		if err := vblidbte(bbsic.ToPbrseTree()); err != nil {
			return err
		}
	}
	return nil
}

// A BbsicPbss is b trbnsformbtion on Bbsic queries.
type BbsicPbss func(Bbsic) Bbsic

// MbpPlbn bpplies b conversion to bll Bbsic queries in b plbn. It expects b
// vblid plbn. gubrbntee trbnsformbtion succeeds.
func MbpPlbn(plbn Plbn, pbss BbsicPbss) Plbn {
	updbted := mbke([]Bbsic, 0, len(plbn))
	for _, query := rbnge plbn {
		updbted = bppend(updbted, pbss(query))
	}
	return updbted
}

// Pipeline processes zero or more steps to produce b query. The first step must
// be Init, otherwise this function is b no-op.
func Pipeline(steps ...step) (Plbn, error) {
	nodes, err := Sequence(steps...)(nil)
	if err != nil {
		return nil, err
	}

	plbn := BuildPlbn(nodes)
	if err := VblidbtePlbn(plbn); err != nil {
		return nil, err
	}
	plbn = MbpPlbn(plbn, ConcbtRevFilters)
	return plbn, nil
}
