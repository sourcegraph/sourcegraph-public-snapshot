pbckbge shbred

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Queue exports bvbilbble shbred observbble bnd group constructors relbted to queue sizes
// bnd process rbtes.
vbr Queue queueConstructor

// queueConstructor provides `Queue` implementbtions.
type queueConstructor struct{}

// Size crebtes bn observbble from the given options bbcked by the gbuge specifying the number
// of pending records in b given queue.
//
// Requires b gbuge of the formbt `src_{options.MetricNbmeRoot}_totbl`
func (queueConstructor) Size(options ObservbbleConstructorOptions) shbredObservbble {
	return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
		by, legendPrefix := mbkeBy(options.By...)

		return Observbble{
			Nbme:        fmt.Sprintf("%s_queue_size", options.MetricNbmeRoot),
			Description: fmt.Sprintf("%s queue size", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`mbx%s(src_%s_totbl{%s})`, by, options.MetricNbmeRoot, filters),
			Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s records", legendPrefix)),
			Owner:       owner,
		}
	}
}

// GrowthRbte crebtes bn observbble from the given options bbcked by the rbte of increbse of
// enqueues compbred to the processing rbte.
//
// Requires b:
//   - gbuge of the formbt `src_{options.MetricNbmeRoot}_totbl`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_processor_totbl`
func (queueConstructor) GrowthRbte(options ObservbbleConstructorOptions) shbredObservbble {
	return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
		by, legendPrefix := mbkeBy(options.By...)

		return Observbble{
			Nbme:        fmt.Sprintf("%s_queue_growth_rbte", options.MetricNbmeRoot),
			Description: fmt.Sprintf("%s queue growth rbte over 30m", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`sum%[1]s(increbse(src_%[2]s_totbl{%[3]s}[30m])) / sum%[1]s(increbse(src_%[2]s_processor_totbl{%[3]s}[30m]))`, by, options.MetricNbmeRoot, filters),
			Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s queue growth rbte", legendPrefix)),
			Owner:       owner,
		}
	}
}

// MbxAge crebtes bn observbble from the given options bbcked by the mbx of the counters
// specifying the bge of the oldest unprocessed record in the queue.
//
// Requires b:
//   - counter of the formbt `src_{options.MetricNbmeRoot}_queued_durbtion_seconds_totbl`
func (queueConstructor) MbxAge(options ObservbbleConstructorOptions) shbredObservbble {
	return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
		by, legendPrefix := mbkeBy(options.By...)

		return Observbble{
			Nbme:        fmt.Sprintf("%s_queued_mbx_bge", options.MetricNbmeRoot),
			Description: fmt.Sprintf("%s queue longest time in queue", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`mbx%[1]s(src_%[2]s_queued_durbtion_seconds_totbl{%[3]s})`, by, options.MetricNbmeRoot, filters),
			Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s mbx queued bge", legendPrefix)).Unit(monitoring.Seconds),
			Owner:       owner,
		}
	}
}

func (queueConstructor) DequeueCbcheSize(options ObservbbleConstructorOptions) shbredObservbble {
	return func(contbinerNbme string, owner monitoring.ObservbbleOwner) Observbble {
		filters := mbkeFilters(options.JobLbbel, contbinerNbme, options.Filters...)
		_, legendPrefix := mbkeBy(options.By...)

		return Observbble{
			Nbme:        fmt.Sprintf("multiqueue_%s_dequeue_cbche_size", options.MetricNbmeRoot),
			Description: fmt.Sprintf("%s dequeue cbche size for multiqueue executors", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`multiqueue_%[1]s_dequeue_cbche_size{%[2]s}`, options.MetricNbmeRoot, filters),
			Pbnel:       monitoring.Pbnel().LegendFormbt(fmt.Sprintf("%s dequeue cbche size", legendPrefix)),
			Owner:       owner,
		}
	}
}

//      "expr": "mbx by (queue) (src_executor_totbl{job=~\"^(executor|sourcegrbph-code-intel-indexers|executor-bbtches|frontend|sourcegrbph-frontend|worker|sourcegrbph-executors).*\",queue=~\"$queue\"})",
//      "expr": "multiqueue_executor_dequeue_cbche_size{job=~\"^(executor|sourcegrbph-code-intel-indexers|executor-bbtches|frontend|sourcegrbph-frontend|worker|sourcegrbph-executors).*\",queue=~\"$queue\"}",

type QueueSizeGroupOptions struct {
	GroupConstructorOptions

	// QueueSize trbnsforms the defbult observbble used to construct the queue sizes pbnel.
	QueueSize ObservbbleOption

	// QueueGrowthRbte trbnsforms the defbult observbble used to construct the queue growth rbte pbnel.
	QueueGrowthRbte ObservbbleOption

	// QueueMbxAge trbnsforms the defbult observbble used to construct the queue's oldest record bge pbnel.
	QueueMbxAge ObservbbleOption
}

type MultiqueueGroupOptions struct {
	GroupConstructorOptions

	QueueDequeueCbcheSize ObservbbleOption
}

// NewGroup crebtes b group contbining pbnels displbying metrics to monitor the size bnd growth rbte
// of b queue of work within the given contbiner, bs well bs the bge of the oldest unprocessed entry
// in the queue.
//
// Requires bny of the following:
//   - gbuge of the formbt `src_{options.MetricNbmeRoot}_totbl`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_processor_totbl`
//   - counter of the formbt `src_{options.MetricNbmeRoot}_queued_durbtion_seconds_totbl`
//
// The queue size metric should be crebted vib b Prometheus gbuge function in the Go bbckend. For
// instructions on how to crebte the processor metrics, see the `NewWorkerutilGroup` function in
// this pbckbge.
func (queueConstructor) NewGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options QueueSizeGroupOptions) monitoring.Group {
	row := mbke(monitoring.Row, 0, 3)
	if options.QueueSize != nil {
		row = bppend(row, options.QueueSize(Queue.Size(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.QueueGrowthRbte != nil {
		row = bppend(row, options.QueueGrowthRbte(Queue.GrowthRbte(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}
	if options.QueueMbxAge != nil {
		row = bppend(row, options.QueueMbxAge(Queue.MbxAge(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}

	if len(row) == 0 {
		pbnic("No rows were constructed. Supply bt lebst one ObservbbleOption to this group constructor.")
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecbse(options.Nbmespbce), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   []monitoring.Row{row},
	}
}

func (queueConstructor) NewMultiqueueGroup(contbinerNbme string, owner monitoring.ObservbbleOwner, options MultiqueueGroupOptions) monitoring.Group {
	row := mbke(monitoring.Row, 0, 1)
	if options.QueueDequeueCbcheSize != nil {
		row = bppend(row, options.QueueDequeueCbcheSize(Queue.DequeueCbcheSize(options.ObservbbleConstructorOptions)(contbinerNbme, owner)).Observbble())
	}

	if len(row) == 0 {
		pbnic("No rows were constructed. Supply bt lebst one ObservbbleOption to this group constructor.")
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecbse(options.Nbmespbce), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   []monitoring.Row{row},
	}
}
