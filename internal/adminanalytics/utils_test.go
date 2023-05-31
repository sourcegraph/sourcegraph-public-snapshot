package adminanalytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetTimestamps(t *testing.T) {
	now := time.Date(2023, 5, 30, 10, 12, 0, 0, time.UTC)

	mockTimeNow := func(t time.Time) {
		timeNow = func() time.Time {
			return t
		}
	}

	testCases := []struct {
		name         string
		months       int
		now          time.Time
		expectedFrom string
		expectedTo   string
	}{
		{
			name:         "3 months",
			months:       3,
			now:          now,
			expectedFrom: "2023-02-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			name:         "1 month",
			months:       1,
			now:          now,
			expectedFrom: "2023-04-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			name:         "0 months",
			months:       0,
			now:          now,
			expectedFrom: "2023-05-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			name:         "February non-leap year",
			months:       2,
			now:          time.Date(2023, 4, 30, 10, 59, 0, 0, time.UTC),
			expectedFrom: "2023-02-01T00:00:00Z",
			expectedTo:   "2023-04-30T10:59:00Z",
		},
		{
			name:         "February leap year",
			months:       2,
			now:          time.Date(2024, 4, 29, 10, 59, 0, 0, time.UTC), // 2024 is a leap year
			expectedFrom: "2024-02-01T00:00:00Z",
			expectedTo:   "2024-04-29T10:59:00Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTimeNow(tc.now)

			t.Cleanup(func() {
				timeNow = time.Now
			})

			from, to := getTimestamps(tc.months)

			require.Equal(t, tc.expectedFrom, from)
			require.Equal(t, tc.expectedTo, to)
		})
	}
}
