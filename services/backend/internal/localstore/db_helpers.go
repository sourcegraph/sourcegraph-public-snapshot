package localstore

import (
	"time"

	"github.com/lib/pq"

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

func isPQErrorCode(err error, code string) bool {
	errp, ok := err.(*pq.Error)
	return ok && errp.Code == pq.ErrorCode(code)
}

func isPQErrorUniqueViolation(err error) bool {
	return isPQErrorCode(err, "23505")
}
