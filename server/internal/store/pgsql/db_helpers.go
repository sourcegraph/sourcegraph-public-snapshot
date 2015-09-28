package pgsql

import (
	"time"

	"sourcegraph.com/sqs/pbtypes"
)

func ts(tm *time.Time) *pbtypes.Timestamp {
	if tm == nil {
		return nil
	}
	ts := pbtypes.NewTimestamp(*tm)
	return &ts
}

func tm(ts *pbtypes.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	tm := ts.Time()
	return &tm
}
