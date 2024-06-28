package service

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
)

func newCodyGatewayEventsService(logger log.Logger, config *codygatewayevents.ServiceBigQueryOptions) *codygatewayevents.Service {
	if config == nil {
		logger.Warn("CodyGatewayEvents service is not configured")
		return nil
	}
	logger.Info("CodyGatewayEvents service is configured",
		log.String("projectID", config.ProjectID),
		log.String("dataset", config.Dataset),
		log.String("eventsTable", config.EventsTable))
	return codygatewayevents.NewService(codygatewayevents.ServiceOptions{
		BigQuery: *config,
	})
}
