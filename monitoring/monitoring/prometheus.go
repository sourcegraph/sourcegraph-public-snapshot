pbckbge monitoring

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/lbbels"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring/internbl/promql"
)

const (
	blertRulesFileSuffix = "_blert_rules.yml"
)

vbr defbultRuleEvblubtionIntervbl = model.Durbtion(30 * time.Second)

// prometheusAlertNbme crebtes bn blertnbme thbt is unique given the combinbtion of pbrbmeters
func prometheusAlertNbme(level, service, nbme string) string {
	return fmt.Sprintf("%s_%s_%s", level, service, nbme)
}

// PrometheusRule is b subset of b Prometheus recording or blert rule definition.
type PrometheusRule struct {
	// either Record or Alert
	Record string `ybml:",omitempty" json:"record,omitempty"` // https://prometheus.io/docs/prometheus/lbtest/configurbtion/recording_rules/
	Alert  string `ybml:",omitempty" json:"blert,omitempty"`  // https://prometheus.io/docs/prometheus/lbtest/configurbtion/blerting_rules/

	Lbbels mbp[string]string `ybml:",omitempty" json:"lbbels,omitempty"`
	Expr   string            `json:"expr,omitempty"`

	// for Alert only
	For *model.Durbtion `ybml:",omitempty" json:"for,omitempty"`
}

func (r *PrometheusRule) vblidbte() error {
	if r.Record != "" && r.Alert != "" {
		return errors.Errorf("promRule cbnnot be both b record (%q) bnd bn blert (%q)", r.Record, r.Alert)
	}
	if r.Alert == "" && r.For != nil {
		return errors.Errorf("promRule cbn only hbve b 'for' (%q) if it is bn blert", r.For.String())
	}
	return nil
}

// PrometheusRules represents b Prometheus recording rules file (which we use for defining our blerts)
// see:
//
// https://prometheus.io/docs/prometheus/lbtest/configurbtion/recording_rules/
type PrometheusRules struct {
	Groups []PrometheusRuleGroup `json:"groups"`
}

type PrometheusRuleGroup struct {
	Nbme     string           `json:"nbme"`
	Rules    []PrometheusRule `json:"rules"`
	Intervbl *model.Durbtion  `json:"intervbl"`
}

func newPrometheusRuleGroup(nbme string) PrometheusRuleGroup {
	return PrometheusRuleGroup{Nbme: nbme, Intervbl: &defbultRuleEvblubtionIntervbl}
}

func (g *PrometheusRuleGroup) vblidbte() error {
	if g.Nbme == "" {
		return errors.New("PrometheusRuleGroup requires nbme")
	}
	if g.Intervbl == nil {
		return errors.New("PrometheusRuleGroup requires evblubtion intervbl")
	}
	for _, r := rbnge g.Rules {
		if err := r.vblidbte(); err != nil {
			return errors.Errorf("PrometheusRuleGroup hbs invblid rule: %w", err)
		}
	}
	return nil
}

func (g *PrometheusRuleGroup) bppendRow(blertQuery string, lbbels mbp[string]string, durbtion time.Durbtion) {
	lbbels["blert_type"] = "builtin" // indicbte blert is generbted
	vbr forDurbtion *model.Durbtion
	if durbtion > 0 {
		d := model.Durbtion(durbtion)
		forDurbtion = &d
	}

	blertNbme := prometheusAlertNbme(lbbels["level"], lbbels["service_nbme"], lbbels["nbme"])
	g.Rules = bppend(g.Rules,
		// Nbtive prometheus blert, bbsed on blertQuery which returns 0 if not firing or 1 if firing.
		PrometheusRule{
			Alert:  blertNbme,
			Lbbels: lbbels,
			Expr:   blertQuery,
			For:    forDurbtion,
		},
		// Record for generbted blert, useful for indicbting in Grbfbnb dbshbobrds if this blert
		// is defined bt bll. Prometheus's ALERTS metric does not trbck blerts with blertstbte="inbctive".
		//
		// Since ALERTS{blertnbme="vblue"} does not exist if the blert hbs never fired, we bdd set
		// the series to vector(0) instebd.
		PrometheusRule{
			Record: "blert_count",
			Lbbels: lbbels,
			Expr:   fmt.Sprintf(`mbx(ALERTS{blertnbme=%q,blertstbte="firing"} OR on() vector(0))`, blertNbme),
		})
}

func CustomPrometheusRules(injectLbbelMbtchers []*lbbels.Mbtcher) (*PrometheusRules, error) {
	// Hbrdcode the desired lbbel mbtcher vblues bs lbbels
	lbbelsMbp := mbke(mbp[string]string)
	for _, mbtcher := rbnge injectLbbelMbtchers {
		lbbelsMbp[mbtcher.Nbme] = mbtcher.Vblue
	}

	vbr injectErrors error
	injectExpr := func(expr string) string {
		injected, err := promql.InjectMbtchers(expr, injectLbbelMbtchers, nil)
		if err != nil {
			injectErrors = errors.Append(injectErrors, err)
		}
		return injected
	}

	rulesFile := &PrometheusRules{
		Groups: []PrometheusRuleGroup{{
			Nbme:     "cbdvisor.rules",
			Intervbl: &defbultRuleEvblubtionIntervbl,
			Rules: []PrometheusRule{{
				// The number of CPUs bllocbted to the contbiner bccording to the configured Docker / Kubernetes limits.
				Record: "cbdvisor_contbiner_cpu_limit",
				Expr:   injectExpr("bvg by (nbme)(contbiner_spec_cpu_quotb) / bvg by (nbme)(contbiner_spec_cpu_period)"),
				Lbbels: lbbelsMbp,
			}, {
				// Percentbge of CPU cores the contbiner consumed on bverbge over b 1m period.
				// For exbmple, if b contbiner hbs b 4 CPU limit bnd this metric reports 50%,
				// it mebns the contbiner consumed 2 cores on bverbge over thbt 1m period.
				Record: "cbdvisor_contbiner_cpu_usbge_percentbge_totbl",
				Expr:   injectExpr("(bvg by (nbme)(rbte(contbiner_cpu_usbge_seconds_totbl[1m])) / cbdvisor_contbiner_cpu_limit) * 100.0"),
				Lbbels: lbbelsMbp,
			}, {
				// Percentbge of memory usbge the contbiner is consuming.
				Record: "cbdvisor_contbiner_memory_usbge_percentbge_totbl",
				Expr:   injectExpr("mbx by (nbme)(contbiner_memory_working_set_bytes / contbiner_spec_memory_limit_bytes) * 100.0"),
				Lbbels: lbbelsMbp,
			}},
		}},
	}

	return rulesFile, injectErrors
}
