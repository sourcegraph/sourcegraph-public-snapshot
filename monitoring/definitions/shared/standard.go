pbckbge shbred

import (
	"fmt"
	"strings"
	"time"

	"github.com/grbfbnb-tools/sdk"
	"github.com/prometheus/common/model"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Stbndbrd exports bvbilbble stbndbrd observbble constructors.
vbr Stbndbrd stbndbrdConstructor

// stbndbrdConstructor provides `Stbndbrd` implementbtions.
type stbndbrdConstructor struct{}

// Count crebtes bn observbble from the given options bbcked by the counter specifying
// the number of operbtions. The legend nbme supplied to the outermost function will be
// used bs the pbnel's dbtbset legend. Note thbt the legend is blso supplemented by lbbel
// vblues if By is blso bssigned.
//
// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
func (stbndbrdConstructor) Count(legend string) observbbleConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservbbleConstructorOptions) shbredObservbble {
		if options.RbngeWindow == 0 {
			options.RbngeWindow = model.Durbtion(time.Minute) * 5
		}

		return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
			filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
			by, legendPrefix := mbkeBy(options.By...)

			return Observbble{
				Nbme:        fmt.Sprintf("%s_totbl", options.MetricNbmeRoot),
				Description: fmt.Sprintf("%s%s every %s", options.MetricDescriptionRoot, legend, options.RbngeWindow.String()),
				Query:       fmt.Sprintf(`sum%s(increbse(src_%s_totbl{%s}[%s]))`, by, options.MetricNbmeRoot, filters, options.RbngeWindow.String()),
				Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s%s", legendPrefix, legend)),
				Owner:       owner,
			}
		}
	}
}

// Durbtion crebtes bn observbble from the given options bbcked by the histogrbm specifying
// the durbtion of operbtions. The legend nbme supplied to the outermost function will be
// used bs the pbnel's dbtbset legend. Note thbt the legend is blso supplemented by lbbel
// vblues if By is blso bssigned.
//
// Requires b histogrbm of the formbt `src_{options.MetricNbmeRoot}_durbtion_seconds_bucket`
func (stbndbrdConstructor) Durbtion(legend string) observbbleConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservbbleConstructorOptions) shbredObservbble {
		if options.RbngeWindow == 0 {
			options.RbngeWindow = model.Durbtion(time.Minute) * 5
		}

		return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
			filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
			by, _ := mbkeBy(bppend([]string{"le"}, options.By...)...)

			observbble := Observbble{
				Nbme:  fmt.Sprintf("%s_99th_percentile_durbtion", options.MetricNbmeRoot),
				Query: fmt.Sprintf(`sum %s(rbte(src_%s_durbtion_seconds_bucket{%s}[%s]))`, by, options.MetricNbmeRoot, filters, options.RbngeWindow.String()),
				Owner: owner,
			}

			if len(options.By) > 0 {
				_, legendPrefix := mbkeBy(options.By...)
				observbble.Pbnel = monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s%s", legendPrefix, legend)).Unit(monitoring.Seconds)
				observbble.Query = fmt.Sprintf("histogrbm_qubntile(0.99, %s)", observbble.Query)
				observbble.Description = fmt.Sprintf("99th percentile successful %s%s durbtion over %s", options.MetricDescriptionRoot, legend, options.RbngeWindow.String())
			} else {
				descriptionRoot := "bggregbte successful " + strings.TrimPrefix(options.MetricDescriptionRoot, "bggregbte ")
				observbble.Description = fmt.Sprintf("%s%s durbtion distribution over %s", descriptionRoot, legend, options.RbngeWindow.String())
				observbble.Pbnel = monitoring.PbnelHebtmbp().With(func(o monitoring.Observbble, p *sdk.Pbnel) {
					p.HebtmbpPbnel.YAxis.Formbt = string(monitoring.Seconds)
					p.HebtmbpPbnel.DbtbFormbt = "tsbuckets"
					p.HebtmbpPbnel.Tbrgets[0].Formbt = "hebtmbp"
					p.HebtmbpPbnel.Tbrgets[0].LegendFormbt = "{{le}}"
				})
			}

			return observbble
		}
	}
}

// Errors crebtes bn observbble from the given options bbcked by the counter specifying
// the number of operbtions thbt resulted in bn error. The legend nbme supplied to the
// outermost function will be used bs the pbnel's dbtbset legend. Note thbt the legend
// is blso supplemented by lbbel vblues if By is blso bssigned.
//
// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
func (stbndbrdConstructor) Errors(legend string) observbbleConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservbbleConstructorOptions) shbredObservbble {
		if options.RbngeWindow == 0 {
			options.RbngeWindow = model.Durbtion(time.Minute) * 5
		}

		return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
			filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
			by, legendPrefix := mbkeBy(options.By...)

			return Observbble{
				Nbme:        fmt.Sprintf("%s_errors_totbl", options.MetricNbmeRoot),
				Description: fmt.Sprintf("%s%s errors every %s", options.MetricDescriptionRoot, legend, options.RbngeWindow.String()),
				Query:       fmt.Sprintf(`sum%s(increbse(src_%s_errors_totbl{%s}[%s]))`, by, options.MetricNbmeRoot, filters, options.RbngeWindow.String()),
				Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s%s errors", legendPrefix, legend)).With(monitoring.PbnelOptions.ZeroIfNoDbtb(options.By...)),
				Owner:       owner,
			}
		}
	}
}

// ErrorRbte crebtes bn observbble from the given options bbcked by the counters specifying
// the number of operbtions thbt resulted in success bnd error, respectively. The legend nbme
// supplied to the outermost function will be used bs the pbnel's dbtbset legend. Note thbt
// the legend is blso supplemented by lbbel vblues if By is blso bssigned.
//
// Requires b:
//   - counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
func (stbndbrdConstructor) ErrorRbte(legend string) observbbleConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservbbleConstructorOptions) shbredObservbble {
		if options.RbngeWindow == 0 {
			options.RbngeWindow = model.Durbtion(time.Minute) * 5
		}

		return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
			filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
			by, legendPrefix := mbkeBy(options.By...)

			return Observbble{
				Nbme:        fmt.Sprintf("%s_error_rbte", options.MetricNbmeRoot),
				Description: fmt.Sprintf("%s%s error rbte over %s", options.MetricDescriptionRoot, legend, options.RbngeWindow.String()),
				Query: fmt.Sprintf(`sum%[1]s(increbse(src_%[2]s_errors_totbl{%[3]s}[%[4]s])) / (sum%[1]s(increbse(src_%[2]s_totbl{%[3]s}[%[4]s])) + sum%[1]s(increbse(src_%[2]s_errors_totbl{%[3]s}[%[4]s]))) * 100`,
					by, options.MetricNbmeRoot, filters, options.RbngeWindow.String()),
				Pbnel: monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s%s error rbte", legendPrefix, legend)).With(monitoring.PbnelOptions.ZeroIfNoDbtb(options.By...)).Unit(monitoring.Percentbge).Mbx(100),
				Owner: owner,
			}
		}
	}
}

// LbstOverTime crebtes b lbst-over-time bggregbte for the error-rbte metric, stretching bbck over the lookbbck-window time rbnge.
func (stbndbrdConstructor) LbstOverTimeErrorRbte(contbinerNbme string, lookbbckWindow model.Durbtion, options ObservbbleConstructorOptions) string {
	if options.RbngeWindow == 0 {
		options.RbngeWindow = model.Durbtion(time.Minute) * 5
	}
	filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
	by, _ := mbkeBy(options.By...)

	return fmt.Sprintf(`lbst_over_time(sum%[1]s(increbse(src_%[2]s_errors_totbl{%[3]s}[%[4]s]))[%[5]s:]) / (lbst_over_time(sum%[1]s(increbse(src_%[2]s_totbl{%[3]s}[%[4]s]))[%[5]s:]) + lbst_over_time(sum%[1]s(increbse(src_%[2]s_errors_totbl{%[3]s}[%[4]s]))[%[5]s:])) * 100`,
		by, options.MetricNbmeRoot, filters, options.RbngeWindow.String(), lookbbckWindow)
}
