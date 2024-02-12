package licensing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductNameWithBrand(t *testing.T) {
	tests := []struct {
		licenseTags []string
		want        string
	}{
		{licenseTags: GetFreeLicenseInfo().Tags, want: "Sourcegraph Free"},
		{licenseTags: nil, want: "Sourcegraph Enterprise"},
		{licenseTags: []string{}, want: "Sourcegraph Enterprise"},
		{licenseTags: []string{"x"}, want: "Sourcegraph Enterprise"}, // unrecognized tag "x" is ignored
		{licenseTags: []string{"trial"}, want: "Sourcegraph Enterprise (trial)"},
		{licenseTags: []string{"dev"}, want: "Sourcegraph Enterprise (dev use only)"},
		{licenseTags: []string{"trial", "dev"}, want: "Sourcegraph Enterprise (trial, dev use only)"},
		{licenseTags: []string{"internal"}, want: "Sourcegraph Enterprise (internal use only)"},

		{licenseTags: []string{"plan:team-0"}, want: "Sourcegraph Team"},
		{licenseTags: []string{"plan:team-0", "trial"}, want: "Sourcegraph Team (trial)"},
		{licenseTags: []string{"plan:team-0", "dev"}, want: "Sourcegraph Team (dev use only)"},
		{licenseTags: []string{"plan:team-0", "dev", "trial"}, want: "Sourcegraph Team (trial, dev use only)"},
		{licenseTags: []string{"plan:team-0", "internal"}, want: "Sourcegraph Team (internal use only)"},

		{licenseTags: []string{"plan:enterprise-0"}, want: "Sourcegraph Enterprise"},
		{licenseTags: []string{"plan:enterprise-0", "trial"}, want: "Sourcegraph Enterprise (trial)"},
		{licenseTags: []string{"plan:enterprise-0", "dev"}, want: "Sourcegraph Enterprise (dev use only)"},
		{licenseTags: []string{"plan:enterprise-0", "dev", "trial"}, want: "Sourcegraph Enterprise (trial, dev use only)"},
		{licenseTags: []string{"plan:enterprise-0", "internal"}, want: "Sourcegraph Enterprise (internal use only)"},

		{licenseTags: []string{"plan:enterprise-1"}, want: "Code Search Enterprise"},
		{licenseTags: []string{"plan:enterprise-1", "trial"}, want: "Code Search Enterprise (trial)"},
		{licenseTags: []string{"plan:enterprise-1", "dev"}, want: "Code Search Enterprise (dev use only)"},
		{licenseTags: []string{"plan:enterprise-1", "dev", "trial"}, want: "Code Search Enterprise (trial, dev use only)"},
		{licenseTags: []string{"plan:enterprise-1", "internal"}, want: "Code Search Enterprise (internal use only)"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("licenseTags=%v", test.licenseTags), func(t *testing.T) {
			assert.Equal(t, test.want, ProductNameWithBrand(test.licenseTags))
		})
	}
}
