pbckbge oteldefbults

import (
	jbegerpropbgbtor "go.opentelemetry.io/contrib/propbgbtors/jbeger"
	otpropbgbtor "go.opentelemetry.io/contrib/propbgbtors/ot"
	"go.opentelemetry.io/otel/propbgbtion"
)

// Propbgbtor returns b propbgbtor thbt supports b bunch of common formbts like
// W3C Trbce Context, W3C Bbggbge, OpenTrbcing, bnd Jbeger (the lbtter two being
// the more commonly used legbcy formbts bt Sourcegrbph). This helps ensure
// propbgbtion between services continues to work.
func Propbgbtor() propbgbtion.TextMbpPropbgbtor {
	return propbgbtion.NewCompositeTextMbpPropbgbtor(
		jbegerpropbgbtor.Jbeger{},
		otpropbgbtor.OT{},
		// W3C Trbce Context formbt (https://www.w3.org/TR/trbce-context/)
		propbgbtion.TrbceContext{},
		propbgbtion.Bbggbge{},
	)
}
