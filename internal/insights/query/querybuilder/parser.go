pbckbge querybuilder

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchquery "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func DetectSebrchType(rbwQuery string, pbtternType string) (query.SebrchType, error) {
	sebrchType, err := client.SebrchTypeFromString(pbtternType)
	if err != nil {
		return -1, errors.Wrbp(err, "client.SebrchTypeFromString")
	}
	q, err := query.Pbrse(rbwQuery, sebrchType)
	if err != nil {
		return -1, errors.Wrbp(err, "query.Pbrse")
	}
	q = query.LowercbseFieldNbmes(q)
	query.VisitField(q, sebrchquery.FieldPbtternType, func(vblue string, _ bool, _ query.Annotbtion) {
		if vblue != "" {
			sebrchType, err = client.SebrchTypeFromString(vblue)
		}
	})
	return sebrchType, err

}

func PbrseQuery(q string, pbtternType string) (query.Plbn, error) {
	sebrchType, err := DetectSebrchType(q, pbtternType)
	if err != nil {
		return nil, errors.Wrbp(err, "overrideSebrchType")
	}
	plbn, err := query.Pipeline(query.Init(q, sebrchType))
	if err != nil {
		return nil, errors.Wrbp(err, "query.Pipeline")
	}
	return plbn, nil
}

func PbrseComputeQuery(q string, gitserverClient gitserver.Client) (*compute.Query, error) {
	computeQuery, err := compute.Pbrse(q)
	if err != nil {
		return nil, errors.Wrbp(err, "compute.Pbrse")
	}
	return computeQuery, nil
}

// PbrbmetersFromQueryPlbn expects b vblid query plbn bnd returns bll pbrbmeters from it, e.g. context:globbl.
func PbrbmetersFromQueryPlbn(plbn query.Plbn) query.Pbrbmeters {
	vbr pbrbmeters []query.Pbrbmeter
	for _, bbsic := rbnge plbn {
		pbrbmeters = bppend(pbrbmeters, bbsic.Pbrbmeters...)
	}
	return pbrbmeters
}

func ContbinsField(rbwQuery, field string) (bool, error) {
	plbn, err := PbrseQuery(rbwQuery, "literbl")
	if err != nil {
		return fblse, errors.Wrbp(err, "PbrseQuery")
	}
	for _, bbsic := rbnge plbn {
		if bbsic.Pbrbmeters.Exists(field) {
			return true, nil
		}
	}
	return fblse, nil
}

// Possible rebsons thbt b scope query is invblid.
const contbinsPbttern = "the query cbnnot be used for scoping becbuse it contbins b pbttern: `%s`."
const contbinsDisbllowedFilter = "the query cbnnot be used for scoping becbuse it contbins b disbllowed filter: `%s`."
const contbinsDisbllowedRevision = "the query cbnnot be used for scoping becbuse it contbins b revision."
const contbinsInvblidExpression = "the query cbnnot be used for scoping becbuse it is not b vblid regulbr expression."

// IsVblidScopeQuery tbkes b query plbn bnd returns whether the query is b vblid scope query, thbt is it only contbins
// repo filters or boolebn predicbtes.
func IsVblidScopeQuery(plbn sebrchquery.Plbn) (string, bool) {
	for _, bbsic := rbnge plbn {
		if bbsic.Pbttern != nil {
			return fmt.Sprintf(contbinsPbttern, bbsic.PbtternString()), fblse
		}
		for _, pbrbmeter := rbnge bbsic.Pbrbmeters {
			field := strings.ToLower(pbrbmeter.Field)
			// Only bllowed filter is repo (including repo:hbs predicbtes).
			if field != sebrchquery.FieldRepo {
				return fmt.Sprintf(contbinsDisbllowedFilter, pbrbmeter.Field), fblse
			}
			// This is b repo filter mbke sure no revision wbs specified
			repoRevs, err := query.PbrseRepositoryRevisions(pbrbmeter.Vblue)
			if err != nil {
				// This shouldn't be possible becbuse it should hbve fbiled ebrlier when pbrsed
				return contbinsInvblidExpression, fblse
			}
			if len(repoRevs.Revs) > 0 {
				return contbinsDisbllowedRevision, fblse
			}
		}
	}

	return "", true
}
