package auth

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRequireLicenseOrSuggestSSOAlerts(t *testing.T) {
	type testCase struct {
		name               string
		hasSSOFeature      bool
		builtInEnabled     bool
		githubSSOEnabled   bool
		featureFlagEnabled bool
		isSiteAdmin        bool
		want               []*graphqlbackend.Alert
	}
	tests := []testCase{
		{
			name:               "do not show anything for non-admin",
			hasSSOFeature:      true,
			builtInEnabled:     true,
			githubSSOEnabled:   true,
			featureFlagEnabled: true,
			isSiteAdmin:        false,

			want: nil,
		},
		{
			name:               "show alert if SSO providers are configured but license doesn't allow it",
			hasSSOFeature:      false,
			builtInEnabled:     false,
			githubSSOEnabled:   true,
			featureFlagEnabled: true,
			isSiteAdmin:        true,
			want: []*graphqlbackend.Alert{{
				GroupValue:   graphqlbackend.AlertGroupLicense,
				TypeValue:    graphqlbackend.AlertTypeError,
				MessageValue: "A Sourcegraph license is required to enable following authentication providers: GitHub OAuth. [**Get a license.**](/site-admin/license)",
			}},
		},
		{
			name:               "show alert if no SSO providers configured but license allows it",
			hasSSOFeature:      true,
			builtInEnabled:     true,
			githubSSOEnabled:   false,
			featureFlagEnabled: true,
			isSiteAdmin:        true,
			want: []*graphqlbackend.Alert{{
				GroupValue:                graphqlbackend.AlertGroupAuthentication,
				TypeValue:                 graphqlbackend.AlertTypeWarning,
				MessageValue:              "We recommend that enterprise instances use SSO or SAML to authenticate users. [Set up authentication now](/site-admin/configuration)",
				IsDismissibleWithKeyValue: "configure-sso-providers",
			}},
		},
		{
			name:               "do not show alert if no SSO providers configured and license has SSO feature and feature flag is disabled",
			hasSSOFeature:      true,
			builtInEnabled:     true,
			githubSSOEnabled:   true,
			featureFlagEnabled: false,
			isSiteAdmin:        true,
			want:               nil,
		},
	}

	var setup = func(test testCase, t *testing.T) graphqlbackend.AlertFuncArgs {
		// mock the auth providers configuration
		providers := make([]schema.AuthProviders, 0)
		if test.githubSSOEnabled {
			providers = append(providers, schema.AuthProviders{
				Github: &schema.GitHubAuthProvider{
					Url:          "https://github.com",
					ClientSecret: "some-secret",
					ClientID:     "some-id",
					AllowOrgs:    []string{"myorg"},
				},
			})
		}
		if test.builtInEnabled {
			providers = append(providers, schema.AuthProviders{
				Builtin: &schema.BuiltinAuthProvider{
					Type:        "builtin",
					AllowSignup: true,
				},
			})
		}
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: providers,
			},
		})

		// mock the featureSSO availability
		licensing.MockCheckFeature = func(feature licensing.Feature) error {
			if test.hasSSOFeature {
				return nil
			}
			return licensing.NewFeatureNotActivatedError("test")
		}

		// cleanup mocks
		t.Cleanup(func() {
			conf.Mock(nil)
			licensing.MockCheckFeature = nil
		})

		// mock the feature flag availability
		ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
		if test.featureFlagEnabled {
			ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"setup-checklist": true}, nil, nil))
		}

		return graphqlbackend.AlertFuncArgs{
			IsSiteAdmin: test.isSiteAdmin,
			Ctx:         ctx,
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := setup(test, t)

			gotAlerts := requireLicenseOrSuggestSSOAlerts(args)

			if len(gotAlerts) != len(test.want) {
				t.Errorf("expected %+v, got %+v", test.want, gotAlerts)
				return
			}
			for i, got := range gotAlerts {
				want := test.want[i]
				if diff := cmp.Diff(*want, *got); diff != "" {
					t.Fatalf("diff mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
