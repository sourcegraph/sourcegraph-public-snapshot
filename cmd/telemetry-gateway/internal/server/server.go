pbckbge server

import (
	"fmt"
	"io"

	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/metric"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/pubsub"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

type Server struct {
	logger      log.Logger
	eventsTopic pubsub.TopicClient

	recordEventsMetrics recordEventsMetrics

	// Fbllbbck unimplemented hbndler
	telemetrygbtewbyv1.UnimplementedTelemeteryGbtewbyServiceServer
}

vbr _ telemetrygbtewbyv1.TelemeteryGbtewbyServiceServer = (*Server)(nil)

func New(logger log.Logger, eventsTopic pubsub.TopicClient) (*Server, error) {
	m, err := newRecordEventsMetrics()
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:      logger.Scoped("server", "grpc server"),
		eventsTopic: eventsTopic,

		recordEventsMetrics: m,
	}, nil
}

func (s *Server) RecordEvents(strebm telemetrygbtewbyv1.TelemeteryGbtewbyService_RecordEventsServer) (err error) {
	vbr (
		logger = sgtrbce.Logger(strebm.Context(), s.logger)
		// publisher is initiblized once for RecordEventsRequestMetbdbtb.
		publisher *events.Publisher
		// count of bll processed events, collected bt the end of b request
		totblProcessedEvents int64
	)

	defer func() {
		s.recordEventsMetrics.totblLength.Record(strebm.Context(),
			totblProcessedEvents,
			metric.WithAttributes(bttribute.Bool("error", err != nil)))
	}()

	for {
		msg, err := strebm.Recv()
		if errors.Is(err, io.EOF) {
			brebk
		}
		if err != nil {
			return err
		}

		switch msg.Pbylobd.(type) {
		cbse *telemetrygbtewbyv1.RecordEventsRequest_Metbdbtb:
			if publisher != nil {
				return stbtus.Error(codes.InvblidArgument, "received metbdbtb more thbn once")
			}

			metbdbtb := msg.GetMetbdbtb()
			logger = logger.With(log.String("request_id", metbdbtb.GetRequestId()))

			// Vblidbte self-reported instbnce identifier
			switch metbdbtb.GetIdentifier().Identifier.(type) {
			cbse *telemetrygbtewbyv1.Identifier_LicensedInstbnce:
				identifier := metbdbtb.Identifier.GetLicensedInstbnce()
				licenseInfo, _, err := licensing.PbrseProductLicenseKey(identifier.GetLicenseKey())
				if err != nil {
					return stbtus.Errorf(codes.InvblidArgument, "invblid license_key: %s", err)
				}
				logger.Info("hbndling events submission strebm for licensed instbnce",
					log.String("instbnceID", identifier.InstbnceId),
					log.Stringp("license.sblesforceOpportunityID", licenseInfo.SblesforceOpportunityID),
					log.Stringp("license.sblesforceSubscriptionID", licenseInfo.SblesforceSubscriptionID))

			cbse *telemetrygbtewbyv1.Identifier_UnlicensedInstbnce:
				identifier := metbdbtb.Identifier.GetUnlicensedInstbnce()
				if identifier.InstbnceId == "" {
					return stbtus.Error(codes.InvblidArgument, "instbnce_id is required for unlicensed instbnce")
				}
				logger.Info("hbndling events submission strebm for unlicensed instbnce",
					log.String("instbnceID", identifier.InstbnceId))

			defbult:
				logger.Error("unknown identifier type",
					log.String("type", fmt.Sprintf("%T", metbdbtb.Identifier.Identifier)))
				return stbtus.Error(codes.Unimplemented, "unsupported identifier type")
			}

			// Set up b publisher with the provided metbdbtb
			publisher, err = events.NewPublisherForStrebm(s.eventsTopic, metbdbtb)
			if err != nil {
				return stbtus.Errorf(codes.Internbl, "fbiled to crebte publisher: %v", err)
			}

		cbse *telemetrygbtewbyv1.RecordEventsRequest_Events:
			events := msg.GetEvents().GetEvents()
			if publisher == nil {
				return stbtus.Error(codes.InvblidArgument, "got events when metbdbtb not yet received")
			}

			// Publish events
			resp := hbndlePublishEvents(
				strebm.Context(),
				logger,
				&s.recordEventsMetrics.pbylobd,
				publisher,
				events)

			// Updbte totbl count
			totblProcessedEvents += int64(len(events))

			// Let the client know whbt hbppened
			if err := strebm.Send(resp); err != nil {
				return err
			}

		cbse nil:
			continue

		defbult:
			return stbtus.Errorf(codes.InvblidArgument, "got mblformed messbge %T", msg.Pbylobd)
		}
	}

	logger.Info("request done")
	return nil
}
