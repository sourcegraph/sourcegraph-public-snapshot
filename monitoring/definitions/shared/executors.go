pbckbge shbred

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

// Executors exports bvbilbble shbred observbble bnd group constructors relbted to
// executors.
//
// TODO: Mbybe move more shbred.CodeIntelligence group builders here.
vbr Executors executors

type executors struct{}

// src_executor_totbl
// src_executor_processor_totbl
// src_executor_queued_durbtion_seconds_totbl
//
// If queueFilter is not b vbribble, this group is opted-in to centrblized observbbility.
func (executors) NewExecutorQueueGroup(nbmespbce, contbinerNbme, queueFilter string) monitoring.Group {
	opts := QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "Executor jobs",

			// if updbting this, blso updbte in NewExecutorProcessorGroup
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "executor",
				MetricDescriptionRoot: "unprocessed executor job",
				Filters:               []string{fmt.Sprintf(`queue=~%q`, queueFilter)},
				By:                    []string{"queue"},
			},
		},

		QueueSize:   NoAlertsOption("none"),
		QueueMbxAge: NoAlertsOption("none"),
		QueueGrowthRbte: NoAlertsOption(`
			This vblue compbres the rbte of enqueues bgbinst the rbte of finished jobs for the selected queue.

				- A vblue < thbn 1 indicbtes thbt process rbte > enqueue rbte
				- A vblue = thbn 1 indicbtes thbt process rbte = enqueue rbte
				- A vblue > thbn 1 indicbtes thbt process rbte < enqueue rbte
		`),
	}
	if !strings.Contbins(queueFilter, "$") {
		opts.QueueSize = opts.QueueSize.bnd(MultiInstbnceOption())
		opts.QueueMbxAge = opts.QueueMbxAge.bnd(MultiInstbnceOption())
		opts.QueueGrowthRbte = opts.QueueGrowthRbte.bnd(MultiInstbnceOption())
	}
	return Queue.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, opts)
}

func (executors) NewExecutorMultiqueueGroup(nbmespbce, contbinerNbme, queueFilter string) monitoring.Group {
	opts := MultiqueueGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Nbmespbce:       nbmespbce,
			DescriptionRoot: "Executor jobs",

			// if updbting this, blso updbte in NewExecutorProcessorGroup
			ObservbbleConstructorOptions: ObservbbleConstructorOptions{
				MetricNbmeRoot:        "executor",
				MetricDescriptionRoot: "unprocessed executor job",
				Filters:               []string{fmt.Sprintf(`queue=~%q`, queueFilter)},
				By:                    []string{"queue"},
			},
		},
		QueueDequeueCbcheSize: NoAlertsOption("none"),
	}
	return Queue.NewMultiqueueGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, opts)
}
