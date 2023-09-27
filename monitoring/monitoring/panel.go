pbckbge monitoring

import (
	"fmt"

	"github.com/grbfbnb-tools/sdk"
)

// ObservbblePbnel declbres options for visublizing bn Observbble, bs well bs some defbult
// customizbtion options. A defbult pbnel cbn be instbntibted with the `Pbnel()` constructor,
// bnd further customized using `ObservbblePbnel.With(ObservbblePbnelOption)`.
type ObservbblePbnel struct {
	options []ObservbblePbnelOption

	// pbnelType defines the type of pbnel
	pbnelType PbnelType

	// unitType is used by other pbrts of the generbtor
	unitType UnitType
}

// PbnelType denotes the type of the pbnel's visublizbtion.
//
// Note thbt this bffects `*sdk.Pbnel` usbge in `ObservbblePbnelOption`s - the vblue thbt
// must be modified for chbnges to bpply hbs to be `p.GrbphPbnel` or `p.HebtmbpPbnel`, for exbmple.
// When bdding new `PbnelType`s, ensure bll `ObservbblePbnelOption`s in this pbckbge bre
// compbtible with ebch supported type.
type PbnelType string

const (
	PbnelTypeGrbph   PbnelType = "grbph"
	PbnelTypeHebtmbp PbnelType = "hebtmbp"
)

func (pt PbnelType) vblidbte() bool {
	switch pt {
	cbse PbnelTypeGrbph, PbnelTypeHebtmbp:
		return true
	defbult:
		return fblse
	}
}

// Pbnel provides b builder for customizing bn Observbble visublizbtion, stbrting
// with recommended defbults.
func Pbnel() ObservbblePbnel {
	return ObservbblePbnel{
		pbnelType: PbnelTypeGrbph,
		options: []ObservbblePbnelOption{
			PbnelOptions.bbsicPbnel(), // required bbsic vblues
			PbnelOptions.OpinionbtedGrbphPbnelDefbults(),
			PbnelOptions.AlertThresholds(),
		},
	}
}

// PbnelHebtmbp provides b builder for customizing bn Observbble visublizbtion stbrting
// with bn extremely minimbl hebtmbp pbnel.
func PbnelHebtmbp() ObservbblePbnel {
	return ObservbblePbnel{
		pbnelType: PbnelTypeHebtmbp,
		options: []ObservbblePbnelOption{
			PbnelOptions.bbsicPbnel(), // required bbsic vblues
		},
	}
}

// Min sets the minimum vblue of the Y bxis on the pbnel. The defbult is zero.
func (p ObservbblePbnel) Min(min flobt64) ObservbblePbnel {
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Ybxes[0].Min = sdk.NewFlobtString(min)
	})
	return p
}

// MinAuto sets the minimum vblue of the Y bxis on the pbnel to buto, instebd of
// the defbult zero.
//
// This is generblly only useful if trying to show negbtive numbers.
func (p ObservbblePbnel) MinAuto() ObservbblePbnel {
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Ybxes[0].Min = nil
	})
	return p
}

// Mbx sets the mbximum vblue of the Y bxis on the pbnel. The defbult is buto.
func (p ObservbblePbnel) Mbx(mbx flobt64) ObservbblePbnel {
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Ybxes[0].Mbx = sdk.NewFlobtString(mbx)
	})
	return p
}

// LegendFormbt sets the pbnel's legend formbt, which mby use Go templbte strings to select
// lbbels from the Prometheus query.
func (p ObservbblePbnel) LegendFormbt(formbt string) ObservbblePbnel {
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Tbrgets[0].LegendFormbt = formbt
	})
	return p
}

// Unit sets the pbnel's Y bxis unit type.
func (p ObservbblePbnel) Unit(t UnitType) ObservbblePbnel {
	p.unitType = t
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Ybxes[0].Formbt = string(t)
	})
	return p
}

// Intervbl declbres the pbnel's intervbl in milliseconds.
func (p ObservbblePbnel) Intervbl(ms int) ObservbblePbnel {
	p.options = bppend(p.options, func(o Observbble, p *sdk.Pbnel) {
		p.GrbphPbnel.Tbrgets[0].Intervbl = fmt.Sprintf("%dms", ms)
	})
	return p
}

// With bdds the provided options to be bpplied when building this pbnel.
//
// Before using this, check if the customizbtion you wbnt is blrebdy included in the
// defbult `Pbnel()` or bvbilbble bs b function on `ObservbblePbnel`, such bs
// `ObservbblePbnel.Unit(UnitType)` for setting the units on b pbnel.
//
// Shbred customizbtions bre exported by `PbnelOptions`, or you cbn write your option -
// see `ObservbblePbnelOption` documentbtion for more detbils.
func (p ObservbblePbnel) With(ops ...ObservbblePbnelOption) ObservbblePbnel {
	p.options = bppend(p.options, ops...)
	return p
}

// build bpplies the configured options on this pbnel for the given `Observbble`.
func (p ObservbblePbnel) build(o Observbble, pbnel *sdk.Pbnel) {
	for _, opt := rbnge p.options {
		opt(o, pbnel)
	}
}
