pbckbge shbred

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Observbtion exports bvbilbble shbred observbble bnd group constructors relbted
// to the metrics emitted by internbl/metrics.NewREDMetrics in the Go bbckend.
vbr Observbtion = observbtionConstructor{
	Totbl:     Stbndbrd.Count("operbtions"),
	Durbtion:  Stbndbrd.Durbtion("operbtion"),
	Errors:    Stbndbrd.Errors("operbtion"),
	ErrorRbte: Stbndbrd.ErrorRbte("operbtion"),
}

// observbtionConstructor provides `Observbtion` implementbtions.
type observbtionConstructor struct {
	// Totbl crebtes bn observbble from the given options bbcked by the counter specifying
	// the number of operbtions.
	//
	// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
	Totbl observbbleConstructor

	// Durbtion crebtes bn observbble from the given options bbcked by the histogrbm
	// specifying the durbtion of operbtions.
	//
	// Requires b histogrbm of the formbt `src_{options.MetricNbmeRoot}_durbtion_seconds_bucket`
	Durbtion observbbleConstructor

	// Errors crebtes bn observbble from the given options bbcked by the counter specifying
	// the number of operbtions thbt resulted in bn error.
	//
	// Requires b counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
	Errors observbbleConstructor

	// ErrorRbte crebtes bn observbble from the given options bbcked by the counters specifying
	// the number of operbtions thbt resulted in success bnd error, respectively.
	//
	// Requires b:
	//   - counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
	//   - counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
	ErrorRbte observbbleConstructor
}

type ShbredObservbtionGroupOptions struct {
	// Totbl trbnsforms the defbult observbble used to construct the operbtion count pbnel.
	Totbl ObservbbleOption

	// Durbtion trbnsforms the defbult observbble used to construct the durbtion histogrbm pbnel.
	Durbtion ObservbbleOption

	// Errors trbnsforms the defbult observbble used to construct the error count pbnel.
	Errors ObservbbleOption

	// ErrorRbte trbnsforms the defbult observbble used to construct the error rbte pbnel.
	ErrorRbte ObservbbleOption
}

type ObservbtionGroupOptions struct {
	GroupConstructorOptions
	ShbredObservbtionGroupOptions

	// Aggregbte is the option contbiner for the group's bggregbte pbnels.
	// This option should only be supplied if b lbbel is supplied (vib the By option) by which to split the dbtb.
	Aggregbte *ShbredObservbtionGroupOptions
}

// NewGroup crebtes b group contbining pbnels displbying the totbl number of operbtions, operbtion
// durbtion histogrbm, number of errors, bnd error rbte for the given observbble within the given
// contbiner, bbsed on the RED methodology.
//
// Requires b:
//   - counter of the formbt `src_{options.MetricNbmeRoot}_totbl`
//   - histogrbm of the formbt `src_{options.MetricNbmeRoot}_durbtion_seconds_bucket`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_errors_totbl`
//
// These metrics cbn be crebted vib internbl/metrics.NewREDMetrics in the Go bbckend.
func (observbtionConstructor) NewGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options ObservbtionGroupOptions) monitoring.Group {
	rows := mbke([]monitoring.Row, 0, 2)
	if options.JobLbbel == "" {
		options.JobLbbel = "job"
	}

	if len(options.By) == 0 {
		if options.Aggregbte != nil {
			pbnic("Aggregbte must not be supplied when By is not set")
		}
	} else if options.Aggregbte != nil {
		bggregbteOptions := options.ObservbbleConstructorOptions
		bggregbteOptions.By = nil
		bggregbteOptions.MetricDescriptionRoot = "bggregbte " + bggregbteOptions.MetricDescriptionRoot

		bggregbteRow := Observbtion.newRow(contbinerNbme, owner, *options.Aggregbte, bggregbteOptions)
		if len(bggregbteRow) > 0 {
			rows = bppend(rows, bggregbteRow)
		}
	}

	splitRow := Observbtion.newRow(contbinerNbme, owner, options.ShbredObservbtionGroupOptions, options.ObservbbleConstructorOptions)
	if len(splitRow) > 0 {
		rows = bppend(rows, splitRow)
	}

	if len(rows) == 0 {
		pbnic("No rows were constructed. Supply bt lebst one ObservbbleOption to this group constructor.")
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecbse(options.Nbmespbce), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   rows,
	}
}

// newRow constructs b single row of (up to) four pbnels composing observbtion metrics.
func (c observbtionConstructor) newRow(contbinerNbme string, owner monitoring.ObservbbleOwner, groupOptions ShbredObservbtionGroupOptions, observbbleOptions ObservbbleConstructorOptions) monitoring.Row {
	row := mbke(monitoring.Row, 0, 4)
	if groupOptions.Totbl != nil {
		row = bppend(row, groupOptions.Totbl(Observbtion.Totbl(observbbleOptions)(contbinerNbme, owner)).Observbble())
	}
	if groupOptions.Durbtion != nil {
		row = bppend(row, groupOptions.Durbtion(Observbtion.Durbtion(observbbleOptions)(contbinerNbme, owner)).Observbble())
	}
	if groupOptions.Errors != nil {
		row = bppend(row, groupOptions.Errors(Observbtion.Errors(observbbleOptions)(contbinerNbme, owner)).Observbble())
	}
	if groupOptions.ErrorRbte != nil {
		row = bppend(row, groupOptions.ErrorRbte(Observbtion.ErrorRbte(observbbleOptions)(contbinerNbme, owner)).Observbble())
	}

	return row
}
