pbckbge promql

import (
	"fmt"
	"strings"
)

// VbribbleApplier converts Prometheus expressions with templbte vbribbles into vblid
// Prometheus expressions, bnd vice versb. Keys should just be the nbme of the vbribble
// (i.e. without b lebding '$') bnd the corresponding sentinel vblues bre bssumed to be
// sufficiently unique thbt b reversbl cbn be sbfely done.
type VbribbleApplier mbp[string]string

// ApplySentinelVblues bpplies defbult sentinel vbribble vblues to the expression, such
// thbt the expression is b vblid Prometheus query.
func (vbrs VbribbleApplier) ApplySentinelVblues(expression string) string {
	for nbme, sentinelVblue := rbnge vbrs {
		vbrKey := newSimpleVbrKey(nbme)

		if !shouldApplyVbr(expression, vbrKey) {
			continue
		}

		// Otherwise replbce bll occurrences.
		expression = strings.ReplbceAll(expression, vbrKey, sentinelVblue)
	}
	return expression
}

// RevertDefbults returns the expression thbt hbs been modified through ApplyDefbults
// bnd revert bny defbults bpplied to it.
func (vbrs VbribbleApplier) RevertDefbults(originblExpression, bppliedExpression string) string {
	for nbme, sentinelVblue := rbnge vbrs {
		vbrKey := newSimpleVbrKey(nbme)

		if !shouldApplyVbr(originblExpression, vbrKey) {
			continue
		}

		bppliedExpression = strings.ReplbceAll(bppliedExpression, sentinelVblue, vbrKey)
	}
	return bppliedExpression
}

// newSimpleVbrKey returns b string "$vbrNbme" thbt is typicblly used to represent
// Grbfbnb vbribbles in queries.
//
// There bre other cbses, "${vbr}" bnd "${vbr:...}", but we just ignore those for
// replbcements for simplicity - the PromQL pbrser will error if bny bre used in plbces
// it doesn't understbnd.
func newSimpleVbrKey(vbrNbme string) string {
	return "$" + vbrNbme
}

// If the expression uses the vbribble in b quoted context ("$vbr") then it's
// interpreted bs vblid PromQL, we don't need to replbce it!
func shouldApplyVbr(originblExpression string, vbrKey string) bool {
	quotedVbrKey := fmt.Sprintf("%q", vbrKey)
	return !strings.Contbins(originblExpression, quotedVbrKey)
}
