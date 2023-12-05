package example

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

func writeBigQueryEvent(ctx context.Context, contract runtime.Contract, eventName string) error {
	bq, err := contract.BigQuery.GetTableWriter(ctx, "example")
	if err != nil {
		return errors.Wrap(err, "BigQuery.GetTableWriter")
	}

	return bq.Write(ctx, bigQueryEntry{
		Name:      "service.started",
		CreatedAt: time.Now(),
	})
}

// bigQueryEntry is based on the schema:
//
//	[{
//		"name": "name",
//		"type": "STRING",
//		"mode": "REQUIRED",
//		"description": "The name of the event"
//	},
//	{
//		"name": "metadata",
//		"type": "JSON",
//		"mode": "NULLABLE",
//		"description": "The event-specific metadata"
//	},
//	{
//		"name": "created_at",
//		"type": "TIMESTAMP",
//		"mode": "REQUIRED",
//		"description": "The event capture time"
//	}]
type bigQueryEntry struct {
	Name      string
	Metadata  map[string]any
	CreatedAt time.Time
}

func (e bigQueryEntry) Save() (map[string]bigquery.Value, string, error) {
	row := map[string]bigquery.Value{
		"name":       e.Name,
		"created_at": e.CreatedAt,
	}
	if e.Metadata != nil {
		metadata, err := json.Marshal(e.Metadata)
		if err != nil {
			return nil, "", errors.Wrap(err, "marshal metadata")
		}
		row["metadata"] = string(metadata)
	}
	return row, "", nil
}
