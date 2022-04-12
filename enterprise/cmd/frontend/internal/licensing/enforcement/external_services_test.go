package enforcement

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestNewPreCreateExternalServiceHook(t *testing.T) {
	if !licensing.EnforceTiers {
		licensing.EnforceTiers = true
		defer func() { licensing.EnforceTiers = false }()
	}

	tests := []struct {
		desc                 string
		license              *license.Info
		externalServiceCount int
		wantErr              bool
	}{
		{
			desc:                 "An older starter plan with unlimited external services",
			license:              &license.Info{Tags: []string{"plan:old-starter-0"}},
			externalServiceCount: 1000,
			wantErr:              false,
		},
		{
			desc:                 "An older enterprise plan with unlimited external services",
			license:              &license.Info{Tags: []string{"plan:old-enterprise-0"}},
			externalServiceCount: 1000,
			wantErr:              false,
		},
		{
			desc:                 "An enterprise plan with unlimited external services",
			license:              &license.Info{Tags: []string{"plan:enterprise-0"}},
			externalServiceCount: 1000,
			wantErr:              false,
		},
		{
			desc:                 "A team plan with limited external services",
			license:              &license.Info{Tags: []string{"plan:team-0"}},
			externalServiceCount: 1,
			wantErr:              true,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("license %s with %d external services", test.license, test.externalServiceCount), func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			externalServices := database.NewMockExternalServiceStore()
			externalServices.CountFunc.SetDefaultReturn(test.externalServiceCount, nil)
			err := NewBeforeCreateExternalServiceHook()(context.Background(), externalServices)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("got error %v, want %v", gotErr, test.wantErr)
			}
		})
	}
}
