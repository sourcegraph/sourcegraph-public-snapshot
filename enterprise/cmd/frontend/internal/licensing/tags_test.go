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
		{hasLicense: true, licenseTags: nil, want: "Sourcegraph Enterprise"},
		{hasLicense: true, licenseTags: []string{}, want: "Sourcegraph Enterprise"},
		{hasLicense: true, licenseTags: []string{"x"}, want: "Sourcegraph Enterprise"}, // unrecognized tag "x" is ignored
		{hasLicense: true, licenseTags: []string{"starter"}, want: "Sourcegraph Enterprise Starter"},
		{hasLicense: true, licenseTags: []string{"trial"}, want: "Sourcegraph Enterprise (trial)"},
		{hasLicense: true, licenseTags: []string{"dev"}, want: "Sourcegraph Enterprise (dev use only)"},
		{hasLicense: true, licenseTags: []string{"starter", "trial"}, want: "Sourcegraph Enterprise Starter (trial)"},
		{hasLicense: true, licenseTags: []string{"starter", "dev"}, want: "Sourcegraph Enterprise Starter (dev use only)"},
		{hasLicense: true, licenseTags: []string{"starter", "trial", "dev"}, want: "Sourcegraph Enterprise Starter (trial, dev use only)"},
		{hasLicense: true, licenseTags: []string{"trial", "dev"}, want: "Sourcegraph Enterprise (trial, dev use only)"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("hasLicense=%v licenseTags=%v", test.hasLicense, test.licenseTags), func(t *testing.T) {
			if got := ProductNameWithBrand(test.hasLicense, test.licenseTags); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
