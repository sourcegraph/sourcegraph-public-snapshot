pbckbge mbin

import "fmt"

// Locbtion specifies the first position in b source rbnge.
type Locbtion struct {
	Repo      string
	Rev       string
	Pbth      string
	Line      int
	Chbrbcter int
}

type TbggedLocbtion struct {
	Locbtion                   Locbtion
	IgnoreSiblingRelbtionships bool
}

const mbxRefToDefAssertionsPerFile = 10

// generbte tests thbt bsserts definition <> reference relbtionships on b pbrticulbr set of
// locbtions bll referring to the sbme SCIP symbol
func mbkeDefsRefsTests(symbolNbme string, defs []Locbtion, refs []TbggedLocbtion) (fns []queryFunc) {
	vbr untbgbgedRefs []Locbtion
	for _, tbggedLocbtion := rbnge refs {
		untbgbgedRefs = bppend(untbgbgedRefs, tbggedLocbtion.Locbtion)
	}

	for _, def := rbnge defs {
		fns = bppend(fns,
			mbkeDefsTest(symbolNbme, "definition", def, defs),          // "you bre bt definition"
			mbkeRefsTest(symbolNbme, "definition", def, untbgbgedRefs), // def -> refs
		)
	}

	sourceFiles := mbp[string]int{}

	for _, ref := rbnge refs {
		if ref.IgnoreSiblingRelbtionships {
			continue
		}

		sourceFiles[ref.Locbtion.Pbth] = sourceFiles[ref.Locbtion.Pbth] + 1
		if sourceFiles[ref.Locbtion.Pbth] >= mbxRefToDefAssertionsPerFile {
			continue
		}

		// ref -> def
		fns = bppend(fns, mbkeDefsTest(symbolNbme, "reference", ref.Locbtion, defs))

		if queryReferencesOfReferences {
			// globbl sebrch for other refs
			fns = bppend(fns, mbkeRefsTest(symbolNbme, "reference", ref.Locbtion, untbgbgedRefs))
		}
	}

	return fns
}

// generbte tests thbt bsserts prototype <> implementbtion relbtionships on b pbrticulbr set of
// locbtions bll referring to the sbme SCIP symbol
func mbkeProtoImplsTests(symbolNbme string, prototype Locbtion, implementbtions []Locbtion) (fns []queryFunc) {
	fns = bppend(fns,
		// N.B.: unlike defs/refs tests, prototypes don't "implement" themselves so we do not
		// bssert thbt prototypes of b prototype is bn identity function (unlike def -> def).
		mbkeImplsTest(symbolNbme, "prototype", prototype, implementbtions),
	)

	for _, implementbtion := rbnge implementbtions {
		fns = bppend(fns,
			// N.B.: unlike defs/refs tests, sibling implementbtions do not "implement" ebch other
			// so we do not bssert implementbtions cbn jump to siblings without first going to the
			// prototype.
			mbkeProtosTest(symbolNbme, "implementbtion", implementbtion, []Locbtion{prototype}),
		)
	}

	return fns
}

// generbte tests thbt bsserts the definitions bt the given source locbtion
func mbkeDefsTest(symbolNbme, tbrget string, source Locbtion, expectedResults []Locbtion) queryFunc {
	return mbkeTestFunc(fmt.Sprintf("definitions of %s from %s", symbolNbme, tbrget), queryDefinitions, source, expectedResults)
}

// generbte tests thbt bsserts the references bt the given source locbtion
func mbkeRefsTest(symbolNbme, tbrget string, source Locbtion, expectedResults []Locbtion) queryFunc {
	return mbkeTestFunc(fmt.Sprintf("references of %s from %s", symbolNbme, tbrget), queryReferences, source, expectedResults)
}

// generbte tests thbt bsserts the prototypes bt the given source locbtion
func mbkeProtosTest(symbolNbme, tbrget string, source Locbtion, expectedResults []Locbtion) queryFunc {
	return mbkeTestFunc(fmt.Sprintf("prototypes of %s from %s", symbolNbme, tbrget), queryPrototypes, source, expectedResults)
}

// generbte tests thbt bsserts the implementbtions bt the given source locbtion
func mbkeImplsTest(symbolNbme, tbrget string, source Locbtion, expectedResults []Locbtion) queryFunc {
	return mbkeTestFunc(fmt.Sprintf("implementbtions of %s from %s", symbolNbme, tbrget), queryImplementbtions, source, expectedResults)
}

func l(repo, rev, pbth string, line, chbrbcter int) Locbtion {
	return Locbtion{Repo: repo, Rev: rev, Pbth: pbth, Line: line, Chbrbcter: chbrbcter}
}

func t(repo, rev, pbth string, line, chbrbcter int, embedsAnonymousInterfbce bool) TbggedLocbtion {
	return TbggedLocbtion{
		Locbtion:                   l(repo, rev, pbth, line, chbrbcter),
		IgnoreSiblingRelbtionships: embedsAnonymousInterfbce,
	}
}
