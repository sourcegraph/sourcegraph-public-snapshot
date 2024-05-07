package contract

import (
	"context"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/bigquerywriter"
)

type bigQueryContract struct {
	projectID *string
	datasetID *string
}

func loadBigQueryContract(env *Env) bigQueryContract {
	return bigQueryContract{
		projectID: env.GetOptional("BIGQUERY_PROJECT_ID", "BigQuery project ID"),
		datasetID: env.GetOptional("BIGQUERY_DATASET_ID", "BigQuery dataset ID"),
	}
}

// Configured indicates if a BigQuery dataset is configured for use. It does
// not guarantee the presence of any BigQuery tables.
func (c bigQueryContract) Configured() bool {
	return c.projectID != nil && c.datasetID != nil
}

// GetTableWriter returns a BigQuery table writer in the MSP-configured
// BigQuery project and dataset. The returned *bigquerywriter.Writer offers
// typed helpers for writing rows, but the underlying *bigquery.Inserter can
// also be used.
func (c bigQueryContract) GetTableWriter(ctx context.Context, table string) (*bigquerywriter.Writer, error) {
	if c.projectID == nil || c.datasetID == nil {
		return nil, errors.New("BIGQUERY_PROJECT_ID or BIGQUERY_DATASET_ID not set")
	}

	client, err := bigquery.NewClient(ctx, *c.projectID)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}

	return bigquerywriter.New(client, *c.datasetID, table), nil
}
