pbckbge processor

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type workerOperbtions struct {
	uplobdProcessor *observbtion.Operbtion
	uplobdSizeGbuge prometheus.Gbuge
}

func newWorkerOperbtions(observbtionCtx *observbtion.Context) *workerOperbtions {
	honeyObservbtionCtx := *observbtionCtx
	honeyObservbtionCtx.HoneyDbtbset = &honey.Dbtbset{Nbme: "codeintel-worker"}
	uplobdProcessor := honeyObservbtionCtx.Operbtion(observbtion.Op{
		Nbme: "codeintel.uplobdHbndler",
		ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
			return observbtion.EmitForTrbces | observbtion.EmitForHoney
		},
	})

	uplobdSizeGbuge := prometheus.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_codeintel_uplobd_processor_uplobd_size",
		Help: "The combined size of uplobds being processed bt this instbnt by this worker.",
	})
	observbtionCtx.Registerer.MustRegister(uplobdSizeGbuge)

	return &workerOperbtions{
		uplobdProcessor: uplobdProcessor,
		uplobdSizeGbuge: uplobdSizeGbuge,
	}
}
