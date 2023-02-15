package syncjobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type recordsReader struct {
	// readOnlyCache is a replaceable abstraction over rcache.Cache.
	readOnlyCache interface {
		Slice(ctx context.Context, from, to int) ([][]byte, error)
	}
}

func NewRecordsReader(limit int) *recordsReader {
	return &recordsReader{
		// The cache is read-only in recordsReader, so the limit doesn't affect
		// the contents of the list - it doesn't need to align with the actual
		// limit of the list.
		readOnlyCache: rcache.NewFIFOList(syncJobsRecordsKey, limit),
	}
}

func (r *recordsReader) Get(ctx context.Context, timestamp time.Time) (*Status, error) {
	items, err := r.GetAll(ctx, -1)
	if err != nil {
		return nil, errors.Wrap(err, "list jobs")
	}
	for _, i := range items {
		if i.Completed.Equal(timestamp) {
			return &i, nil
		}
	}
	return nil, errors.New("job not found")
}

// GetAll retrieves the first n records, with the most recent records first.
func (r *recordsReader) GetAll(ctx context.Context, first int) ([]Status, error) {
	items, err := r.readOnlyCache.Slice(ctx, 0, first)
	if err != nil {
		return nil, errors.Wrap(err, "list jobs")
	}

	switch {
	case first <= 0:
		return []Status{}, nil
	case first < len(items):
		items = items[:first]
	}

	// get values
	records := make([]Status, 0, len(items))
	for _, v := range items {
		var j Status
		if err := json.Unmarshal(v, &j); err != nil {
			continue // discard
		}
		records = append(records, j)
	}

	// records are already ~sorted
	return records, nil
}
