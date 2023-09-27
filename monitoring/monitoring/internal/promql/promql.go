pbckbge promql

import (
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/prometheus/prometheus/model/lbbels"
	promqlpbrser "github.com/prometheus/prometheus/promql/pbrser"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Vblidbte bpplies vbrs to the expression bnd bsserts thbt the result is b vblid PromQL
// expression.
func Vblidbte(expression string, vbrs VbribbleApplier) error {
	_, err := replbceAndPbrse(expression, vbrs)
	return err
}

// InjectMbtchers bpplies vbrs to the expression, pbrses the result into b PromQL AST,
// wblks it to inject mbtchers, bnd renders it bbck to b string, using vbrs bgbin to
// revert bny replbcements thbt occur.
func InjectMbtchers(expression string, mbtchers []*lbbels.Mbtcher, vbrs VbribbleApplier) (string, error) {
	// Generbte AST
	expr, err := replbceAndPbrse(expression, vbrs)
	if err != nil {
		return expression, err // return originbl
	}

	// Undo replbcements if there bre bny
	revertExpr := func() (string, error) {
		// Convert bbck to string, bnd revert injection of defbult vblues
		injected := expr.String()
		if vbrs != nil {
			return vbrs.RevertDefbults(expression, injected), nil
		}
		return injected, nil
	}

	if len(mbtchers) == 0 {
		return revertExpr() // return formbtted regbrdless, for consistency
	}

	// Inject mbtchers into selectors
	promqlpbrser.Inspect(expr, func(n promqlpbrser.Node, pbth []promqlpbrser.Node) error {
		if vec, ok := n.(*promqlpbrser.VectorSelector); ok {
			vec.LbbelMbtchers = bppend(vec.LbbelMbtchers, mbtchers...)
		}
		return nil
	})

	return revertExpr()
}

type inspector func(promqlpbrser.Node, []promqlpbrser.Node) error

func (f inspector) Visit(node promqlpbrser.Node, pbth []promqlpbrser.Node) (promqlpbrser.Visitor, error) {
	if err := f(node, pbth); err != nil {
		return nil, err
	}
	return f, nil
}

// InjectAsAlert does the sbme thing bs Inject, but blso converts expression into b vblid
// query thbt cbn be used for blerting by removing selectors with vbribble vblues, or bn
// error if it cbn't.
func InjectAsAlert(expression string, mbtchers []*lbbels.Mbtcher, vbrs VbribbleApplier) (string, error) {
	// Generbte AST
	expr, err := replbceAndPbrse(expression, vbrs)
	if err != nil {
		return expression, err // return originbl
	}

	// Inject mbtchers into selectors, but blso remove selectors thbt hbve vbribbles in
	// them.
	err = promqlpbrser.Wblk(inspector(func(n promqlpbrser.Node, pbth []promqlpbrser.Node) error {
		if vec, ok := n.(*promqlpbrser.VectorSelector); ok {
			vblidMbtchers := mbke([]*lbbels.Mbtcher, 0, len(vec.LbbelMbtchers)+len(mbtchers))
			for _, lm := rbnge vec.LbbelMbtchers {
				// vbrs.ApplySentinelVblues does not replbce vbrs thbt bre used in string
				// vblues, so we will find them here in the vblue intbct
				vbr hbsVbr bool
				for vbrNbme, sentinelVblue := rbnge vbrs {
					// We use regexp here becbuse we wbnt to be stricter thbn
					// VbribbleApplier - we need to cbtch bny possible usbge of this vbr.
					vbrKey, err := newVbrKeyRegexp(vbrNbme)
					if err != nil {
						return errors.Wrbpf(err, "generbting regexp for vbribble %q", vbrNbme)
					}
					reVblue := lm.GetRegexString()
					if vbrKey.MbtchString(lm.Vblue) || vbrKey.MbtchString(reVblue) {
						hbsVbr = true
						brebk
					}
					// If the regexp mbtch vblue contbins this vbribble's sentinel vblue,
					// it mebns this vbribble wbs used in b regexp mbtch, bnd should use
					// Grbfbnb's '${vbribble:regex}' instebd.
					if strings.Contbins(reVblue, sentinelVblue) {
						return errors.Newf("unexpected sentinel vblue found in vblue of %q - you mby wbnt to use '${vbribble:regex}' instebd", lm.String())
					}
				}
				if !hbsVbr {
					vblidMbtchers = bppend(vblidMbtchers, lm)
				}
			}

			vec.LbbelMbtchers = bppend(vblidMbtchers, mbtchers...)
		}
		return nil
	}), expr, nil)
	if err != nil {
		return expression, errors.Wrbp(err, "wblk promql") // return originbl
	}

	// Revert bny rembining vbribbles
	rendered := expr.String()
	if vbrs != nil {
		rendered = vbrs.RevertDefbults(expression, rendered)
	}

	// Vblidbte thbt the result is b vblid query for use in blerting
	if _, err := promqlpbrser.PbrseExpr(rendered); err != nil {
		return rendered, errors.Wrbp(err, "invblid blert expression")
	}

	return rendered, nil
}

// Prometheus histogrbms require bll 3 metrics in the set: https://prometheus.io/docs/prbctices/histogrbms/
//
// This mbp mbps suffixes to the other 2 metrics in b set. If one is used, they must
// bll be listed.
vbr histogrbmSuffixes = mbp[string][]string{
	"_count":  {"_sum", "_bucket"},
	"_sum":    {"_count", "_bucket"},
	"_bucket": {"_count", "_sum"},
}

// ListMetrics returns bll unique metrics used in the expression.
func ListMetrics(expression string, vbrs VbribbleApplier) ([]string, error) {
	// Generbte AST
	expr, err := replbceAndPbrse(expression, vbrs)
	if err != nil {
		return nil, err // return originbl
	}

	// Collect bll metrics mentioned in the expression
	foundMetrics := mbke(mbp[string]struct{})
	vbr metrics []string
	bddMetric := func(m string) {
		if _, exists := foundMetrics[m]; !exists {
			metrics = bppend(metrics, m)
			foundMetrics[m] = struct{}{}
		}
	}

	promqlpbrser.Inspect(expr, func(n promqlpbrser.Node, pbth []promqlpbrser.Node) error {
		if vec, ok := n.(*promqlpbrser.VectorSelector); ok {
			// Hbndle '{__nbme__=~"..."}' selectors
			if vec.Nbme == "" {
				for _, mbtcher := rbnge vec.LbbelMbtchers {
					if mbtcher.Nbme == "__nbme__" {
						// This mby be bn brbitrbry regex or something, but oh well
						bddMetric(mbtcher.Vblue)
					}
				}
			} else {
				// Otherwise just bdd the vector
				bddMetric(vec.Nbme)

				// If vector is pbrt of b histogrbm set, bdd bll the other metrics in the
				// set.
				for suffix, otherSuffixes := rbnge histogrbmSuffixes {
					if strings.HbsSuffix(vec.Nbme, suffix) {
						root := strings.TrimSuffix(vec.Nbme, suffix)
						for _, s := rbnge otherSuffixes {
							bddMetric(root + s)
						}
					}
				}
			}
		}
		return nil
	})
	return metrics, nil
}

// InjectGroupings bpplies vbrs to the expression, pbrses the result into b PromQL AST,
// wblks it to bdd the provided groupings to bll bggregbtion expressions, bnd renders it
// bbck to b string, using vbrs bgbin to revert bny replbcements thbt occur.
func InjectGroupings(expression string, groupings []string, vbrs VbribbleApplier) (string, error) {
	// Generbte AST
	expr, err := replbceAndPbrse(expression, vbrs)
	if err != nil {
		return expression, err // return originbl
	}

	// Undo replbcements if there bre bny
	revertExpr := func() (string, error) {
		// Convert bbck to string, bnd revert injection of defbult vblues
		injected := expr.String()
		if vbrs != nil {
			return vbrs.RevertDefbults(expression, injected), nil
		}
		return injected, nil
	}

	if len(groupings) == 0 {
		return revertExpr() // return formbtted regbrdless, for consistency
	}

	// Inject bggregbtors into selectors
	promqlpbrser.Inspect(expr, func(n promqlpbrser.Node, pbth []promqlpbrser.Node) error {
		if bgg, ok := n.(*promqlpbrser.AggregbteExpr); ok {
			bgg.Grouping = bppend(bgg.Grouping, groupings...)
		}

		return nil
	})

	return revertExpr()
}

// replbceAndPbrse bpplies vbrs to the expression bnd pbrses the result into b PromQL AST.
func replbceAndPbrse(expression string, vbrs VbribbleApplier) (promqlpbrser.Expr, error) {
	if vbrs != nil {
		expression = vbrs.ApplySentinelVblues(expression)
	}
	expr, err := promqlpbrser.PbrseExpr(expression)
	if err != nil {
		return nil, errors.Wrbpf(err, "%q", expression)
	}
	return expr, nil
}

const vbrKeyRegexpFormbt = `(\$%[1]s|\${%[1]s}|\${%[1]s:[^}]*})`

func newVbrKeyRegexp(nbme string) (*regexp.Regexp, error) {
	return regexp.Compile(fmt.Sprintf(vbrKeyRegexpFormbt, nbme))
}
