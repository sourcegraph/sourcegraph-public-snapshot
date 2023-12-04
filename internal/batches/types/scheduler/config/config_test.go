package config

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestConfiguration(t *testing.T) {
	// Set up some window configurations.
	slow := []*schema.BatchChangeRolloutWindow{
		{Rate: "10/hr"},
	}
	fast := []*schema.BatchChangeRolloutWindow{
		{Rate: "20/hr"},
	}

	for name, tc := range map[string]struct {
		old, new *[]*schema.BatchChangeRolloutWindow
		want     bool
	}{
		"same configuration": {
			old:  &slow,
			new:  &slow,
			want: true,
		},
		"different configuration": {
			old:  &slow,
			new:  &fast,
			want: false,
		},
		"one nil": {
			old:  nil,
			new:  &fast,
			want: false,
		},
		"both nil": {
			old:  nil,
			new:  nil,
			want: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := sameConfiguration(tc.old, tc.new); have != tc.want {
				t.Errorf("unexpected result of comparing configurations: have=%v want=%v", have, tc.want)
			}
		})
	}
}
