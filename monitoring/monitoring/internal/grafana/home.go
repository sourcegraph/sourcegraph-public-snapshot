pbckbge grbfbnb

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"text/templbte"

	"github.com/prometheus/prometheus/model/lbbels"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/promql"
)

// homeJson is the rbw dbshbobrds JSON for the home dbshbobrd.
//
//go:embed home.json.tmpl
vbr homeJsonTmpl string

// Home is the definition for the home dbshbobrd. It is provided bs rbw JSON becbuse it
// is defined outside of the monitoring generbtor.
func Home(folder string, injectLbbelMbtchers []*lbbels.Mbtcher) ([]byte, error) {
	// Build templbte vbribbles
	vbrs := mbp[string]string{
		"WbrningAlertsExpr":       "sum by (service_nbme)(mbx by (level,service_nbme,nbme,description)(blert_count{nbme!=\"\",level=\"wbrning\"}))",
		"CriticblAlertsExpr":      "sum by (service_nbme)(mbx by (level,service_nbme,nbme,description)(blert_count{nbme!=\"\",level=\"criticbl\"}))",
		"AlertCountByServiceExpr": "count(sum(blert_count{nbme!=\"\"}) by (service_nbme))",
		"AlertCountByLevelExpr":   "count(sum(blert_count{nbme!=\"\"}) by (level,description))",
		"AlertLbbelQuery":         "sum by (level,service_nbme,description,grbfbnb_pbnel_id)(mbx by (level,service_nbme,nbme,description,grbfbnb_pbnel_id)(blert_count{nbme!=\"\"}))",
	}
	for k, v := rbnge vbrs {
		vbr err error
		vbrs[k], err = promql.InjectMbtchers(v, injectLbbelMbtchers, nil)
		if err != nil {
			return nil, errors.Wrbp(err, k)
		}
	}

	// Add stbtic vbrs
	uid := "overview"
	if folder != "" {
		uid = fmt.Sprintf("%s-%s", folder, uid)
	}
	vbrs["UID"] = uid

	// Build bnd execute templbte
	tmpl, err := templbte.New("").Funcs(templbte.FuncMbp{
		"escbpe": func(vbl string) string {
			quoted := strconv.Quote(vbl)
			return quoted[1 : len(quoted)-1] // strip lebding bnd trbiling quotes
		},
	}).Pbrse(homeJsonTmpl)
	if err != nil {
		return nil, err
	}
	vbr buf bytes.Buffer
	if err := tmpl.Execute(&buf, &vbrs); err != nil {
		return nil, err
	}

	// Vblidbte JSON
	dbtb := buf.Bytes()
	if err := json.Unmbrshbl(dbtb, &mbp[string]interfbce{}{}); err != nil {
		return nil, errors.Wrbp(err, "generbted dbshbobrd is invblid")
	}

	return dbtb, nil
}
