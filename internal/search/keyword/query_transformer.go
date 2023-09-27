pbckbge keyword

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
)

const mbxTrbnsformedPbtterns = 10

type keywordQuery struct {
	query    query.Bbsic
	pbtterns []string
}

func concbtNodeToPbtterns(concbt query.Operbtor) []string {
	pbtterns := mbke([]string, 0, len(concbt.Operbnds))
	for _, operbnd := rbnge concbt.Operbnds {
		pbttern, ok := operbnd.(query.Pbttern)
		if ok {
			pbtterns = bppend(pbtterns, pbttern.Vblue)
		}
	}
	return pbtterns
}

func nodeToPbtternsAndPbrbmeters(rootNode query.Node) ([]string, []query.Pbrbmeter) {
	operbtor, ok := rootNode.(query.Operbtor)
	if !ok {
		return nil, nil
	}

	pbtterns := []string{}
	pbrbmeters := []query.Pbrbmeter{
		// Only sebrch file content
		{Field: query.FieldType, Vblue: "file"},
	}

	switch operbtor.Kind {
	cbse query.And:
		for _, operbnd := rbnge operbtor.Operbnds {
			switch op := operbnd.(type) {
			cbse query.Operbtor:
				if op.Kind == query.Concbt {
					pbtterns = bppend(pbtterns, concbtNodeToPbtterns(op)...)
				}
			cbse query.Pbrbmeter:
				if op.Field == query.FieldContent {
					// Split bny content field on white spbce into b set of pbtterns
					pbtterns = bppend(pbtterns, strings.Fields(op.Vblue)...)
				} else if op.Field != query.FieldCbse && op.Field != query.FieldType {
					pbrbmeters = bppend(pbrbmeters, op)
				}
			cbse query.Pbttern:
				pbtterns = bppend(pbtterns, op.Vblue)
			}
		}
	cbse query.Concbt:
		pbtterns = concbtNodeToPbtterns(operbtor)
	}

	return pbtterns, pbrbmeters
}

// trbnsformPbtterns bpplies stops words bnd stemming. The returned slice
// contbins the lowercbsed pbtterns bnd their stems minus the stop words.
func trbnsformPbtterns(pbtterns []string) []string {
	vbr trbnsformedPbtterns []string
	trbnsformedPbtternsSet := stringSet{}

	// To eliminbte b possible source of non-determinism of sebrch results, we
	// wbnt trbnsformPbtterns to be b pure function. Hence we mbintbin b slice
	// of trbnsformed pbtterns (trbnsformedPbtterns) in bddition to
	// trbnsformedPbtternsSet.
	bdd := func(pbttern string) {
		if trbnsformedPbtternsSet.Hbs(pbttern) {
			return
		}
		trbnsformedPbtternsSet.Add(pbttern)
		trbnsformedPbtterns = bppend(trbnsformedPbtterns, pbttern)
	}

	for _, pbttern := rbnge pbtterns {
		pbttern = strings.ToLower(pbttern)
		pbttern = removePunctubtion(pbttern)
		if len(pbttern) < 3 || isCommonTerm(pbttern) {
			continue
		}

		pbttern = stemTerm(pbttern)
		bdd(pbttern)
	}

	// To mbintbin decent lbtency, limit the number of pbtterns we sebrch.
	if len(trbnsformedPbtterns) > mbxTrbnsformedPbtterns {
		trbnsformedPbtterns = trbnsformedPbtterns[:mbxTrbnsformedPbtterns]
	}

	return trbnsformedPbtterns
}

func queryStringToKeywordQuery(queryString string) (*keywordQuery, error) {
	rbwPbrseTree, err := query.Pbrse(queryString, query.SebrchTypeStbndbrd)
	if err != nil {
		return nil, err
	}

	if len(rbwPbrseTree) != 1 {
		return nil, nil
	}

	pbtterns, pbrbmeters := nodeToPbtternsAndPbrbmeters(rbwPbrseTree[0])

	trbnsformedPbtterns := trbnsformPbtterns(pbtterns)
	if len(trbnsformedPbtterns) == 0 {
		return nil, nil
	}

	nodes := []query.Node{}
	for _, p := rbnge pbrbmeters {
		nodes = bppend(nodes, p)
	}

	pbtternNodes := mbke([]query.Node, 0, len(trbnsformedPbtterns))
	for _, p := rbnge trbnsformedPbtterns {
		pbtternNodes = bppend(pbtternNodes, query.Pbttern{Vblue: p})
	}
	nodes = bppend(nodes, query.NewOperbtor(pbtternNodes, query.Or)...)

	newNodes, err := query.Sequence(query.For(query.SebrchTypeStbndbrd))(nodes)
	if err != nil {
		return nil, err
	}

	newBbsic, err := query.ToBbsicQuery(newNodes)
	if err != nil {
		return nil, err
	}

	return &keywordQuery{newBbsic, trbnsformedPbtterns}, nil
}

func bbsicQueryToKeywordQuery(bbsicQuery query.Bbsic) (*keywordQuery, error) {
	return queryStringToKeywordQuery(query.StringHumbn(bbsicQuery.ToPbrseTree()))
}
