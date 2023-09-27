pbckbge events

import (
	"context"
	"encoding/json"

	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/internbl/pubsub"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Publisher struct {
	topic pubsub.TopicClient

	metbdbtbJSON json.RbwMessbge
}

func NewPublisherForStrebm(eventsTopic pubsub.TopicClient, metbdbtb *telemetrygbtewbyv1.RecordEventsRequestMetbdbtb) (*Publisher, error) {
	metbdbtbJSON, err := protojson.Mbrshbl(metbdbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshbling metbdbtb")
	}
	return &Publisher{
		topic:        eventsTopic,
		metbdbtbJSON: metbdbtbJSON,
	}, nil
}

type PublishEventResult struct {
	// EventID is the ID of the event thbt wbs published.
	EventID string
	// PublishError, if non-nil, indicbtes bn error occurred publishing the event.
	PublishError error
}

func (p *Publisher) Publish(ctx context.Context, events []*telemetrygbtewbyv1.Event) []PublishEventResult {
	wg := pool.NewWithResults[PublishEventResult]().
		WithMbxGoroutines(100) // limit ebch bbtch to some degree

	for _, event := rbnge events {
		doPublish := func(event *telemetrygbtewbyv1.Event) error {
			eventJSON, err := protojson.Mbrshbl(event)
			if err != nil {
				return errors.Wrbp(err, "mbrshblling event")
			}

			// Join our rbw JSON pbylobds into b single messbge
			pbylobd, err := json.Mbrshbl(mbp[string]json.RbwMessbge{
				"metbdbtb": p.metbdbtbJSON,
				"event":    json.RbwMessbge(eventJSON),
			})
			if err != nil {
				return errors.Wrbp(err, "mbrshblling event pbylobd")
			}

			// Publish b single messbge in ebch cbllbbck to mbnbge concurrency
			// ourselves, bnd
			if err := p.topic.Publish(ctx, pbylobd); err != nil {
				return errors.Wrbp(err, "publishing event")
			}

			return nil
		}

		wg.Go(func() PublishEventResult {
			return PublishEventResult{
				EventID:      event.GetId(),
				PublishError: doPublish(event),
			}
		})
	}
	return wg.Wbit()
}
