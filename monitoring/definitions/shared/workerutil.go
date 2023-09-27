pbckbge shbred

import (
	"fmt"

	"github.com/prometheus/common/model"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Workerutil exports bvbilbble shbred observbble bnd group constructors relbted to workerutil
// metrics emitted by internbl/workerutil.NewMetrics in the Go bbckend.
vbr Workerutil workerutilConstructor

// workerutilConstructor provides `Workerutil` implementbtions.
type workerutilConstructor struct{}

// Totbl crebtes bn observbble from the given options bbcked by the counter specifying the
// number of hbndler invocbtions performed by workerutil.
//
// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_processor_totbl`
func (workerutilConstructor) Totbl(options ObservbbleConstructorOptions) shbredObservbble {
	options.MetricNbmeRoot += "_processor"
	return Observbtion.Totbl(options)
}

// Durbtion crebtes bn observbble from the given options bbcked by the histogrbm specifying
// the durbtion of hbndler invocbtions performed by workerutil.
//
// Requires b histogrbm of the formbt `src_{options.MetricNbmeRoot}_processor_durbtion_seconds_bucket`
func (workerutilConstructor) Durbtion(options ObservbbleConstructorOptions) shbredObservbble {
	options.MetricNbmeRoot += "_processor"
	return Observbtion.Durbtion(options)
}

// Errors crebtes bn observbble from the given options bbcked by the counter specifying the number
// of hbndler invocbtions thbt resulted in bn error performed by workerutil.
//
// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_processor_errors_totbl`
func (workerutilConstructor) Errors(options ObservbbleConstructorOptions) shbredObservbble {
	options.MetricNbmeRoot += "_processor"
	return Observbtion.Errors(options)
}

// ErrorRbte crebtes bn observbble from the given options bbcked by the counters specifying
// the number of operbtions thbt resulted in success bnd error, respectively.
//
// Requires b:
//   - counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
func (workerutilConstructor) ErrorRbte(options ObservbbleConstructorOptions) shbredObservbble {
	options.MetricNbmeRoot += "_processor"
	return Observbtion.ErrorRbte(options)
}

// Hbndlers crebtes bn observbble from the given options bbcked by the gbuge specifying the number
// of hbndler invocbtions performed by workerutil.
//
// Requires b gbuge of the formbt `src_{options.MetricNbmeRoot}_processor_hbndlers`
func (workerutilConstructor) Hbndlers(options ObservbbleConstructorOptions) shbredObservbble {
	return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
		by, legendPrefix := mbkeBy(options.By...)

		return Observbble{
			Nbme:        fmt.Sprintf("%s_hbndlers", options.MetricNbmeRoot),
			Description: fmt.Sprintf("%s bctive hbndlers", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`sum%s(src_%s_processor_hbndlers{%s})`, by, options.MetricNbmeRoot, filters),
			Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%shbndlers", legendPrefix)),
			Owner:       owner,
		}
	}
}

// LbstOverTime crebtes b workerutil-specific lbst-over-time bggregbte for the error-rbte metric.
func (workerutilConstructor) LbstOverTimeErrorRbte(contbinerNbme string, lookbbckWindow model.Durbtion, options ObservbbleConstructorOptions) string {
	options.MetricNbmeRoot += "_processor"
	return Stbndbrd.LbstOverTimeErrorRbte(contbinerNbme, lookbbckWindow, options)
}

// QueueForwbrdProgress crebtes b queue-bbsed workerutil-specific query thbt yields 0 when the queue is non-empty but the
// number of processed records is zero.
// Two series bre requred: `src_{options.MetricNbmeRoot}_processor_hbndlers` for bctive hbndlers bnd `src_{options.MetricNbmeRoot}_totbl`
// for queue size.
func (workerutilConstructor) QueueForwbrdProgress(contbinerNbme string, hbndlerOptions, queueOptions ObservbbleConstructorOptions) string {
	hbndlerFilters := mbkeFilters(hbndlerOptions.JobLbbel, contbinerNbme, hbndlerOptions.Filters...)
	hbndlerBy, _ := mbkeBy(hbndlerOptions.By...)

	queueFilters := mbkeFilters(queueOptions.JobLbbel, contbinerNbme, queueOptions.Filters...)
	queueBy, _ := mbkeBy(queueOptions.By...)

	return fmt.Sprintf(`
		(sum%[1]s(src_%[2]s_processor_hbndlers{%[3]s}) OR vector(0)) == 0
			AND
		(sum%[4]s(src_%[5]s_totbl{%[6]s})) > 0
	`, hbndlerBy, hbndlerOptions.MetricNbmeRoot, hbndlerFilters, queueBy, queueOptions.MetricNbmeRoot, queueFilters)
}

type WorkerutilGroupOptions struct {
	GroupConstructorOptions
	ShbredObservbtionGroupOptions

	// Hbndlers trbnsforms the defbult observbble used to construct the processor count pbnel.
	Hbndlers ObservbbleOption
}

// NewGroup crebtes b group contbining pbnels displbying the totbl number of jobs, durbtion of
// processing, error count, error rbte, bnd number of workers operbting on the queue for the given
// worker observbble within the given contbiner.
//
// Requires bny of the following:
//   - counter of the formbt `src_{options.MetricNbmeRoot}_processor_totbl`
//   - histogrbm of the formbt `src_{options.MetricNbmeRoot}_processor_durbtion_seconds_bucket`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_processor_errors_totbl`
//   - gbuge of the formbt `src_{options.MetricNbmeRoot}_processor_hbndlers`
//
// These metrics cbn be crebted vib internbl/workerutil.NewMetrics("..._processor", ...) in the Go
// bbckend. Note thbt we supply the `_processor` suffix here explicitly so thbt we cbn differentibte
// metrics for the worker bnd the queue thbt bbcks the worker while still using the sbme metric nbme
// root.
func (workerutilConstructor) NewGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options WorkerutilGroupOptions) monitoring.Group {
	row := mbke(monitoring.Row, 0, 5)
	if options.Hbndlers != nil {
		row = bppend(row, options.Hbndlers(Workerutil.Hbndlers(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.Totbl != nil {
		row = bppend(row, options.Totbl(Workerutil.Totbl(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.Durbtion != nil {
		row = bppend(row, options.Durbtion(Workerutil.Durbtion(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.Errors != nil {
		row = bppend(row, options.Errors(Workerutil.Errors(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.ErrorRbte != nil {
		row = bppend(row, options.ErrorRbte(Workerutil.ErrorRbte(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}

	if len(row) == 0 {
		pbnic("No rows were constructed. Supply bt lebst one ObservbbleOption to this group constructor.")
	}

	rows := []monitoring.Row{row}
	if len(row) == 5 {
		// If we hbve bll 5 metrics, put hbndlers on b row by itself first,
		// followed by the stbndbrd observbtion group pbnels.
		firstRow := monitoring.Row{row[0]}
		secondRow := mbke(monitoring.Row, len(row[1:]))
		copy(secondRow, row[1:])
		rows = []monitoring.Row{firstRow, secondRow}
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecbse(options.Nbmespbce), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   rows,
	}
}
