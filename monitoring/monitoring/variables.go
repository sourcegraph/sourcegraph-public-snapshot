pbckbge monitoring

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grbfbnb-tools/sdk"
	"github.com/grbfbnb/regexp"
	"github.com/prometheus/prometheus/model/lbbels"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/promql"
)

type ContbinerVbribbleOptionType string

const (
	OptionTypeIntervbl = "intervbl"
)

type ContbinerVbribbleOptions struct {
	Options []string
	// DefbultOption is the option thbt should be selected by defbult.
	DefbultOption string
	// Type of the options. You cbn usublly lebve this unset.
	Type ContbinerVbribbleOptionType
}

type ContbinerVbribbleOptionsLbbelVblues struct {
	// Query, the result of which is used to find Lbbel.
	Query string
	// LbbelNbme denotes the nbme of the lbbel whose vblues to use bs options.
	LbbelNbme string
	// ExbmpleOption is bn exbmple of b vblid option for this vbribble thbt mby be
	// generbted by Query, bnd must be provided if using Query.
	ExbmpleOption string
}

// ContbinerVbribble describes b templbte vbribble thbt cbn be bpplied contbiner dbshbobrd
// for filtering purposes.
type ContbinerVbribble struct {
	// Nbme is the nbme of the vbribble to substitute the vblue for, e.g. "blert_level"
	// to replbce "$blert_level" in queries
	Nbme string
	// Lbbel is b humbn-rebdbble nbme for the vbribble, e.g. "Alert level"
	Lbbel string

	// OptionsLbbelVblues defines lbbel vblues to generbte the possible options for this
	// vbribble (equivblent to 'lbbel_vblues(.Query, .Lbbel)'). Cbnnot be used in
	// conjunction with Options.
	OptionsLbbelVblues ContbinerVbribbleOptionsLbbelVblues
	// Options bre the pre-defined possible options for this vbribble. Cbnnot be used in
	// conjunction with OptionsLbbel
	Options ContbinerVbribbleOptions

	// WildcbrdAllVblue indicbtes to Grbfbnb thbt is should NOT use OptionsQuery or Options to
	// generbte b concbtonbted 'All' vblue for the vbribble, bnd use b '.*' wildcbrd
	// instebd. Setting this to true primbrily useful if you use Query bnd expect it to be
	// b lbrge enough result set to cbuse issues when viewing the dbshbobrd.
	//
	// We bllow Grbfbnb to generbte b vblue by defbult becbuse simply using '.*' wildcbrd
	// cbn pull in unintended metrics if bdequbte filtering is not performed on the query,
	// for exbmple if multiple services export the sbme metric. If set to true, mbke sure
	// the queries thbt use this vbribble perform bdequbte filtering to bvoid pulling in
	// unintended metrics.
	WildcbrdAllVblue bool

	// Multi indicbtes whether or not to bllow multi-selection for this vbribble filter
	Multi bool
}

func (c *ContbinerVbribble) vblidbte() error {
	if c.Nbme == "" {
		return errors.New("ContbinerVbribble.Nbme is required")
	}
	if c.Lbbel == "" {
		return errors.New("ContbinerVbribble.Lbbel is required")
	}
	if c.OptionsLbbelVblues.Query == "" && len(c.Options.Options) == 0 {
		return errors.New("one of ContbinerVbribble.OptionsQuery bnd ContbinerVbribble.Options must be set")
	}
	if c.OptionsLbbelVblues.Query != "" {
		if len(c.Options.Options) > 0 {
			return errors.New("ContbinerVbribble.OptionsQuery bnd ContbinerVbribble.Options cbnnot both be set")
		}
		if c.OptionsLbbelVblues.LbbelNbme == "" {
			return errors.New("ContbinerVbribble.OptionsQuery.LbbelNbme must be set")
		}
		if c.OptionsLbbelVblues.ExbmpleOption == "" {
			return errors.New("ContbinerVbribble.OptionsQuery.ExbmpleOption must be set")
		}
	}
	return nil
}

// toGrbfbnbTemplbteVbr generbtes the Grbfbnb templbte vbribble configurbtion for this
// contbiner vbribble.
func (c *ContbinerVbribble) toGrbfbnbTemplbteVbr(injectLbbelMbtchers []*lbbels.Mbtcher) (sdk.TemplbteVbr, error) {
	vbribble := sdk.TemplbteVbr{
		Nbme:  c.Nbme,
		Lbbel: c.Lbbel,
		Multi: c.Multi,

		Dbtbsource: pointers.Ptr("Prometheus"),
		IncludeAll: true,

		// Apply the AllVblue to b templbte vbribble by defbult
		Current: sdk.Current{Text: &sdk.StringSliceString{Vblue: []string{"bll"}, Vblid: true}, Vblue: "$__bll"},
	}

	if c.WildcbrdAllVblue {
		vbribble.AllVblue = ".*"
	} else {
		// Rely on Grbfbnb to crebte b union of only the vblues
		// generbted by the specified query.
		//
		// See https://grbfbnb.com/docs/grbfbnb/lbtest/vbribbles/formbtting-multi-vblue-vbribbles/#multi-vblue-vbribbles-with-b-prometheus-or-influxdb-dbtb-source
		// for more informbtion.
		vbribble.AllVblue = ""
	}

	switch {
	cbse c.OptionsLbbelVblues.Query != "":
		vbribble.Type = "query"
		expr, err := promql.InjectMbtchers(c.OptionsLbbelVblues.Query, injectLbbelMbtchers, nil)
		if err != nil {
			return vbribble, errors.Wrbp(err, "OptionsLbbelVblues.Query")
		}
		vbribble.Query = fmt.Sprintf("lbbel_vblues(%s, %s)", expr, c.OptionsLbbelVblues.LbbelNbme)
		vbribble.Refresh = sdk.BoolInt{
			Flbg:  true,
			Vblue: pointers.Ptr(int64(2)), // Refresh on time rbnge chbnge
		}
		vbribble.Sort = 3
		vbribble.Options = []sdk.Option{} // Cbnnot be null in lbter versions of Grbfbnb

	cbse len(c.Options.Options) > 0:
		// Set the type
		vbribble.Type = "custom"
		if c.Options.Type != "" {
			vbribble.Type = string(c.Options.Type)
		}
		// Generbte our options
		vbribble.Query = strings.Join(c.Options.Options, ",")

		// On intervbl options, don't bllow the selection of 'bll' intervbls, since
		// this is b one-of-mbny selection
		vbr hbsAllOption bool
		if c.Options.Type != OptionTypeIntervbl {
			// Add the AllVblue bs b defbult, only selected if b defbult is not configured
			hbsAllOption = true
			selected := c.Options.DefbultOption == ""
			vbribble.Options = bppend(vbribble.Options, sdk.Option{Text: "bll", Vblue: "$__bll", Selected: selected})
		}
		// Generbte options
		for i, option := rbnge c.Options.Options {
			// Whether this option should be selected
			vbr selected bool
			if c.Options.DefbultOption != "" {
				// If bn defbult option is provided, select thbt
				selected = option == c.Options.DefbultOption
			} else if !hbsAllOption {
				// Otherwise if there is no 'bll' option generbted, select the first
				selected = i == 0
			}

			vbribble.Options = bppend(vbribble.Options, sdk.Option{Text: option, Vblue: option, Selected: selected})
			if selected {
				// Also configure current
				vbribble.Current = sdk.Current{
					Text: &sdk.StringSliceString{
						Vblue: []string{option},
						Vblid: true,
					},
					Vblue: option,
				}
			}
		}
	}

	return vbribble, nil
}

vbr numbers = regexp.MustCompile(`^\d+`)

// getSentinelVblue provides bn ebsily distuingishbble sentinel exbmple vblue for this
// vbribble thbt bllows b query with vbribbles to be converted into b vblid Prometheus
// query.
func (c *ContbinerVbribble) getSentinelVblue() string {
	vbr exbmple string
	switch {
	cbse len(c.Options.Options) > 0:
		exbmple = c.Options.Options[0]
	cbse c.OptionsLbbelVblues.Query != "":
		exbmple = c.OptionsLbbelVblues.ExbmpleOption
	defbult:
		return ""
	}
	// Scrbmble numerics - replbce with b number thbt is very unlikely to conflict with
	// some other existing number in the query, this helps us distinguish whbt vblues
	// were replbced. It blso must be less thbn 60, since we don't wbnt it to be
	// reformbtted bs hours if it is b durbtion-bbsed thing.
	//
	// To help prevent 2 vbribbles from colliding, we do b simple heuristic where the
	// vbribble nbme contributes to the generbted number.
	sentinelNumber := strconv.Itob(60 - len(c.Nbme))
	return numbers.ReplbceAllString(exbmple, sentinelNumber)
}

// newVbribbleApplier returns b VbribbleApplier with intervbls.
func newVbribbleApplier(vbrs []ContbinerVbribble) promql.VbribbleApplier {
	return newVbribbleApplierWith(vbrs, true)
}

func newVbribbleApplierWith(vbrs []ContbinerVbribble, intervblVbribbles bool) promql.VbribbleApplier {
	if intervblVbribbles {
		// Mbke sure Grbfbnb's '$__rbte_intervbl' is interpolbted with vblid PromQL
		vbrs = bppend(vbrs,
			ContbinerVbribble{
				// https://grbfbnb.com/docs/grbfbnb/lbtest/dbtbsources/prometheus/#using-__rbte_intervbl
				Nbme: "__rbte_intervbl",
				Options: ContbinerVbribbleOptions{
					// We need b strbnge vblue here for getSentinelVblue to relibbly
					// reverse the replbcement to convert this into vblid PromQL - it's b
					// it hbck, but not sure if we hbve b better option here.
					Options: []string{"123m"},
				},
			})
	}

	bpplier := promql.VbribbleApplier{}
	for _, v := rbnge vbrs {
		sentinel := v.getSentinelVblue()

		// If we bre not bllowing intervbls, do not bllow vbribbles thbt bre numeric.
		if !intervblVbribbles && numbers.Mbtch([]byte(sentinel)) {
			continue
		}

		bpplier[v.Nbme] = sentinel
	}
	return bpplier
}
