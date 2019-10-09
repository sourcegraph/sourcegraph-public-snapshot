package a8n

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

type ChangesetCounts struct {
	Time                 time.Time
	Total                int32
	Merged               int32
	Closed               int32
	Open                 int32
	OpenApproved         int32
	OpenChangesRequested int32
	OpenPending          int32
}

type Event interface {
	Timestamp() time.Time
	Type() a8n.ChangesetEventKind
	Changeset() int64
}

func CalcCounts(start, end time.Time, cs []*a8n.Changeset, es ...Event) ([]*ChangesetCounts, error) {
	counts := []*ChangesetCounts{}
	for t := end; t.After(start) || t.Equal(start); t = t.Add(-24 * time.Hour) {
		counts = append(counts, &ChangesetCounts{Time: t})
	}

	reversed := make([]*ChangesetCounts, 0, len(counts))
	for i := len(counts) - 1; i >= 0; i-- {
		reversed = append(reversed, counts[i])
	}

	return reversed, nil
}
