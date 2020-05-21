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

func Test_diff(t *testing.T) {
	tests := []struct {
		name      string
		timeA     string
		timeB     string
		wantYear  int
		wantMonth int
		wantDay   int
		wantHour  int
		wantMin   int
		wantSec   int
	}{
		// TODO: Add test cases.
		{
			"base",
			"01-02-2010",
			"01-03-2010",
			0,
			0,
			1,
			0,
			0,
			0,
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
			gotYear, gotMonth, gotDay, gotHour, gotMin, gotSec := diff(a, b)
			if gotYear != tt.wantYear {
				t.Errorf("diff() gotYear = %v, want %v", gotYear, tt.wantYear)
			}
			if gotMonth != tt.wantMonth {
				t.Errorf("diff() gotMonth = %v, want %v", gotMonth, tt.wantMonth)
			}
			if gotDay != tt.wantDay {
				t.Errorf("diff() gotDay = %v, want %v", gotDay, tt.wantDay)
			}
			if gotHour != tt.wantHour {
				t.Errorf("diff() gotHour = %v, want %v", gotHour, tt.wantHour)
			}
			if gotMin != tt.wantMin {
				t.Errorf("diff() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotSec != tt.wantSec {
				t.Errorf("diff() gotSec = %v, want %v", gotSec, tt.wantSec)
			}
		})
	}
}
