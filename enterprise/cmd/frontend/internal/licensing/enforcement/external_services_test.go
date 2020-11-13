package enforcement

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/db"
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

			store := mockExternalServicesStore{extSvcCount: test.externalServiceCount}
			err := NewPreCreateExternalServiceHook(&store)(context.Background())
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("got error %v, want %v", gotErr, test.wantErr)
			}
		})
	}
}

// A mockExternalServicesStore implements the ExternalServicesStore interface for test purposes.
type mockExternalServicesStore struct {
	extSvcCount int
	err         error
}

// Count returns the number of external services currently configured.
func (m *mockExternalServicesStore) Count(_ context.Context, _ db.ExternalServicesListOptions) (int, error) {
	return m.extSvcCount, m.err
}
