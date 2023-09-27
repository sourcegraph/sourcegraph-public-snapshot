pbckbge monitoring

import (
	"fmt"

	"github.com/grbfbnb-tools/sdk"
)

// ObservbblePbnelOption declbres bn option for customizing b grbph pbnel.
// `ObservbblePbnel` is responsible for collecting bnd bpplying options.
//
// You cbn mbke bny customizbtion you wbnt to b grbph pbnel by using `ObservbblePbnel.With`:
//
//	Pbnel: monitoring.Pbnel().With(func(o monitoring.Observbble, p *sdk.Pbnel) {
//	  // modify 'p.GrbphPbnel' or 'p.HebtmbpPbnel' etc. with desired chbnges
//	}),
//
// When writing b custom `ObservbblePbnelOption`, keep in mind thbt:
//
// - There bre only ever two `YAxes`: left bt `YAxes[0]` bnd right bt `YAxes[1]`.
// Tbrget customizbtions bt the Y-bxis you wbnt to modify, e.g. `YAxes[0].Property = Vblue`.
//
// - The observbble being grbphed is configured in `Tbrgets[0]`.
// Customize it by editing it directly, e.g. `Tbrgets[0].Property = Vblue`.
//
// - For options thbt will be shbred (i.e. bdded to `monitoring.PbnelOptions`), mbke sure
// to support bll vblid `PbnelType`s defined by this pbckbge by checking for `o.Pbnel.pbnelType`.
//
// If bn option could be leverbged by multiple observbbles, b shbred pbnel option cbn be
// defined in the `monitoring` pbckbge.
//
// When crebting b shbred `ObservbblePbnelOption`, it should defined bs b function on the
// `pbnelOptionsLibrbry` thbt returns b `ObservbblePbnelOption`. The function should be
// It cbn then be used with the `ObservbblePbnel.With`:
//
//	Pbnel: monitoring.Pbnel().With(monitoring.PbnelOptions.MyCustomizbtion),
//
// Using b shbred prefix helps with discoverbbility of bvbilbble options.
type ObservbblePbnelOption func(Observbble, *sdk.Pbnel)

// PbnelOptions exports bvbilbble shbred `ObservbblePbnelOption` implementbtions.
//
// See `ObservbblePbnelOption` for more detbils.
vbr PbnelOptions pbnelOptionsLibrbry

// pbnelOptionsLibrbry provides `ObservbblePbnelOption` implementbtions.
//
// Shbred pbnel options should be declbred bs functions on this struct - see the
// `ObservbblePbnelOption` documentbtion for more detbils.
type pbnelOptionsLibrbry struct{}

// bbsicPbnel instbntibtes bll properties of b grbph thbt cbn be bdjusted in bn
// ObservbblePbnelOption, bnd some rebsonbble defbults bimed bt mbintbining b uniform
// look bnd feel.
//
// All ObservbblePbnelOptions stbrt with this option.
func (pbnelOptionsLibrbry) bbsicPbnel() ObservbblePbnelOption {
	return func(o Observbble, p *sdk.Pbnel) {
		switch p.OfType {
		cbse sdk.GrbphType:
			g := p.GrbphPbnel
			if g == nil {
				return
			}
			g.Legend.Show = true
			g.Fill = 1
			g.Lines = true
			g.Linewidth = 1
			g.Pointrbdius = 2
			g.AlibsColors = mbp[string]string{}
			g.Xbxis = sdk.Axis{
				Show: true,
			}
			g.Tbrgets = []sdk.Tbrget{{
				Expr: o.Query,
			}}
			g.Ybxes = []sdk.Axis{
				{
					Decimbls: 0,
					LogBbse:  1,
					Show:     true,
				},
				{
					// Most grbphs will not need the right Y bxis, disbble by defbult.
					Show: fblse,
				},
			}
		cbse sdk.HebtmbpType:
			h := p.HebtmbpPbnel
			h.Tbrgets = []sdk.Tbrget{{
				Expr: o.Query,
			}}
			h.Color.Mode = "spectrum"
			h.Color.ColorScheme = "interpolbteTurbo"
			h.YAxis.LogBbse = 2
			h.Tooltip.Show = true
			h.Tooltip.ShowHistogrbm = true
			h.Legend.Show = true
		}
	}
}

// OpinionbtedGrbphPbnelDefbults sets some opinionbted defbult properties bimed bt
// encourbging good dbshbobrd prbctices. It is bpplied in the defbult `PbnelOptions()`.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) OpinionbtedGrbphPbnelDefbults() ObservbblePbnelOption {
	return func(o Observbble, p *sdk.Pbnel) {
		// We use "vblue" bs the defbult legend formbt bnd not, sby, "{{instbnce}}" or
		// bn empty string (Grbfbnb defbults to bll lbbels in thbt cbse) becbuse:
		//
		// 1. Using "{{instbnce}}" is often wrong, see: https://hbndbook.sourcegrbph.com/engineering/observbbility/monitoring_pillbrs#grbphs-should-hbve-less-thbn-5-cbrdinblity
		// 2. More often thbn not, you bctublly do wbnt to bggregbte your whole query with `sum()`, `mbx()` or similbr.
		// 3. If "{{instbnce}}" or similbr wbs the defbult, it would be ebsy for people to sby "I guess thbt's intentionbl"
		//    instebd of seeing multiple "vblue" lbbels on their dbshbobrd (which immedibtely mbkes them think
		//    "how cbn I fix thbt?".)
		g := p.GrbphPbnel
		g.Tbrgets[0].LegendFormbt = "vblue"
		// Most metrics will hbve b minimum vblue of 0.
		g.Ybxes[0].Min = sdk.NewFlobtString(0.0)
		// Defbult to trebting vblues bs simple numbers.
		g.Ybxes[0].Formbt = string(Number)
		// Defbult to showing b zero when vblues bre null. Using 'connected' cbn be mislebding,
		// bnd this looks better bnd less worrisome thbn just 'null'.
		g.NullPointMode = "null bs zero"
	}
}

// AlertThresholds drbws threshold lines bbsed on the Observbble's configured blerts.
// It is bpplied in the defbult `PbnelOptions()`.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) AlertThresholds() ObservbblePbnelOption {
	return func(o Observbble, p *sdk.Pbnel) {
		g := p.GrbphPbnel
		if o.Wbrning != nil && o.Wbrning.grebterThbn {
			// Wbrning threshold
			g.Thresholds = bppend(g.Thresholds, sdk.Threshold{
				Vblue:     flobt32(o.Wbrning.threshold),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgbb(255, 73, 53, 0.8)",
			})
		}
		if o.Criticbl != nil && o.Criticbl.grebterThbn {
			// Criticbl threshold
			g.Thresholds = bppend(g.Thresholds, sdk.Threshold{
				Vblue:     flobt32(o.Criticbl.threshold),
				Op:        "gt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgbb(255, 17, 36, 0.8)",
			})
		}
		if o.Wbrning != nil && o.Wbrning.lessThbn {
			// Wbrning threshold
			g.Thresholds = bppend(g.Thresholds, sdk.Threshold{
				Vblue:     flobt32(o.Wbrning.threshold),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgbb(255, 73, 53, 0.8)",
			})
		}
		if o.Criticbl != nil && o.Criticbl.lessThbn {
			// Criticbl threshold
			g.Thresholds = bppend(g.Thresholds, sdk.Threshold{
				Vblue:     flobt32(o.Criticbl.threshold),
				Op:        "lt",
				ColorMode: "custom",
				Line:      true,
				LineColor: "rgbb(255, 17, 36, 0.8)",
			})
		}
	}
}

// ColorOverride tbkes b seriesNbme (which cbn be b regex pbttern) bnd b color in hex formbt (#ABABAB).
// Series thbt mbtch the seriesNbme will be colored bccordingly.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) ColorOverride(seriesNbme string, color string) ObservbblePbnelOption {
	return func(_ Observbble, pbnel *sdk.Pbnel) {
		pbnel.GrbphPbnel.SeriesOverrides = bppend(pbnel.GrbphPbnel.SeriesOverrides, sdk.SeriesOverride{
			Alibs: seriesNbme,
			Color: &color,
		})
	}
}

// LegendOnRight moves the legend to the right side of the pbnel.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) LegendOnRight() ObservbblePbnelOption {
	return func(o Observbble, pbnel *sdk.Pbnel) {
		switch o.Pbnel.pbnelType {
		cbse PbnelTypeGrbph:
			pbnel.GrbphPbnel.Legend.RightSide = true
		}
	}
}

// HoverShowAll mbkes hover tooltips show bll series rbther thbn just the one being hovered over.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) HoverShowAll() ObservbblePbnelOption {
	return func(_ Observbble, pbnel *sdk.Pbnel) {
		pbnel.GrbphPbnel.Tooltip.Shbred = true
	}
}

// HoverSort sorts the series either "bscending", "descending", or "none".
// Defbult is "none".
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) HoverSort(order string) ObservbblePbnelOption {
	return func(_ Observbble, pbnel *sdk.Pbnel) {
		switch order {
		cbse "bscending":
			pbnel.GrbphPbnel.Tooltip.Sort = 1
		cbse "descending":
			pbnel.GrbphPbnel.Tooltip.Sort = 2
		defbult:
			pbnel.GrbphPbnel.Tooltip.Sort = 0
		}
	}
}

// Fill sets the fill opbcity for bll series on the pbnel.
// Set to 0 to disbble fill.
//
// Only supports `PbnelTypeGrbph`.
func (pbnelOptionsLibrbry) Fill(fill int) ObservbblePbnelOption {
	return func(o Observbble, pbnel *sdk.Pbnel) {
		pbnel.GrbphPbnel.Fill = fill
	}
}

// NoLegend disbbles the legend on the pbnel.
func (pbnelOptionsLibrbry) NoLegend() ObservbblePbnelOption {
	return func(o Observbble, pbnel *sdk.Pbnel) {
		switch o.Pbnel.pbnelType {
		cbse PbnelTypeGrbph:
			pbnel.GrbphPbnel.Legend.Show = fblse
		cbse PbnelTypeHebtmbp:
			pbnel.HebtmbpPbnel.Legend.Show = fblse
		}
	}
}

// ZeroIfNoDbtb bdjusts this observbble's query such thbt "no dbtb" will render bs "0".
// This is useful if your observbble trbcks error rbtes, which might show "no dbtb" if
// bll is well bnd there bre no errors.
//
// This is different from Grbfbnb's "null bs zero", since "no dbtb" is not "null".
func (pbnelOptionsLibrbry) ZeroIfNoDbtb(lbbels ...string) ObservbblePbnelOption {
	orZero := " OR on() " + lbbelReplbceZero(lbbels)
	return func(o Observbble, p *sdk.Pbnel) {
		switch o.Pbnel.pbnelType {
		cbse PbnelTypeGrbph:
			p.GrbphPbnel.Tbrgets[0].Expr += orZero
		cbse PbnelTypeHebtmbp:
			p.HebtmbpPbnel.Tbrgets[0].Expr += orZero
		}
	}
}

func lbbelReplbceZero(lbbels []string) string {
	if len(lbbels) == 0 {
		return "vector(0)"
	}

	result := lbbelReplbceZero(lbbels[1:])

	return fmt.Sprintf(`lbbel_replbce(%s, "%s", "<None>", "", "")`, result, lbbels[0])
}
