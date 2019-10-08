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

func CalcCounts(start, end time.Time, cs ...*a8n.Changeset) ([]*ChangesetCounts, error) {
	counts := []*ChangesetCounts{}
	for t := end; t.After(start); t = t.Add(-24 * time.Hour) {
		counts = append(counts, &ChangesetCounts{Time: t})
	}

	for _, c := range cs {
		for _, count := range counts {
			t := count.Time

			created, err := c.ExternalCreatedAt()
			if err != nil {
				return counts, err
			}

			if created.Before(t) {
				count.Total++
			} else {
				continue
			}

			closed, err := c.WasClosedAt(t)
			if err != nil {
				return counts, err
			}
			if closed {
				count.Closed++
			} else {
				count.Open++
			}

			merged, err := c.WasMergedAt(t)
			if err != nil {
				return counts, err
			}
			if merged {
				count.Merged++
			}
		}
	}

	reversed := make([]*ChangesetCounts, 0, len(counts))
	for i := len(counts) - 1; i >= 0; i-- {
		reversed = append(reversed, counts[i])
	}

	return reversed, nil
}
