package window

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/schema"
)

// We have a bunch of tests in here that rely on unexported fields in the window
// structs. Since we control all of this, we're going to provide a common set of
// options that will allow that.
var (
	cmpAllowUnexported = cmp.AllowUnexported(Window{}, rate{}, windowTime{})
	cmpOptions         = cmp.Options{
		cmpAllowUnexported,
		cmp.Comparer(func(a, b *Configuration) bool {
			return cmp.Equal(a.windows, b.windows, cmpAllowUnexported)
		}),
	}
)

func TestUpdate(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   *[]*schema.BatchChangeRolloutWindow
			want int
		}{
			"one bad window": {
				in: &[]*schema.BatchChangeRolloutWindow{
					{Rate: "xx"},
					{Rate: 0},
				},
				want: 1,
			},
			"two bad windows, ha ha ha": {
				in: &[]*schema.BatchChangeRolloutWindow{
					{Rate: "xx"},
					{Rate: "yy"},
				},
				want: 2,
			},
		} {
			t.Run(name, func(t *testing.T) {
				if err := (&Configuration{}).update(tc.in); err == nil {
					t.Error("unexpected nil error")
				} else if have := len(err.(*multierror.Error).Errors); have != tc.want {
					t.Errorf("unexpected number of errors: have=%d want=%d", have, tc.want)
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   *[]*schema.BatchChangeRolloutWindow
			want *Configuration
		}{
			"nil": {
				in:   nil,
				want: &Configuration{windows: []Window{}},
			},
			"valid windows": {
				in: &[]*schema.BatchChangeRolloutWindow{
					{
						Rate:  "20/hr",
						Days:  []string{"monday"},
						Start: "01:15",
						End:   "02:30",
					},
					{
						Rate: "2/hr",
					},
				},
				want: &Configuration{
					windows: []Window{
						{
							rate:  rate{n: 20, unit: ratePerHour},
							days:  newWeekdaySet(time.Monday),
							start: &windowTime{hour: 1, minute: 15},
							end:   &windowTime{hour: 2, minute: 30},
						},
						{
							rate: rate{n: 2, unit: ratePerHour},
							days: newWeekdaySet(),
						},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				cfg := &Configuration{}
				if err := cfg.update(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(cfg, tc.want, cmpOptions); diff != "" {
					t.Errorf("unexpected configuration (-have +want):\n%s", diff)
				}
			})
		}
	})
}
