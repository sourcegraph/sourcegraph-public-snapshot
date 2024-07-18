package exhaustive

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestIsEnabled(t *testing.T) {
	enabled := true
	disabled := false

	cases := []struct {
		name                 string
		experimentalFeatures *schema.ExperimentalFeatures
		want                 bool
	}{
		{
			name:                 "explicitly enabled",
			experimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &enabled},
			want:                 true,
		},
		{
			name:                 "ExperimentalFeatures=nil",
			experimentalFeatures: nil,
			want:                 true,
		},
		{
			name:                 "SearchJobs=nil",
			experimentalFeatures: &schema.ExperimentalFeatures{},
			want:                 true,
		},
		{
			name:                 "explicitly disabled",
			experimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &disabled},
			want:                 false,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			require.Equal(t, c.want, IsEnabled(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: c.experimentalFeatures}}))
		})
	}
}
