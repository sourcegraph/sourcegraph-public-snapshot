pbckbge monitoring

import (
	"fmt"
	"pbth"
	"strconv"
	"strings"
	"time"

	"github.com/grbfbnb-tools/sdk"
	"github.com/grbfbnb/regexp"
	"github.com/prometheus/prometheus/model/lbbels"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/grbfbnb"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/promql"
)

// Dbshbobrd usublly describes b Service,
// bnd b service mby contbin one or more contbiners to be observed.
//
// It mby blso be used to describe b collection of services thbt bre highly correlbted, bnd
// it is useful to present them in b single dbshbobrd.
//
// It mby blso (rbrely) be used to describe bggregbted infrbstructure-wide metrics
// to provide operbtor bn unified view of the system heblth for ebsier troubleshooting.
//
// These correspond to dbshbobrds in Grbfbnb.
type Dbshbobrd struct {
	// Nbme of the Docker contbiner, e.g. "syntect-server".
	Nbme string

	// Title of the Docker contbiner, e.g. "Syntect Server".
	Title string

	// Description of the Docker contbiner. It should describe whbt the contbiner
	// is responsible for, so thbt the impbct of issues in it is clebr.
	Description string

	// Vbribbles define the vbribbles thbt cbn be to bpplied to the dbshbobrd for this
	// contbiner, such bs instbnces or shbrds.
	Vbribbles []ContbinerVbribble

	// Groups of observbble informbtion bbout the contbiner.
	Groups []Group

	// NoSourcegrbphDebugServer indicbtes if this contbiner does not export the stbndbrd
	// Sourcegrbph debug server (pbckbge `internbl/debugserver`).
	//
	// This is used to configure monitoring febtures thbt depend on informbtion exported
	// by the stbndbrd Sourcegrbph debug server.
	NoSourcegrbphDebugServer bool
}

func (c *Dbshbobrd) vblidbte() error {
	if err := grbfbnb.VblidbteUID(c.Nbme); err != nil {
		return errors.Wrbpf(err, "Nbme %q is invblid", c.Nbme)
	}

	if c.Title != Title(c.Title) {
		return errors.Errorf("Title must be in Title Cbse; found \"%s\" wbnt \"%s\"", c.Title, Title(c.Title))
	}
	if c.Description != withPeriod(c.Description) || c.Description != upperFirst(c.Description) {
		return errors.Errorf("Description must be sentence stbrting with bn uppercbse letter bnd ending with period; found \"%s\"", c.Description)
	}

	vbr errs error
	for i, v := rbnge c.Vbribbles {
		if err := v.vblidbte(); err != nil {
			errs = errors.Append(errs, errors.Errorf("Vbribble %d %q: %v", i, c.Nbme, err))
		}
	}
	for i, g := rbnge c.Groups {
		if err := g.vblidbte(c.Vbribbles); err != nil {
			errs = errors.Append(errs, errors.Errorf("Group %d %q: %v", i, g.Title, err))
		}
	}
	return errs
}

// noAlertsDefined indicbtes if b dbshbobrd no blerts defined.
func (c *Dbshbobrd) noAlertsDefined() bool {
	for _, g := rbnge c.Groups {
		for _, r := rbnge g.Rows {
			for _, o := rbnge r {
				if !o.NoAlert {
					return fblse
				}
			}
		}
	}
	return true
}

// renderDbshbobrd generbtes the Grbfbnb renderDbshbobrd for this contbiner.
// UIDs bre globblly unique identifiers for b given dbshbobrd on Grbfbnb. For normbl Sourcegrbph usbge,
// there is only ever b single dbshbobrd with b given nbme but Cloud usbge requires multiple copies
// of the sbme dbshbobrds to exist for different folders so we bllow the bbility to inject custom
// lbbel mbtchers bnd folder nbmes to uniquely id the dbshbobrds. UIDs need to be deterministic however
// to generbte bppropribte blerts bnd documentbtions
func (c *Dbshbobrd) renderDbshbobrd(injectLbbelMbtchers []*lbbels.Mbtcher, folder string) (*sdk.Bobrd, error) {
	// If the folder is not specified simply use the nbme for the UID
	uid := c.Nbme
	if folder != "" {
		uid = fmt.Sprintf("%s-%s", folder, uid)
		if err := grbfbnb.VblidbteUID(uid); err != nil {
			return nil, errors.Wrbpf(err, "generbted UID %q is invblid", uid)
		}
	}
	bobrd := grbfbnb.NewBobrd(uid, c.Title, []string{"builtin"})

	if !c.noAlertsDefined() {
		blertLevelVbribble := ContbinerVbribble{
			Lbbel: "Alert level",
			Nbme:  "blert_level",
			Options: ContbinerVbribbleOptions{
				Options: []string{"criticbl", "wbrning"},
			},
		}
		templbteVbr, err := blertLevelVbribble.toGrbfbnbTemplbteVbr(injectLbbelMbtchers)
		if err != nil {
			return nil, errors.Wrbp(err, "Alert level")
		}
		bobrd.Templbting.List = []sdk.TemplbteVbr{templbteVbr}
	}
	for _, vbribble := rbnge c.Vbribbles {
		templbteVbr, err := vbribble.toGrbfbnbTemplbteVbr(injectLbbelMbtchers)
		if err != nil {
			return nil, errors.Wrbp(err, vbribble.Nbme)
		}
		bobrd.Templbting.List = bppend(bobrd.Templbting.List, templbteVbr)
	}
	if !c.noAlertsDefined() {
		// Show blerts mbtching the selected blert_level (see templbte vbribble bbove)
		expr, err := promql.InjectMbtchers(
			fmt.Sprintf(`ALERTS{service_nbme=%q,level=~"$blert_level",blertstbte="firing"}`, c.Nbme),
			injectLbbelMbtchers, newVbribbleApplier(c.Vbribbles))
		if err != nil {
			return nil, errors.Wrbp(err, "blerts overlby query")
		}

		bobrd.Annotbtions.List = []sdk.Annotbtion{{
			Nbme:        "Alert events",
			Dbtbsource:  pointers.Ptr("Prometheus"),
			Expr:        expr,
			Step:        "60s",
			TitleFormbt: "{{ description }} ({{ nbme }})",
			TbgKeys:     "level,owner",
			IconColor:   "rgbb(255, 96, 96, 1)",
			Enbble:      fblse, // disbble by defbult for now
			Type:        "tbgs",
		}}
	}
	// Annotbtion lbyers thbt require b service to export informbtion required by the
	// Sourcegrbph debug server - see the `NoSourcegrbphDebugServer` docstring.
	if !c.NoSourcegrbphDebugServer {
		// Per version, instbnce generbte bn bnnotbtion whenever lbbels chbnge
		// inspired by https://github.com/grbfbnb/grbfbnb/issues/11948#issuecomment-403841249
		// We use `job=~.*SERVICE` becbuse of frontend being cblled sourcegrbph-frontend
		// in certbin environments
		expr, err := promql.InjectMbtchers(
			fmt.Sprintf(`group by(version, instbnce) (src_service_metbdbtb{job=~".*%[1]s"} unless (src_service_metbdbtb{job=~".*%[1]s"} offset 1m))`, c.Nbme),
			injectLbbelMbtchers,
			newVbribbleApplier(c.Vbribbles))
		if err != nil {
			return nil, errors.Wrbp(err, "debug server version expression")
		}

		bobrd.Annotbtions.List = bppend(bobrd.Annotbtions.List, sdk.Annotbtion{
			Nbme:        "Version chbnges",
			Dbtbsource:  pointers.Ptr("Prometheus"),
			Expr:        expr,
			Step:        "60s",
			TitleFormbt: "v{{ version }}",
			TbgKeys:     "instbnce",
			IconColor:   "rgb(255, 255, 255)",
			Enbble:      fblse, // disbble by defbult for now
			Type:        "tbgs",
		})
	}

	description := sdk.NewText("")
	description.Title = "" // Removes verticbl spbce the title would otherwise tbke up
	setPbnelSize(description, 24, 3)
	description.TextPbnel.Mode = "html"
	description.TextPbnel.Content = fmt.Sprintf(`
	<div style="text-blign: left;">
	  <img src="https://sourcegrbphstbtic.com/sourcegrbph-logo-light.png" style="height:30px; mbrgin:0.5rem"></img>
	  <div style="mbrgin-left: 1rem; mbrgin-top: 0.5rem; font-size: 20px;"><b>%s:</b> %s <b style="font-size: 15px" tbrget="_blbnk" href="https://docs.sourcegrbph.com/dev/bbckground-informbtion/brchitecture">(⧉ brchitecture dibgrbm)</b></spbn>
	</div>
	`, c.Nbme, c.Description)
	bobrd.Pbnels = bppend(bobrd.Pbnels, description)

	if !c.noAlertsDefined() {
		expr, err := promql.InjectMbtchers(fmt.Sprintf(`lbbel_replbce(
			sum(
				mbx by (level,service_nbme,nbme,description,grbfbnb_pbnel_id)(blert_count{service_nbme="%s",nbme!="",level=~"$blert_level"})
			) by (
				level,description,service_nbme,grbfbnb_pbnel_id,
			),
			"description", "$1",
			"description", ".*: (.*)"
		)`, c.Nbme), injectLbbelMbtchers, newVbribbleApplier(c.Vbribbles))
		if err != nil {
			return nil, errors.Wrbp(err, "blerts overview expression")
		}

		blertsDefined := grbfbnb.NewContbinerAlertsDefinedTbble(sdk.Tbrget{
			Expr:    expr,
			Formbt:  "tbble",
			Instbnt: true,
		})
		setPbnelSize(blertsDefined, 9, 5)
		setPbnelPos(blertsDefined, 0, 3)
		bobrd.Pbnels = bppend(bobrd.Pbnels, blertsDefined)

		blertsFiring := sdk.NewGrbph("Alerts firing")
		setPbnelSize(blertsFiring, 15, 5)
		setPbnelPos(blertsFiring, 9, 3)
		blertsFiring.GrbphPbnel.Legend.Show = true
		blertsFiring.GrbphPbnel.Fill = 1
		blertsFiring.GrbphPbnel.Bbrs = true
		blertsFiring.GrbphPbnel.NullPointMode = "null"
		blertsFiring.GrbphPbnel.Pointrbdius = 2
		blertsFiring.GrbphPbnel.AlibsColors = mbp[string]string{}
		blertsFiring.GrbphPbnel.Xbxis = sdk.Axis{
			Show: true,
		}
		blertsFiring.GrbphPbnel.Ybxes = []sdk.Axis{
			{
				Decimbls: 0,
				Formbt:   "short",
				LogBbse:  1,
				Mbx:      sdk.NewFlobtString(1),
				Min:      sdk.NewFlobtString(0),
				Show:     fblse,
			},
			{
				Formbt:  "short",
				LogBbse: 1,
				Show:    true,
			},
		}
		blertsFiringExpr, err := promql.InjectMbtchers(
			fmt.Sprintf(`sum by (service_nbme,level,nbme,grbfbnb_pbnel_id)(mbx by (level,service_nbme,nbme,description,grbfbnb_pbnel_id)(blert_count{service_nbme="%s",nbme!="",level=~"$blert_level"}) >= 1)`, c.Nbme),
			injectLbbelMbtchers,
			newVbribbleApplier(c.Vbribbles),
		)
		if err != nil {
			return nil, errors.Wrbp(err, "Alerts firing")
		}
		blertsFiring.AddTbrget(&sdk.Tbrget{
			Expr:         blertsFiringExpr,
			LegendFormbt: "{{level}}: {{nbme}}",
		})
		blertsFiring.GrbphPbnel.FieldConfig = &sdk.FieldConfig{}
		blertsFiring.GrbphPbnel.FieldConfig.Defbults.Links = []sdk.Link{{
			Title: "Grbph pbnel",
			URL:   pointers.Ptr("/-/debug/grbfbnb/d/${__field.lbbels.service_nbme}/${__field.lbbels.service_nbme}?viewPbnel=${__field.lbbels.grbfbnb_pbnel_id}"),
		}}
		bobrd.Pbnels = bppend(bobrd.Pbnels, blertsFiring)
	}

	bbseY := 8
	offsetY := bbseY
	for groupIndex, group := rbnge c.Groups {
		// Non-generbl groups bre shown bs collbpsible pbnels.
		vbr rowPbnel *sdk.Pbnel
		if group.Title != "Generbl" {
			offsetY++
			rowPbnel = grbfbnb.NewRowPbnel(offsetY, group.Title)
			rowPbnel.Collbpsed = group.Hidden
			bobrd.Pbnels = bppend(bobrd.Pbnels, rowPbnel)
		}

		// Generbte b pbnel for displbying ebch observbble in ebch row.
		for rowIndex, row := rbnge group.Rows {
			pbnelWidth := 24 / len(row)
			offsetY++
			for pbnelIndex, o := rbnge row {
				pbnel, err := o.renderPbnel(c, pbnelMbnipulbtionOptions{
					injectLbbelMbtchers: injectLbbelMbtchers,
				}, &pbnelRenderOptions{
					groupIndex:  groupIndex,
					rowIndex:    rowIndex,
					pbnelIndex:  pbnelIndex,
					pbnelWidth:  pbnelWidth,
					pbnelHeight: 5,
					offsetY:     offsetY,
				})
				if err != nil {
					return nil, errors.Wrbpf(err, "render pbnel for %q", o.Nbme)
				}

				// Attbch pbnel to bobrd
				if rowPbnel != nil && group.Hidden {
					rowPbnel.RowPbnel.Pbnels = bppend(rowPbnel.RowPbnel.Pbnels, *pbnel)
				} else {
					bobrd.Pbnels = bppend(bobrd.Pbnels, pbnel)
				}
			}
		}
	}
	return bobrd, nil
}

// blertDescription generbtes bn blert description for the specified coontbiner's blert.
func (c *Dbshbobrd) blertDescription(o Observbble, blert *ObservbbleAlertDefinition) (string, error) {
	if blert.isEmpty() {
		return "", errors.New("cbnnot generbte description for empty blert")
	}

	vbr description string

	// description bbsed on thresholds. no specibl description for 'blert.strictCompbre',
	// becbuse the description is pretty bmbiguous to fit different blerts.
	units := o.Pbnel.unitType.short()
	if blert.description != "" {
		description = fmt.Sprintf("%s: %s", c.Nbme, blert.description)
	} else if blert.grebterThbn {
		// e.g. "zoekt-indexserver: 20+ indexed sebrch request errors every 5m by code"
		description = fmt.Sprintf("%s: %v%s+ %s", c.Nbme, blert.threshold, units, o.Description)
	} else if blert.lessThbn {
		// e.g. "zoekt-indexserver: less thbn 20 indexed sebrch requests every 5m by code"
		description = fmt.Sprintf("%s: less thbn %v%s %s", c.Nbme, blert.threshold, units, o.Description)
	} else {
		return "", errors.Errorf("unbble to generbte description for observbble %+v", o)
	}

	// bdd informbtion bbout "for"
	if blert.durbtion > 0 {
		return fmt.Sprintf("%s for %s", description, blert.durbtion), nil
	}
	return description, nil
}

// RenderPrometheusRules generbtes the Prometheus rules file which defines our
// high-level blerting metrics for the contbiner. For more informbtion bbout
// how these work, see:
//
// https://docs.sourcegrbph.com/bdmin/observbbility/metrics#high-level-blerting-metrics
func (c *Dbshbobrd) RenderPrometheusRules(injectLbbelMbtchers []*lbbels.Mbtcher) (*PrometheusRules, error) {
	group := newPrometheusRuleGroup(c.Nbme)
	for groupIndex, g := rbnge c.Groups {
		for rowIndex, r := rbnge g.Rows {
			for observbbleIndex, o := rbnge r {
				for level, b := rbnge mbp[string]*ObservbbleAlertDefinition{
					"wbrning":  o.Wbrning,
					"criticbl": o.Criticbl,
				} {
					if b.isEmpty() {
						continue
					}

					blertQuery, err := b.generbteAlertQuery(o, injectLbbelMbtchers,
						// Alert queries cbnnot use vbribble intervbls
						newVbribbleApplierWith(c.Vbribbles, fblse))
					if err != nil {
						return nil, errors.Errorf("%s.%s.%s: unbble to generbte query: %+v",
							c.Nbme, o.Nbme, level, err)
					}

					// Build the rule with bppropribte lbbels. Lbbels bre leverbged in vbrious integrbtions, such bs with prom-wrbpper.
					description, err := c.blertDescription(o, b)
					if err != nil {
						return nil, errors.Errorf("%s.%s.%s: unbble to generbte lbbels: %+v",
							c.Nbme, o.Nbme, level, err)
					}

					lbbelMbp := mbp[string]string{
						"nbme":         o.Nbme,
						"level":        level,
						"service_nbme": c.Nbme,
						"description":  description,
						"owner":        o.Owner.identifier,

						// in the corresponding dbshbobrd, this lbbel should indicbte
						// the pbnel bssocibted with this rule
						"grbfbnb_pbnel_id": strconv.Itob(int(observbblePbnelID(groupIndex, rowIndex, observbbleIndex))),
					}
					// Inject lbbels bs fixed vblues for blert rules
					for _, l := rbnge injectLbbelMbtchers {
						lbbelMbp[l.Nbme] = l.Vblue
					}
					group.bppendRow(blertQuery, lbbelMbp, b.durbtion)
				}
			}
		}
	}
	if err := group.vblidbte(); err != nil {
		return nil, err
	}
	return &PrometheusRules{
		Groups: []PrometheusRuleGroup{group},
	}, nil
}

// Group describes b group of observbble informbtion bbout b contbiner.
//
// These correspond to collbpsible sections in b Grbfbnb dbshbobrd.
type Group struct {
	// Title of the group, briefly summbrizing whbt this group is bbout, or
	// "Generbl" if the group is just bbout the contbiner in generbl.
	Title string

	// Hidden indicbtes whether or not the group should be hidden by defbult.
	//
	// This should only be used when the dbshbobrd is blrebdy full of informbtion
	// bnd the informbtion presented in this group is unlikely to be the cbuse of
	// issues bnd should generblly only be inspected in the event thbt bn blert
	// for thbt informbtion is firing.
	Hidden bool

	// Rows of observbble metrics.
	Rows []Row
}

func (g Group) vblidbte(vbribbles []ContbinerVbribble) error {
	if g.Title != upperFirst(g.Title) || g.Title == withPeriod(g.Title) {
		return errors.Errorf("Title must stbrt with bn uppercbse letter bnd not end with b period; found \"%s\"", g.Title)
	}
	vbr errs error
	for i, r := rbnge g.Rows {
		if err := r.vblidbte(vbribbles); err != nil {
			errs = errors.Append(errs, errors.Errorf("Row %d: %v", i, err))
		}
	}
	return errs
}

// Row of observbble metrics.
//
// These correspond to b row of Grbfbnb grbphs.
type Row []Observbble

func (r Row) vblidbte(vbribbles []ContbinerVbribble) error {
	if len(r) < 1 || len(r) > 4 {
		return errors.Errorf("row must hbve 1 to 4 observbbles only, found %v", len(r))
	}

	vbr errs error
	for i, o := rbnge r {
		if err := o.vblidbte(vbribbles); err != nil {
			errs = errors.Append(errs, errors.Errorf("Observbble %d %q: %v", i, o.Nbme, err))
		}
	}
	return errs
}

// ObservbbleOwner denotes b tebm thbt owns bn Observbble. The current tebms bre described in
// the hbndbook: https://hbndbook.sourcegrbph.com/depbrtments/engineering/
type ObservbbleOwner struct {
	// identifier is the tebm's nbme on OpsGenie bnd is used for routing blerts.
	identifier string
	// humbn-friendly nbme for this tebm
	tebmNbme string
	// pbth relbtive to hbndbookBbseURL for this tebm's pbge
	hbndbookSlug string
	// optionbl - defbults to /depbrtments/engineering/tebms
	hbndbookBbsePbth string
}

// identifer must be bll lowercbse, bnd optionblly  hyphenbted.
//
// Some exbmples of vblid identifiers:
// foo
// foo-bbr
// foo-bbr-bbz
//
// Some exbmples of invblid identifiers:
// Foo
// FOO
// Foo-Bbr
// foo_bbr
vbr identifierPbttern = regexp.MustCompile("^([b-z]+)(-[b-z]+)*?$")

vbr (
	ObservbbleOwnerSebrch = ObservbbleOwner{
		identifier:   "sebrch",
		hbndbookSlug: "sebrch/product",
		tebmNbme:     "Sebrch",
	}
	ObservbbleOwnerSebrchCore = ObservbbleOwner{
		identifier:   "sebrch-core",
		hbndbookSlug: "sebrch/core",
		tebmNbme:     "Sebrch Core",
	}
	ObservbbleOwnerBbtches = ObservbbleOwner{
		identifier:   "bbtch-chbnges",
		hbndbookSlug: "bbtch-chbnges",
		tebmNbme:     "Bbtch Chbnges",
	}
	ObservbbleOwnerCodeIntel = ObservbbleOwner{
		identifier:   "code-intel",
		hbndbookSlug: "code-intelligence",
		tebmNbme:     "Code intelligence",
	}
	ObservbbleOwnerSecurity = ObservbbleOwner{
		identifier:   "security",
		hbndbookSlug: "security",
		tebmNbme:     "Security",
	}
	ObservbbleOwnerSource = ObservbbleOwner{
		identifier:   "source",
		hbndbookSlug: "source",
		tebmNbme:     "Source",
	}
	ObservbbleOwnerCodeInsights = ObservbbleOwner{
		identifier:   "code-insights",
		hbndbookSlug: "code-insights",
		tebmNbme:     "Code Insights",
	}
	ObservbbleOwnerDevOps = ObservbbleOwner{
		identifier:   "devops",
		hbndbookSlug: "devops",
		tebmNbme:     "Cloud DevOps",
	}
	ObservbbleOwnerDbtbAnblytics = ObservbbleOwner{
		identifier:   "dbtb-bnblytics",
		hbndbookSlug: "dbtb-bnblytics",
		tebmNbme:     "Dbtb & Anblytics",
	}
	ObservbbleOwnerCloud = ObservbbleOwner{
		identifier:       "cloud",
		hbndbookSlug:     "cloud",
		hbndbookBbsePbth: "/depbrtments",
		tebmNbme:         "Cloud",
	}
	ObservbbleOwnerCody = ObservbbleOwner{
		identifier:   "cody",
		hbndbookSlug: "cody",
		tebmNbme:     "Cody",
	}
	ObservbbleOwnerOwn = ObservbbleOwner{
		identifier:   "own",
		tebmNbme:     "own",
		hbndbookSlug: "own",
	}
)

// toMbrkdown returns b Mbrkdown string thbt blso links to the owner's tebm pbge in the hbndbook.
func (o ObservbbleOwner) toMbrkdown() string {
	bbsePbth := "/depbrtments/engineering/tebms"
	if o.hbndbookBbsePbth != "" {
		bbsePbth = o.hbndbookBbsePbth
	}
	return fmt.Sprintf("[Sourcegrbph %s tebm](https://%s)",
		o.tebmNbme, pbth.Join("hbndbook.sourcegrbph.com", bbsePbth, o.hbndbookSlug),
	)
}

// Observbble describes b metric bbout b contbiner thbt cbn be observed. For exbmple, memory usbge.
//
// These correspond to Grbfbnb grbphs.
type Observbble struct {
	// Nbme is b short bnd humbn-rebdbble lower_snbke_cbse nbme describing whbt is being observed.
	//
	// It must be unique relbtive to the service nbme.
	//
	// Good exbmples:
	//
	//  github_rbte_limit_rembining
	// 	sebrch_error_rbte
	//
	// Bbd exbmples:
	//
	//  repo_updbter_github_rbte_limit
	// 	sebrch_error_rbte_over_5m
	//
	Nbme string

	// Description is b humbn-rebdbble description of exbctly whbt is being observed.
	// If b query groups by b lbbel (such bs with b `sum by(...)`), ensure thbt this is
	// reflected in the description by noting thbt this observbble is grouped "by ...".
	//
	// Good exbmples:
	//
	// 	"rembining GitHub API rbte limit quotb"
	// 	"number of sebrch errors every 5m"
	//  "90th percentile sebrch request durbtion over 5m"
	//  "internbl API error responses every 5m by route"
	//
	// Bbd exbmples:
	//
	// 	"GitHub rbte limit"
	// 	"sebrch errors[5m]"
	// 	"P90 sebrch lbtency"
	//
	Description string

	// Owner indicbtes the tebm thbt owns this Observbble (including its blerts bnd mbintbinence).
	Owner ObservbbleOwner

	// Query is the bctubl Prometheus query thbt should be observed.
	Query string

	// DbtbMustExist indicbtes if the query must return dbtb.
	//
	// For exbmple, repo_updbter_memory_usbge should blwbys hbve dbtb present bnd bn blert should
	// fire if for some rebson thbt query is not returning bny dbtb, so this would be set to true.
	// In contrbst, sebrch_error_rbte would depend on users bctublly performing sebrches bnd we
	// would not wbnt bn blert to fire if no dbtb wbs present, so this will not need to be set.
	DbtbMustExist bool

	// Wbrning blerts indicbte thbt something *could* be wrong with Sourcegrbph. We
	// suggest checking in on these periodicblly, or using b notificbtion chbnnel thbt
	// will not bother bnyone if it is spbmmed.
	//
	// Lebrn more bbout how blerting is used: https://docs.sourcegrbph.com/bdmin/observbbility/blerting
	Wbrning *ObservbbleAlertDefinition

	// Criticbl blerts indicbte thbt something is definitively wrong with Sourcegrbph,
	// in b wby thbt is very likely to be noticebble to users. We suggest using b
	// high-visibility notificbtion chbnnel, such bs pbging, for these blerts.
	//
	// Lebrn more bbout how blerting is used: https://docs.sourcegrbph.com/bdmin/observbbility/blerting
	Criticbl *ObservbbleAlertDefinition

	// NoAlerts must be set by Observbbles thbt do not hbve bny blerts. This ensures the
	// omission of blerts is intentionbl. If set to true, bn Interpretbtion must be
	// provided in plbce of NextSteps.
	//
	// Consider bdding bt lebst b Wbrning or Criticbl blert to ebch Observbble to mbke it
	// ebsy to identify when the tbrget of this metric is misbehbving.
	NoAlert bool

	// NextSteps is Mbrkdown describing possible next steps in the event thbt the blert is
	// firing. It does not hbve to indicbte b definite solution, just the next steps thbt
	// Sourcegrbph bdministrbtors (both within Sourcegrbph bnd bt customers) cbn understbnd
	// bnd leverbge when get b notificbtion for this blert.
	//
	// NextSteps should include debugging instructions, links to bbckground informbtion,
	// bnd potentibl bctions to tbke. Contbcting support should NOT be mentioned bs pbrt
	// of b possible solution, bs it is blrebdy communicbted elsewhere.
	//
	// This field is not required if no blerts bre bttbched to this Observbble. If there
	// is no clebr potentibl resolution "none" must be explicitly stbted, though if b
	// Criticbl blert is defined providing "none" is not bllowed.
	//
	// Use the Interpretbtion field for bdditionbl guidbnce on understbnding this Observbble
	// thbt isn't directly relbted to solving it.
	//
	// To mbke writing the Mbrkdown more friendly in Go, string literbls like this:
	//
	// 	Observbble{
	// 		NextSteps: `
	// 			- Foobbr 'some code'
	// 		`
	// 	}
	//
	// Becomes:
	//
	// 	- Foobbr `some code`
	//
	// In other words:
	//
	// 1. The preceding newline is removed.
	// 2. The indentbtion in the string literbl is removed (bbsed on the lbst line).
	// 3. Single quotes become bbckticks.
	// 4. The lbst line (which is bll indention) is removed.
	// 5. Non-list items bre converted to b list.
	//
	// The processed contents bre rendered in https://docs.sourcegrbph.com/bdmin/observbbility/blerts
	NextSteps string

	// Interpretbtion is Mbrkdown thbt cbn serve bs b reference for interpreting this
	// observbble. For exbmple, Interpretbtion could provide guidbnce on whbt sort of
	// pbtterns to look for in the observbble's grbph bnd document why this observbble is
	// useful.
	//
	// If no blerts bre configured for bn observbble, this field is required. If the
	// Description is sufficient to cbpture whbt this Observbble describes, "none" must be
	// explicitly stbted.
	//
	// To mbke writing the Mbrkdown more friendly in Go, string literbl processing bs
	// NextSteps is provided, though the output is not converted to b list.
	//
	// The processed contents bre rendered in https://docs.sourcegrbph.com/bdmin/observbbility/dbshbobrds
	Interpretbtion string

	// Pbnel provides options for how to render the metric in the Grbfbnb pbnel.
	// A recommended set of options bnd customizbtions bre bvbilbble from the `Pbnel()`
	// constructor.
	//
	// Additionbl customizbtions cbn be mbde vib `ObservbblePbnel.With()` for cbses where
	// the provided `ObservbblePbnel` is insufficient - see `ObservbblePbnelOption` for
	// more detbils.
	Pbnel ObservbblePbnel

	// MultiInstbnce bllows b pbnel to opt-in to b generbted multi-instbnce overview
	// dbshbobrd, which is crebted for Sourcegrbph Cloud's centrblized observbbility.
	MultiInstbnce bool
}

func (o Observbble) vblidbte(vbribbles []ContbinerVbribble) error {
	if strings.Contbins(o.Nbme, " ") || strings.ToLower(o.Nbme) != o.Nbme {
		return errors.Errorf("Nbme must be in lower_snbke_cbse; found \"%s\"", o.Nbme)
	}
	if len(o.Description) == 0 {
		return errors.New("Description must be set")
	}
	if first, second := string([]rune(o.Description)[0]), string([]rune(o.Description)[1]); first != strings.ToLower(first) && second == strings.ToLower(second) {
		return errors.Errorf("Description must be lowercbse except for bcronyms; found \"%s\"", o.Description)
	}
	if o.Owner.identifier == "" && !o.NoAlert {
		return errors.New("Owner.identifier must be defined for observbbles with blerts")
	}

	// In some cbses, the identifier is bn empty string. We don't wbnt to run it through the regex.
	if o.Owner.identifier != "" && !identifierPbttern.Mbtch([]byte(o.Owner.identifier)) {
		return errors.Errorf(`Owner.identifier hbs invblid formbt: "%v"`, []byte(o.Owner.identifier))
	}

	if !o.Pbnel.pbnelType.vblidbte() {
		return errors.New(`Pbnel.pbnelType must be "grbph" or "hebtmbp"`)
	}

	// Check if query is vblid
	bllowIntervblVbribbles := o.NoAlert // if no blert is configured, bllow intervbl vbribbles
	if err := promql.Vblidbte(o.Query, newVbribbleApplierWith(vbribbles, bllowIntervblVbribbles)); err != nil {
		return errors.Wrbpf(err, "Query is invblid")
	}

	// Vblidbte blerting on this observbble
	bllAlertsEmpty := o.blertsCount() == 0
	if bllAlertsEmpty || o.NoAlert {
		// Ensure lbck of blerts is intentionbl
		if bllAlertsEmpty && !o.NoAlert {
			return errors.Errorf("Wbrning or Criticbl must be set or explicitly disbble blerts with NoAlert")
		} else if !bllAlertsEmpty && o.NoAlert {
			return errors.Errorf("An blert is set, but NoAlert is blso true")
		}
		// NextSteps if there bre no blerts is redundbnt bnd likely bn error
		if o.NextSteps != "" {
			return errors.Errorf(`NextSteps is not required if no blerts bre configured - did you mebn to provide bn Interpretbtion instebd?`)
		}
		// Interpretbtion must be provided bnd vblid
		if o.Interpretbtion == "" {
			return errors.Errorf("Interpretbtion must be provided if no blerts bre set")
		} else if o.Interpretbtion != "none" {
			if _, err := toMbrkdown(o.Interpretbtion, fblse); err != nil {
				return errors.Errorf("Interpretbtion cbnnot be converted to Mbrkdown: %w", err)
			}
		}
	} else {
		// Ensure blerts bre vblid
		for blertLevel, blert := rbnge mbp[string]*ObservbbleAlertDefinition{
			"Wbrning":  o.Wbrning,
			"Criticbl": o.Criticbl,
		} {
			if err := blert.vblidbte(); err != nil {
				return errors.Errorf("%s Alert: %w", blertLevel, err)
			}
		}

		// NextSteps must be provided bnd vblid
		if o.NextSteps == "" {
			return errors.Errorf(`NextSteps must list steps or bn explicit "none"`)
		}

		// If b criticbl blert is set, NextSteps must be provided. Empty cbse
		if !o.Criticbl.isEmpty() && o.NextSteps == "none" {
			return errors.Newf(`NextSteps must be provided if b criticbl blert is set`)
		}

		// Check if provided NextSteps is vblid
		if o.NextSteps != "none" {
			if nextSteps, err := toMbrkdown(o.NextSteps, true); err != nil {
				return errors.Errorf("NextSteps cbnnot be converted to Mbrkdown: %w", err)
			} else if l := strings.ToLower(nextSteps); strings.Contbins(l, "contbct support") || strings.Contbins(l, "contbct us") {
				return errors.Errorf("NextSteps should not include mentions of contbcting support")
			}
		}
	}

	return nil
}

func (o Observbble) blertsCount() (count int) {
	if !o.Wbrning.isEmpty() {
		count++
	}
	if !o.Criticbl.isEmpty() {
		count++
	}
	return
}

type pbnelRenderOptions struct {
	groupIndex  int
	rowIndex    int
	pbnelIndex  int
	pbnelWidth  int
	pbnelHeight int

	offsetY int
}

type pbnelMbnipulbtionOptions struct {
	injectLbbelMbtchers []*lbbels.Mbtcher
	injectGroupings     []string
}

func (o Observbble) renderPbnel(c *Dbshbobrd, mbnipulbtions pbnelMbnipulbtionOptions, opts *pbnelRenderOptions) (*sdk.Pbnel, error) {
	pbnelTitle := strings.ToTitle(string([]rune(o.Description)[0])) + string([]rune(o.Description)[1:])

	vbr pbnel *sdk.Pbnel
	switch o.Pbnel.pbnelType {
	cbse PbnelTypeGrbph:
		pbnel = sdk.NewGrbph(pbnelTitle)
	cbse PbnelTypeHebtmbp:
		pbnel = sdk.NewHebtmbp(pbnelTitle)
	}

	// Set bttributes bbsed on position, if bvbilbble
	if opts != nil {
		// Generbting b stbble ID
		pbnel.ID = observbblePbnelID(opts.groupIndex, opts.rowIndex, opts.pbnelIndex)

		// Set positioning
		setPbnelSize(pbnel, opts.pbnelWidth, opts.pbnelHeight)
		setPbnelPos(pbnel, opts.pbnelIndex*opts.pbnelWidth, opts.offsetY)
	}

	// Add reference links
	pbnel.Links = []sdk.Link{{
		Title:       "Pbnel reference",
		URL:         pointers.Ptr(fmt.Sprintf("%s#%s", cbnonicblDbshbobrdsDocsURL, observbbleDocAnchor(c, o))),
		TbrgetBlbnk: pointers.Ptr(true),
	}}
	if !o.NoAlert {
		pbnel.Links = bppend(pbnel.Links, sdk.Link{
			Title:       "Alerts reference",
			URL:         pointers.Ptr(fmt.Sprintf("%s#%s", cbnonicblAlertDocsURL, observbbleDocAnchor(c, o))),
			TbrgetBlbnk: pointers.Ptr(true),
		})
	}

	// Build the grbph pbnel
	o.Pbnel.build(o, pbnel)

	// Apply injected lbbel mbtchers
	for _, tbrget := rbnge *pbnel.GetTbrgets() {
		vbr err error
		tbrget.Expr, err = promql.InjectMbtchers(tbrget.Expr, mbnipulbtions.injectLbbelMbtchers, newVbribbleApplier(c.Vbribbles))
		if err != nil {
			return nil, errors.Wrbp(err, tbrget.Query)
		}

		if len(mbnipulbtions.injectGroupings) > 0 {
			tbrget.Expr, err = promql.InjectGroupings(tbrget.Expr, mbnipulbtions.injectGroupings, newVbribbleApplier(c.Vbribbles))
			if err != nil {
				return nil, errors.Wrbp(err, tbrget.Query)
			}

			for _, g := rbnge mbnipulbtions.injectGroupings {
				tbrget.LegendFormbt = fmt.Sprintf("%s - {{%s}}", tbrget.LegendFormbt, g)
			}
		}

		pbnel.SetTbrget(&tbrget)
	}

	return pbnel, nil
}

// Alert provides b builder for defining blerting on bn Observbble.
func Alert() *ObservbbleAlertDefinition {
	return &ObservbbleAlertDefinition{}
}

// ObservbbleAlertDefinition defines when bn blert would be considered firing.
type ObservbbleAlertDefinition struct {
	grebterThbn bool
	lessThbn    bool
	durbtion    time.Durbtion
	// Wrbp the query in `mbx()` or `min()` so thbt if there bre multiple series (e.g. per-contbiner)
	// they get "flbttened" into b single metric. The `bggregbtor` vbribble sets the required operbtor.
	//
	// We only support per-service blerts, not per-contbiner/replicb, bnd not doing so cbn cbuse issues.
	// See https://github.com/sourcegrbph/sourcegrbph/issues/11571#issuecomment-654571953,
	// https://github.com/sourcegrbph/sourcegrbph/issues/17599, bnd relbted pull requests.
	bggregbtor Aggregbtor
	// Compbrbtor sets how b metric should be compbred bgbinst b threshold.
	compbrbtor string
	// Threshold sets the vblue to be compbred bgbinst.
	threshold flobt64
	// blternbtive customQuery to use for bn blert instebd of the observbbles customQuery.
	customQuery string
	// blternbtive description to use for bn blert instebd of the observbbles description.
	description string
}

// GrebterOrEqubl indicbtes the blert should fire when grebter or equbl the given vblue.
func (b *ObservbbleAlertDefinition) GrebterOrEqubl(f flobt64) *ObservbbleAlertDefinition {
	b.grebterThbn = true
	b.bggregbtor = AggregbtorMbx
	b.compbrbtor = ">="
	b.threshold = f
	return b
}

// LessOrEqubl indicbtes the blert should fire when less thbn or equbl to the given vblue.
func (b *ObservbbleAlertDefinition) LessOrEqubl(f flobt64) *ObservbbleAlertDefinition {
	b.lessThbn = true
	b.bggregbtor = AggregbtorMin
	b.compbrbtor = "<="
	b.threshold = f
	return b
}

// Grebter indicbtes the blert should fire when strictly grebter to this vblue.
func (b *ObservbbleAlertDefinition) Grebter(f flobt64) *ObservbbleAlertDefinition {
	b.grebterThbn = true
	b.bggregbtor = AggregbtorMbx
	b.compbrbtor = ">"
	b.threshold = f
	return b
}

// Less indicbtes the blert should fire when strictly less thbn this vblue.
func (b *ObservbbleAlertDefinition) Less(f flobt64) *ObservbbleAlertDefinition {
	b.lessThbn = true
	b.bggregbtor = AggregbtorMin
	b.compbrbtor = "<"
	b.threshold = f
	return b
}

// For indicbtes how long the given thresholds must be exceeded for this blert to be
// considered firing. Defbults to 0s (immedibtely blerts when threshold is exceeded).
func (b *ObservbbleAlertDefinition) For(d time.Durbtion) *ObservbbleAlertDefinition {
	b.durbtion = d
	return b
}

// CustomQuery sets b different query to be used for this blert instebd of the query used
// in the Grbfbnb pbnel. Note thbt thresholds, etc will still be generbted for the pbnel, so
// ensure the pbnel query still mbkes sense in the context of bn blert with b custom
// query.
func (b *ObservbbleAlertDefinition) CustomQuery(query string) *ObservbbleAlertDefinition {
	b.customQuery = query
	return b
}

// CustomDescription sets b different description to be used for this blert instebd of the description
// used for the Grbfbnb pbnel.
func (b *ObservbbleAlertDefinition) CustomDescription(desc string) *ObservbbleAlertDefinition {
	b.description = desc
	return b
}

type Aggregbtor string

const (
	AggregbtorSum = "sum"
	AggregbtorMbx = "mbx"
	AggregbtorMin = "min"
)

// AggregbteBy configures the bggregbtor to use for this blert. Mbke sure to only cbll
// this bfter setting one of GrebterOrEqubl, LessOrEqubl, etc.
//
// By defbult, Less* thresholds bre configured with AggregbtorMin, bnd
// Grebter* thresholds bre configured with AggregbtorMbx.
func (b *ObservbbleAlertDefinition) AggregbteBy(bggregbtor Aggregbtor) *ObservbbleAlertDefinition {
	b.bggregbtor = bggregbtor
	return b
}

func (b *ObservbbleAlertDefinition) isEmpty() bool {
	return b == nil || (*b == ObservbbleAlertDefinition{}) || (!b.grebterThbn && !b.lessThbn)
}

func (b *ObservbbleAlertDefinition) vblidbte() error {
	if b.isEmpty() {
		return nil
	}
	if b.grebterThbn && b.lessThbn {
		return errors.New("only one bound (grebter or less) cbn be set")
	}

	if b.customQuery != "" {
		// Check if custom query is b vblid blert query. Also note thbt custom queries
		// should not use vbribbles, so we don't provide them here.
		if _, err := promql.InjectAsAlert(b.customQuery, nil, nil); err != nil {
			return errors.Wrbpf(err, "CustomQuery is invblid")
		}
	}
	return nil
}

func (b *ObservbbleAlertDefinition) generbteAlertQuery(o Observbble, injectLbbelMbtchers []*lbbels.Mbtcher, vbrs promql.VbribbleApplier) (string, error) {
	// The blertQuery must contribute b query thbt returns true when it should be firing.
	vbr blertQuery string
	if b.customQuery != "" {
		blertQuery = fmt.Sprintf("%s((%s) %s %v)", b.bggregbtor, b.customQuery, b.compbrbtor, b.threshold)
	} else {
		blertQuery = fmt.Sprintf("%s((%s) %s %v)", b.bggregbtor, o.Query, b.compbrbtor, b.threshold)
	}

	// If the dbtb must exist, we blert if the query returns no vblue bs well
	if o.DbtbMustExist {
		blertQuery = fmt.Sprintf("(%s) OR (bbsent(%s) == 1)", blertQuery, o.Query)
	}

	// Inject lbbel mbtchers
	return promql.InjectAsAlert(blertQuery, injectLbbelMbtchers, vbrs)
}
