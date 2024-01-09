package version

import (
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	t.Run("dev", func(t *testing.T) {
		Mock(devVersion)
		if got, want := Version(), devVersion; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("non-dev", func(t *testing.T) {
		Mock("1.2.3")
		if got, want := Version(), "1.2.3"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestIsDev(t *testing.T) {
	tests := map[string]bool{
		devVersion: true,
		"1.2.3":    false,
	}
	for version, want := range tests {
		if got := IsDev(version); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func Test_monthsFromDays(t *testing.T) {
	tests := []struct {
		name       string
		timeA      string
		timeB      string
		wantMonths int
	}{
		{
			"0 case",
			"01-01-2020",
			"01-26-2020",
			0,
		},
		{
			"base",
			"01-01-2020",
			"02-01-2020",
			1,
		},
		{
			"2 months",
			"01-01-2020",
			"03-01-2020",
			2,
		},
		{
			"3 months",
			"01-01-2020",
			"04-01-2020",
			3,
		},
		{
			"4 months",
			"01-01-2020",
			"05-01-2020",
			4,
		},

		{
			"6+ months",
			"12-01-2019",
			"07-01-2020",
			7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := time.Parse("01-02-2006", tt.timeA)
			if err != nil {
				t.Fatal(err)
			}

			b, err := time.Parse("01-02-2006", tt.timeB)
			if err != nil {
				t.Fatal(err)
			}
			timeSince := b.Sub(a)
			days := timeSince.Hours() / 24

			if got := monthsFromDays(days); got != tt.wantMonths {
				t.Errorf("monthsFromDays() = %v, want %v", got, tt.wantMonths)
			}
		})
	}
}

func TestHowLongOutOfDate(t *testing.T) {
	tests := []struct {
		name           string
		now            time.Time
		buildTimestamp string
		want           int
		wantErr        bool
	}{
		{
			"build is in the future",
			time.Unix(1577577600, 0), // 2019-12-29
			"1577836800",             // 2020-01-01
			0,
			true,
		},
		{
			"6+ months",
			time.Unix(1593561600, 0), // 2020-07-01
			"1577836800",             // 2020-01-01
			6,
			false,
		},
		{
			"3 months",
			time.Unix(1585699200, 0), // 2020-04-01
			"1577836800",             // 2020-01-01
			3,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTimestamp(tt.buildTimestamp)
			got, err := HowLongOutOfDate(tt.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("HowLongOutOfDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HowLongOutOfDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
