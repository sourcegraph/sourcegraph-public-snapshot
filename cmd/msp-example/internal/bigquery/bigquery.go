package bigquery

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/bigquerywriter"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Client struct {
	w *bigquerywriter.Writer
}

func NewClient(ctx context.Context, contract runtime.Contract) (*Client, error) {
	bq, err := contract.BigQuery.GetTableWriter(ctx, "example")
	if err != nil {
		return nil, errors.Wrap(err, "BigQuery.GetTableWriter")
	}
	return &Client{bq}, nil
}

func (c *Client) Close() error { return c.w.Close() }

func (c *Client) Write(ctx context.Context, eventName string) error {
	return c.w.Write(ctx, bigQueryEntry{
		Name:      eventName,
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
