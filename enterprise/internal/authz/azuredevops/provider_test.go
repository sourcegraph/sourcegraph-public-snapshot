package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/goware/urlx"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestProvider_FetchUserPerms(t *testing.T) {
	db := database.NewMockDB()

	// Ignore the error. Confident that the value of this will parse successfully.
	baseURL, _ := urlx.Parse("https://dev.azure.com")

	type input struct {
		mockCheckFeature func(licensing.Feature) error
		connection       *schema.AzureDevOpsConnection
		account          *extsvc.Account
		mockServer       *httptest.Server
	}

	type output struct {
		invalidConnections []string
		problems           []string
		warnings           []string
		providers          []authz.Provider
		error              string
		permissions        *authz.ExternalUserPermissions
	}

	testCases := []struct {
		name     string
		setup    func()
		teardown func()
		input
		output
	}{
		{
			name: "licensing feature disabled",
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return errors.New("not allowed")
				},
				connection: &schema.AzureDevOpsConnection{},
			},
			output: output{
				invalidConnections: []string{"azuredevops"},
				problems:           []string{"not allowed"},
			},
		},
		{
			name: "malformed auth data",
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData(json.RawMessage{}),
					},
				},
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				error: "failed to load external account data from database with external account with ID: 0: unexpected end of JSON input",
			},
		},
		{
			name: "no auth providers configured",
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData([]byte(`{}`)),
					},
				},
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				error: "failed to generate oauth context, this is likely a misconfiguration with the Azure OAuth provider (bad URL?), please check the auth.providers configuration in your site config: No authprovider configured for AzureDevOps, check site configuration.",
			},
		},
		{
			name: "auth provider config with orgs",
			setup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						AuthProviders: []schema.AuthProviders{
							{
								AzureDevOps: &schema.AzureDevOpsAuthProvider{
									ClientID:     "unique-id",
									ClientSecret: "strongsecret",
									Type:         "azureDevOps",
								},
							},
						},
					},
				})
			},
			teardown: func() { conf.Mock(nil) },
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{
					Orgs: []string{"solarsystem"},
				},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData([]byte(`
{}`)),
					},
				},
				mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := azuredevops.ListRepositoriesResponse{
						Value: []azuredevops.Repository{
							{
								ID:   "1",
								Name: "one",
								Project: azuredevops.Project{
									ID:   "1",
									Name: "mercury",
								},
							},
						},
						Count: 1,
					}

					json.NewEncoder(w).Encode(response)
					return
				})),
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{
						"1",
					},
				},
			},
		},
		{
			name: "auth provider config with projects",
			setup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						AuthProviders: []schema.AuthProviders{
							{
								AzureDevOps: &schema.AzureDevOpsAuthProvider{
									ClientID:     "unique-id",
									ClientSecret: "strongsecret",
									Type:         "azureDevOps",
								},
							},
						},
					},
				})
			},
			teardown: func() { conf.Mock(nil) },
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{
					Projects: []string{"solar/system"},
				},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData([]byte(`
{}`)),
					},
				},
				mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := azuredevops.ListRepositoriesResponse{
						Value: []azuredevops.Repository{
							{
								ID:   "1",
								Name: "one",
								Project: azuredevops.Project{
									ID:   "1",
									Name: "mercury",
								},
							},
						},
						Count: 1,
					}

					json.NewEncoder(w).Encode(response)
					return
				})),
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1"},
				},
			},
		},
		{
			name: "auth provider config with both orgs and projects",
			setup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						AuthProviders: []schema.AuthProviders{
							{
								AzureDevOps: &schema.AzureDevOpsAuthProvider{
									ClientID:     "unique-id",
									ClientSecret: "strongsecret",
									Type:         "azureDevOps",
								},
							},
						},
					},
				})
			},
			teardown: func() { conf.Mock(nil) },
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem"},
					Projects: []string{"solar/system"},
				},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData([]byte(`
{}`)),
					},
				},
				mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := azuredevops.ListRepositoriesResponse{
						Value: []azuredevops.Repository{},
						Count: 1,
					}

					if strings.HasPrefix(r.URL.Path, "/solarsystem") {
						response.Value = append(response.Value, azuredevops.Repository{
							ID:   "1",
							Name: "one",
							Project: azuredevops.Project{
								ID:   "1",
								Name: "mercury",
							},
						})
					}

					if strings.HasPrefix(r.URL.Path, "/solar/system") {
						response.Value = append(response.Value, azuredevops.Repository{
							ID:   "2",
							Name: "two",
							Project: azuredevops.Project{
								ID:   "2",
								Name: "venus",
							},
						})
					}

					json.NewEncoder(w).Encode(response)
					return
				})),
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1", "2"},
				},
			},
		},

		{
			name: "auth provider config with both orgs and projects, ignores 4xx API responses",
			setup: func() {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						AuthProviders: []schema.AuthProviders{
							{
								AzureDevOps: &schema.AzureDevOpsAuthProvider{
									ClientID:     "unique-id",
									ClientSecret: "strongsecret",
									Type:         "azureDevOps",
								},
							},
						},
					},
				})
			},
			teardown: func() { conf.Mock(nil) },
			input: input{
				mockCheckFeature: func(_ licensing.Feature) error {
					return nil
				},
				connection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem", "simulate-401", "simulate-403", "simulate-404"},
					Projects: []string{"solar/system", "testorg/simulate-401", "testorg/simulate-403", "testorg/simulate-404"},
				},
				account: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.azure.com/",
						AccountID:   "1",
					},
					AccountData: extsvc.AccountData{
						AuthData: extsvc.NewUnencryptedData([]byte(`
{}`)),
					},
				},
				mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := azuredevops.ListRepositoriesResponse{
						Value: []azuredevops.Repository{},
						Count: 1,
					}

					fmt.Println(r.URL.Path)

					if strings.HasPrefix(r.URL.Path, "/solarsystem") {
						response.Value = append(response.Value, azuredevops.Repository{
							ID:   "1",
							Name: "one",
							Project: azuredevops.Project{
								ID:   "1",
								Name: "mercury",
							},
						})
					} else if strings.HasPrefix(r.URL.Path, "/solar/system") {
						response.Value = append(response.Value, azuredevops.Repository{
							ID:   "2",
							Name: "two",
							Project: azuredevops.Project{
								ID:   "2",
								Name: "venus",
							},
						})
					} else if strings.Contains(r.URL.Path, "401") {
						w.WriteHeader(http.StatusUnauthorized)
						return
					} else if strings.Contains(r.URL.Path, "403") {
						w.WriteHeader(http.StatusForbidden)
					} else if strings.Contains(r.URL.Path, "404") {
						w.WriteHeader(http.StatusNotFound)
					}

					json.NewEncoder(w).Encode(response)
					return
				})),
			},
			output: output{
				providers: []authz.Provider{
					&Provider{
						db: db,
						codeHost: &extsvc.CodeHost{
							ServiceID:   "https://dev.azure.com/",
							ServiceType: "azuredevops",
							BaseURL:     baseURL,
						},
					},
				},
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1", "2"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rcache.SetupForTest(t)

			if tc.setup != nil {
				tc.setup()
			}

			if tc.teardown != nil {
				defer tc.teardown()
			}

			licensing.MockCheckFeature = tc.mockCheckFeature
			result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
				{
					URN:                   "",
					AzureDevOpsConnection: tc.connection,
				},
			})

			if diff := cmp.Diff(tc.invalidConnections, result.InvalidConnections); diff != "" {
				t.Errorf("mismatched InvalidConnections (-want, +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.problems, result.Problems); diff != "" {
				t.Errorf("mismatched Problems (-want, +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.warnings, result.Warnings); diff != "" {
				t.Errorf("mismatched Warnings (-want, +got)\n%s", diff)
			}

			// We dont need to test for the inner type yet. Asserting the length is sufficient.
			if len(tc.providers) != len(result.Providers) {
				t.Fatalf(
					"mismatched Providers want %d, but got %d provider(s)\n(-want, +got)\n%s",
					len(tc.providers), len(result.Providers), cmp.Diff(tc.providers, result.Providers),
				)
			}

			// Return early because rest of the test will only work if we have a non-nil provider.
			if len(tc.providers) == 0 {
				return
			}

			p := result.Providers[0]

			if tc.mockServer != nil {
				defer tc.mockServer.Close()
				MOCK_API_URL = tc.mockServer.URL
			}

			permissions, err := p.FetchUserPerms(
				context.Background(),
				tc.account,
				authz.FetchPermsOptions{},
			)

			if err != nil {
				if diff := cmp.Diff(tc.error, err.Error()); diff != "" {
					t.Fatalf("Mismatched error, (-want, +got)\n%s", diff)
				}
			}

			if diff := cmp.Diff(tc.permissions, permissions); diff != "" {
				t.Errorf("Mismatched perms, (-want, +got)\n%s", diff)
			}
		})
	}
}
