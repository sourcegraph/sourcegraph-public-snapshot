package bigquerywriter

import (
	"context"
	"io"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Writer adds helpers for best practices around using a bigquery.Inserter.
//
// On service shutdown, (*Writer).Close() must be called.
type Writer struct {
	client io.Closer
	// Inserter is the underlying bigquery.Inserter that can be used if you
	// know what you are doing. Use with care!
	Inserter *bigquery.Inserter
}

func New(client *bigquery.Client, dataset, table string) *Writer {
	return &Writer{
		client:   client,
		Inserter: client.Dataset(dataset).Table(table).Inserter(),
	}
}

// Write will insert all events into the underlying table inserter using the
// bigquery.ValueSaver pattern, where inserted values must implement
// bigquery.ValueSaver to match their configured schema.
//
// The underlying Put returns a bigquery.PutMultiError if one or more rows failed
// to be uploaded. The PutMultiError contains a RowInsertionError for each failed row.
//
// The underlying Put will retry on temporary errors (see
// https://cloud.google.com/bigquery/troubleshooting-errors). This can result
// in duplicate rows if you do not use insert IDs. Also, if the error persists,
// the call will run indefinitely. Pass a context with a timeout to prevent
// hanging calls.
func (w *Writer) Write(ctx context.Context, values ...bigquery.ValueSaver) error {
	if len(values) == 0 {
		return errors.New("no values to insert")
	}
	// If only one value, provide just the first value, to save the underlying
	// inserter.Put from using reflection.
	if len(values) == 1 {
		return w.Inserter.Put(ctx, values[0])
	}
	// Otherwise, insert all values.
	return w.Inserter.Put(ctx, values)
}

// Close closes any underlying resources.
func (w *Writer) Close() error {
	return w.client.Close()
}
