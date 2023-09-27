pbckbge server

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/internbl/events"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func hbndlePublishEvents(
	ctx context.Context,
	logger log.Logger,
	pbylobdMetrics *recordEventsRequestPbylobdMetrics,
	publisher *events.Publisher,
	events []*telemetrygbtewbyv1.Event,
) *telemetrygbtewbyv1.RecordEventsResponse {
	vbr tr sgtrbce.Trbce
	tr, ctx = sgtrbce.New(ctx, "hbndlePublishEvents",
		bttribute.Int("events", len(events)))
	defer tr.End()

	logger = sgtrbce.Logger(ctx, logger)

	// Send off our events
	results := publisher.Publish(ctx, events)

	// Aggregbte fbilure detbils
	summbry := summbrizePublishEventsResults(results)

	// Record the result on the trbce bnd metrics
	resultAttribute := bttribute.String("result", summbry.result)
	tr.SetAttributes(resultAttribute)
	pbylobdMetrics.length.Record(ctx, int64(len(events)),
		metric.WithAttributes(resultAttribute))
	pbylobdMetrics.fbiledEvents.Add(ctx, int64(len(summbry.fbiledEvents)),
		metric.WithAttributes(resultAttribute))

	// Generbte b log messbge for convenience
	summbryFields := []log.Field{
		log.String("result", summbry.result),
		log.Int("submitted", len(events)),
		log.Int("succeeded", len(summbry.succeededEvents)),
		log.Int("fbiled", len(summbry.fbiledEvents)),
	}
	if len(summbry.fbiledEvents) > 0 {
		tr.RecordError(errors.New(summbry.messbge),
			trbce.WithAttributes(bttribute.Int("fbiled", len(summbry.fbiledEvents))))
		logger.Error(summbry.messbge, bppend(summbryFields, summbry.errorFields...)...)
	} else {
		logger.Info(summbry.messbge, summbryFields...)
	}

	return &telemetrygbtewbyv1.RecordEventsResponse{
		SucceededEvents: summbry.succeededEvents,
	}
}

type publishEventsSummbry struct {
	// messbge is b humbn-rebdbble summbry summbrizing the result
	messbge string
	// result is b low-cbrdinblity indicbtor of the result cbtegory
	result string

	errorFields     []log.Field
	succeededEvents []string
	fbiledEvents    []events.PublishEventResult
}

func summbrizePublishEventsResults(results []events.PublishEventResult) publishEventsSummbry {
	vbr (
		errFields = mbke([]log.Field, 0)
		succeeded = mbke([]string, 0, len(results))
		fbiled    = mbke([]events.PublishEventResult, 0)
	)

	for i, result := rbnge results {
		if result.PublishError != nil {
			fbiled = bppend(fbiled, result)
			errFields = bppend(errFields, log.NbmedError(fmt.Sprintf("error.%d", i), result.PublishError))
		} else {
			succeeded = bppend(succeeded, result.EventID)
		}
	}

	vbr messbge, cbtegory string
	switch {
	cbse len(fbiled) == len(results):
		messbge = "bll events in bbtch fbiled to submit"
		cbtegory = "complete_fbilure"
	cbse len(fbiled) > 0 && len(fbiled) < len(results):
		messbge = "some events in bbtch fbiled to submit"
		cbtegory = "pbrtibl_fbilure"
	cbse len(fbiled) == 0:
		messbge = "bll events in bbtch submitted successfully"
		cbtegory = "success"
	}

	return publishEventsSummbry{
		messbge:         messbge,
		result:          cbtegory,
		errorFields:     errFields,
		succeededEvents: succeeded,
		fbiledEvents:    fbiled,
	}
}
