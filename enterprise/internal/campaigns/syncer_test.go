package campaigns

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNextSync(t *testing.T) {
	clock := func() time.Time { return time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC) }
	type args struct {
		lastSync   time.Time
		lastChange time.Time
	}
	tests := []struct {
		name       string
		clock      func() time.Time
		lastSync   time.Time
		lastChange time.Time
		want       time.Time
	}{
		{
			name:       "No time passed",
			clock:      clock,
			lastSync:   clock(),
			lastChange: clock(),
			want:       clock().Add(minSyncDelay),
		},
		{
			name:       "Linear backoff",
			clock:      clock,
			lastSync:   clock(),
			lastChange: clock().Add(-1 * time.Hour),
			want:       clock().Add(1 * time.Hour),
		},
		{
			// Could happen due to clock skew
			name:       "Future change",
			clock:      clock,
			lastSync:   clock(),
			lastChange: clock().Add(1 * time.Hour),
			want:       clock().Add(minSyncDelay),
		},
		{
			name:       "Diff max is capped",
			clock:      clock,
			lastSync:   clock(),
			lastChange: clock().Add(-2 * maxSyncDelay),
			want:       clock().Add(maxSyncDelay),
		},
		{
			name:       "Diff min is capped",
			clock:      clock,
			lastSync:   clock(),
			lastChange: clock().Add(-1 * minSyncDelay / 2),
			want:       clock().Add(minSyncDelay),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextSync(tt.clock, tt.lastSync, tt.lastChange)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
