package syncjobs

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type recordsReader struct {
	// readOnlyCache is a replaceable abstraction over rcache.Cache.
	readOnlyCache interface {
		ListKeys(ctx context.Context) ([]string, error)
		GetMulti(keys ...string) [][]byte
	}
}

func NewRecordsReader() *recordsReader {
	return &recordsReader{
		readOnlyCache: rcache.New(syncJobsRecordsPrefix),
	}
}

// Get retrieves a record by timestamp.
func (r *recordsReader) Get(timestamp time.Time) (*Status, error) {
	res := r.readOnlyCache.GetMulti(strconv.FormatInt(timestamp.UTC().UnixNano(), 10))
	if len(res) == 0 || len(res[0]) == 0 {
		return nil, errors.New("record not found")
	}
	var s Status
	if err := json.Unmarshal(res[0], &s); err != nil {
		return nil, errors.Wrap(err, "invalid record")
	}
	return &s, nil
}

// GetAll retrieves the first n records, with the most recent records first.
func (r *recordsReader) GetAll(ctx context.Context, first int) ([]Status, error) {
	keys, err := r.readOnlyCache.ListKeys(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list jobs")
	}

	// keys are timestamps
	sort.Strings(keys)

	switch {
	case first <= 0:
		return []Status{}, nil
	case first < len(keys):
		keys = keys[:first]
	}

	// get values
	vals := r.readOnlyCache.GetMulti(keys...)
	records := make([]Status, 0, len(vals))
	for _, v := range vals {
		var j Status
		if err := json.Unmarshal(v, &j); err != nil {
			continue // discard
		}
		records = append(records, j)
	}

	return records, nil
}
