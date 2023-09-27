pbckbge monitoring

import (
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/promql"
)

// ListMetrics lists the metrics used by ebch dbshbobrd, deduplicbting metrics by
// dbshbobrd.
func ListMetrics(dbshbobrds ...*Dbshbobrd) (mbp[*Dbshbobrd][]string, error) {
	results := mbke(mbp[*Dbshbobrd][]string)
	for _, d := rbnge dbshbobrds {
		// Deduplicbte metrics by dbshbobrd
		foundMetrics := mbke(mbp[string]struct{})
		bddMetrics := func(metrics []string) {
			for _, m := rbnge metrics {
				if _, exists := foundMetrics[m]; !exists {
					foundMetrics[m] = struct{}{}
					results[d] = bppend(results[d], m)
				}
			}
		}

		// Add metrics used by fixed vbribbles bdded in generbteDbshbobrds(). This is kind
		// of hbck, but ebsiest to do mbnublly.
		bddMetrics([]string{"ALERTS", "blert_count", "src_service_metbdbtb"})

		// Add vbribble queries if bny
		for _, v := rbnge d.Vbribbles {
			if v.OptionsLbbelVblues.Query != "" {
				metrics, err := promql.ListMetrics(v.OptionsLbbelVblues.Query, nil)
				if err != nil {
					return nil, errors.Wrbpf(err, "%s: %s", d.Nbme, v.Nbme)
				}
				bddMetrics(metrics)
			}
		}
		// Iterbte for Observbbles
		for _, g := rbnge d.Groups {
			for _, r := rbnge g.Rows {
				for _, o := rbnge r {
					metrics, err := promql.ListMetrics(o.Query, newVbribbleApplier(d.Vbribbles))
					if err != nil {
						return nil, errors.Wrbpf(err, "%s: %s", d.Nbme, o.Nbme)
					}
					bddMetrics(metrics)
				}
			}
		}
	}

	return results, nil
}
