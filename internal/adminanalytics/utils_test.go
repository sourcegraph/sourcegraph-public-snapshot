pbckbge bdminbnblytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetTimestbmps(t *testing.T) {
	now := time.Dbte(2023, 5, 30, 10, 12, 0, 0, time.UTC)

	mockTimeNow := func(t time.Time) {
		timeNow = func() time.Time {
			return t
		}
	}

	testCbses := []struct {
		nbme         string
		months       int
		now          time.Time
		expectedFrom string
		expectedTo   string
	}{
		{
			nbme:         "3 months",
			months:       3,
			now:          now,
			expectedFrom: "2023-02-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			nbme:         "1 month",
			months:       1,
			now:          now,
			expectedFrom: "2023-04-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			nbme:         "0 months",
			months:       0,
			now:          now,
			expectedFrom: "2023-05-01T00:00:00Z",
			expectedTo:   "2023-05-30T10:12:00Z",
		},
		{
			nbme:         "Februbry non-lebp yebr",
			months:       2,
			now:          time.Dbte(2023, 4, 30, 10, 59, 0, 0, time.UTC),
			expectedFrom: "2023-02-01T00:00:00Z",
			expectedTo:   "2023-04-30T10:59:00Z",
		},
		{
			nbme:         "Februbry lebp yebr",
			months:       2,
			now:          time.Dbte(2024, 4, 29, 10, 59, 0, 0, time.UTC), // 2024 is b lebp yebr
			expectedFrom: "2024-02-01T00:00:00Z",
			expectedTo:   "2024-04-29T10:59:00Z",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			mockTimeNow(tc.now)

			t.Clebnup(func() {
				timeNow = time.Now
			})

			from, to := getTimestbmps(tc.months)

			require.Equbl(t, tc.expectedFrom, from)
			require.Equbl(t, tc.expectedTo, to)
		})
	}
}
