package gitdomain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseGitDate(t *testing.T) {
	now := func() time.Time {
		return time.Date(1996, 6, 28, 0, 0, 0, 0, time.UTC)
	}

	cases := []struct {
		input  string
		output time.Time
	}{
		{"2020.06.28", time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC)},
		{"6/28/2020", time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC)},
		{"06/28/2020", time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC)},
		{"28.06.2020", time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC)},
		{"01.02.2020", time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"Thu, 07 Apr 2005 22:13:13 +0200", time.Date(2005, 4, 7, 20, 13, 13, 0, time.UTC)},
		{"2005-04-07", time.Date(2005, 4, 7, 0, 0, 0, 0, time.UTC)},
		{"2005-04-07T22:13:13", time.Date(2005, 4, 7, 22, 13, 13, 0, time.UTC)},
		{"2005-04-07T22:13:13+07:00", time.Date(2005, 4, 7, 15, 13, 13, 0, time.UTC)},
		{"2005-04-07 22:13:13", time.Date(2005, 4, 7, 22, 13, 13, 0, time.UTC)},
		{"2005-04-07 22:13:13+07:00", time.Date(2005, 4, 7, 15, 13, 13, 0, time.UTC)},
		{"yesterday", time.Date(1996, 6, 27, 0, 0, 0, 0, time.UTC)},
		{"5 days ago", time.Date(1996, 6, 23, 0, 0, 0, 0, time.UTC)},
		{"20 minutes ago", time.Date(1996, 6, 27, 23, 40, 0, 0, time.UTC)},
		{"2 weeks ago", time.Date(1996, 6, 14, 0, 0, 0, 0, time.UTC)},
		{"3:00", time.Date(1996, 6, 28, 3, 0, 0, 0, time.UTC)},
		{"3pm", time.Date(1996, 6, 28, 15, 0, 0, 0, time.UTC)},
		{"1632782809 +0100", time.Date(2021, 9, 27, 21, 46, 49, 0, time.UTC)},
		{"1632782809 -0100", time.Date(2021, 9, 27, 23, 46, 49, 0, time.UTC)},
		{"1632782809", time.Date(2021, 9, 27, 22, 46, 49, 0, time.UTC)},
		{"november 1 2019", time.Date(2019, 11, 1, 0, 0, 0, 0, time.UTC)},
		{"1 november 2019", time.Date(2019, 11, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			output, err := ParseGitDate(tc.input, now)
			require.NoError(t, err)
			// Convert to UTC because generating matching timezones is stupid difficult
			output = output.In(time.UTC)
			require.Equal(t, tc.output, output)
		})
	}

	t.Run("errors", func(t *testing.T) {
		cases := []string{
			"not a date",
			"",
		}

		for _, tc := range cases {
			_, err := ParseGitDate(tc, now)
			require.Error(t, err, "expected error for value %q", tc)
		}
	})
}
