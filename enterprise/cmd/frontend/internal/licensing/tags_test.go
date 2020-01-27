package licensing

import (
	"fmt"
	"testing"
)

func TestProductNameWithBrand(t *testing.T) {
	tests := []struct {
		hasLicense  bool
		licenseTags []string
		want        string
	}{
		{hasLicense: false, want: "Sourcegraph Core"},
		{hasLicense: true, licenseTags: nil, want: "Sourcegraph Elite"},
		{hasLicense: true, licenseTags: []string{}, want: "Sourcegraph Elite"},
		{hasLicense: true, licenseTags: []string{"x"}, want: "Sourcegraph Elite"}, // unrecognized tag "x" is ignored
		{hasLicense: true, licenseTags: []string{EnterpriseBasicTag}, want: "Sourcegraph Enterprise"},
		{hasLicense: true, licenseTags: []string{EnterpriseStarterTag}, want: "Sourcegraph Enterprise Starter"},
		{hasLicense: true, licenseTags: []string{EnterprisePlusTag}, want: "Sourcegraph Enterprise Plus"},
		{hasLicense: true, licenseTags: []string{"trial"}, want: "Sourcegraph Elite (trial)"},
		{hasLicense: true, licenseTags: []string{"dev"}, want: "Sourcegraph Elite (dev use only)"},
		{hasLicense: true, licenseTags: []string{EnterpriseBasicTag, "trial", "dev"}, want: "Sourcegraph Enterprise (trial, dev use only)"},
		{hasLicense: true, licenseTags: []string{EnterpriseStarterTag, "trial"}, want: "Sourcegraph Enterprise Starter (trial)"},
		{hasLicense: true, licenseTags: []string{EnterprisePlusTag, "dev"}, want: "Sourcegraph Enterprise Plus (dev use only)"},
		{hasLicense: true, licenseTags: []string{"trial", "dev"}, want: "Sourcegraph Elite (trial, dev use only)"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("hasLicense=%v licenseTags=%v", test.hasLicense, test.licenseTags), func(t *testing.T) {
			if got := ProductNameWithBrand(test.hasLicense, test.licenseTags); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
