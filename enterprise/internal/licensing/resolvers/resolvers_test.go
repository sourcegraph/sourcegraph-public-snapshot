package resolvers

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
)

func TestEnterpriseLicenseHasFeature(t *testing.T) {
	ctx := context.Background()
	r := &LicenseResolver{}

	buildMock := func(allow ...licensing.Feature) func(feature licensing.Feature) error {
		return func(feature licensing.Feature) error {
			for _, allowed := range allow {
				if feature == allowed {
					return nil
				}
			}

			return licensing.NewFeatureNotActivatedError("feature not allowed")
		}
	}

	for name, tc := range map[string]struct {
		args    *graphqlbackend.EnterpriseLicenseHasFeatureArgs
		mock    func(feature licensing.Feature) error
		want    bool
		wantErr bool
	}{
		"real feature, enabled": {
			args: &graphqlbackend.EnterpriseLicenseHasFeatureArgs{
				Feature: string(licensing.FeatureCampaigns),
			},
			mock:    buildMock(licensing.FeatureCampaigns),
			want:    true,
			wantErr: false,
		},
		"real feature, disabled": {
			args: &graphqlbackend.EnterpriseLicenseHasFeatureArgs{
				Feature: string(licensing.FeatureMonitoring),
			},
			mock:    buildMock(licensing.FeatureCampaigns),
			want:    false,
			wantErr: false,
		},
		"fake feature, enabled": {
			args: &graphqlbackend.EnterpriseLicenseHasFeatureArgs{
				Feature: "foo",
			},
			mock:    buildMock("foo"),
			want:    true,
			wantErr: false,
		},
		"fake feature, disabled": {
			args: &graphqlbackend.EnterpriseLicenseHasFeatureArgs{
				Feature: "foo",
			},
			mock:    buildMock("bar"),
			want:    false,
			wantErr: false,
		},
		"error from check": {
			args: &graphqlbackend.EnterpriseLicenseHasFeatureArgs{
				Feature: string(licensing.FeatureMonitoring),
			},
			mock: func(feature licensing.Feature) error {
				return errors.New("this is a different error")
			},
			want:    false,
			wantErr: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			oldEnforce := licensing.EnforceTiers
			oldMock := licensing.MockCheckFeature
			licensing.EnforceTiers = true
			licensing.MockCheckFeature = tc.mock
			defer func() {
				licensing.EnforceTiers = oldEnforce
				licensing.MockCheckFeature = oldMock
			}()

			have, err := r.EnterpriseLicenseHasFeature(ctx, tc.args)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("got error when no error was expected: %v", err)
				}
			} else if tc.wantErr {
				t.Error("did not get expected error")
			}

			if have != tc.want {
				t.Errorf("unexpected has feature response: have=%v want=%v", have, tc.want)
			}
		})

		t.Run(name+" without enforcement", func(t *testing.T) {
			oldEnforce := licensing.EnforceTiers
			licensing.EnforceTiers = false
			defer func() { licensing.EnforceTiers = oldEnforce }()

			have, err := r.EnterpriseLicenseHasFeature(ctx, tc.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !have {
				t.Error("unexpected disallowance when tiers aren't enforced")
			}
		})
	}
}
