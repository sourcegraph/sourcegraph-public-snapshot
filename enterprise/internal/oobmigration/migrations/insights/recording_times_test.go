package insights

import (
	"testing"
	"time"

	"github.com/hexops/autogold"
)

func Test_calculateRecordingTimes(t *testing.T) {
	defaultInterval := timeInterval{
		unit:  month,
		value: 2,
	}
	createdAt := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	defaultTimes := buildRecordingTimes(12, defaultInterval, createdAt)

	testCases := []struct {
		lastRecordedAt time.Time
		interval       timeInterval
		existingTimes  []time.Time
		want           autogold.Value
	}{
		{
			interval: defaultInterval,
			want:     autogold.Want("no existing times returns all generated times", defaultTimes),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			calculated := calculateRecordingTimes(createdAt, tc.lastRecordedAt, tc.interval, []time.Time{})
			tc.want.Equal(t, calculated)
		})
	}
}
