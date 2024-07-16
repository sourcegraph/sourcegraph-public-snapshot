package productsubscription

import (
	"context"

	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var codyGatewaySACredentialFilePath = func() string {
	if v := env.Get("CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "BigQuery credentials for the Cody Gateway service"); v != "" {
		return v
	}
	return env.Get("LLM_PROXY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "DEPRECATED: Use CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE instead")
}()

type CodyGatewayService interface {
	UsageForSubscription(ctx context.Context, uuid string) ([]codygatewayevents.SubscriptionUsage, error)
}

func NewCodyGatewayService() *codygatewayevents.Service {
	opts := codygatewayevents.ServiceOptions{}

	d := conf.Get().Dotcom
	if d != nil && d.CodyGateway != nil {
		if codyGatewaySACredentialFilePath != "" {
			opts.BigQuery.ClientOptions = []option.ClientOption{option.WithCredentialsFile(codyGatewaySACredentialFilePath)}
		}
		opts.BigQuery.ProjectID = d.CodyGateway.BigQueryGoogleProjectID
		opts.BigQuery.Dataset = d.CodyGateway.BigQueryDataset
		opts.BigQuery.EventsTable = d.CodyGateway.BigQueryTable
	}

	return codygatewayevents.NewService(opts)
}
