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

	setup := func() {
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
	}

	passLicensingCheck := func(_ licensing.Feature) error { return nil }
	account := &extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeAzureDevOps,
			ServiceID:   "https://dev.azure.com/",
			AccountID:   "1",
		},
		AccountData: extsvc.AccountData{
			AuthData: extsvc.NewUnencryptedData([]byte(`
{}`)),
		},
	}

	expectedProviders := []authz.Provider{
		&Provider{
			db: db,
			codeHost: &extsvc.CodeHost{
				ServiceID:   "https://dev.azure.com/",
				ServiceType: "azuredevops",
				BaseURL:     baseURL,
			},
		},
	}

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
		name  string
		setup func()
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
				mockCheckFeature: passLicensingCheck,
				connection:       &schema.AzureDevOpsConnection{},
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
				providers: expectedProviders,
				error:     "failed to load external account data from database with external account with ID: 0: unexpected end of JSON input",
			},
		},
		{
			name: "no auth providers configured",
			input: input{
				mockCheckFeature: passLicensingCheck,
				connection:       &schema.AzureDevOpsConnection{},
				account:          account,
			},
			output: output{
				providers: expectedProviders,
				error:     "failed to generate oauth context, this is likely a misconfiguration with the Azure OAuth provider (bad URL?), please check the auth.providers configuration in your site config: No authprovider configured for AzureDevOps, check site configuration.",
			},
		},
		{
			name:  "auth provider config with orgs",
			setup: setup,
			input: input{
				mockCheckFeature: passLicensingCheck,
				connection: &schema.AzureDevOpsConnection{
					Orgs: []string{"solarsystem"},
				},
				account: account,
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

					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				providers: expectedProviders,
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{
						"1",
					},
				},
			},
		},
		{
			name:  "auth provider config with projects",
			setup: setup,
			input: input{
				mockCheckFeature: passLicensingCheck,
				connection: &schema.AzureDevOpsConnection{
					Projects: []string{"solar/system"},
				},
				account: account,
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
				providers: expectedProviders,
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1"},
				},
			},
		},
		{
			name:  "auth provider config with both orgs and projects",
			setup: setup,
			input: input{
				mockCheckFeature: passLicensingCheck,
				connection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem"},
					Projects: []string{"solar/system"},
				},
				account: account,
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
				providers: expectedProviders,
				permissions: &authz.ExternalUserPermissions{
					Exacts: []extsvc.RepoID{"1", "2"},
				},
			},
		},
		{
			name:  "auth provider config with both orgs and projects, ignores 4xx API responses",
			setup: setup,
			input: input{
				mockCheckFeature: passLicensingCheck,
				connection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem", "simulate-401", "simulate-403", "simulate-404"},
					Projects: []string{"solar/system", "testorg/simulate-401", "testorg/simulate-403", "testorg/simulate-404"},
				},
				account: account,
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
				providers: expectedProviders,
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

			t.Cleanup(func() {
				mockServerURL = ""
				conf.Mock(nil)
			})

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

			// We don't need to test for the inner type yet. Asserting the length is sufficient.
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
				mockServerURL = tc.mockServer.URL
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

	//  This test is different than the other ones because we test with multiple code host
	//  connections and want to test for things like number of times the API call was made. Instead
	//  of trying to retro-fit all the other tests, it is cleaner to have this as a separate test at
	//  the cost of a little bit of code duplication.
	t.Run("auth provider config with multiple code host connections", func(t *testing.T) {
		// Setup mocks.
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

		defer func() {
			mockServerURL = ""
			conf.Mock(nil)
		}()

		licensing.MockCheckFeature = passLicensingCheck

		result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
			{
				URN: "1",
				AzureDevOpsConnection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem"},
					Projects: []string{"solar/system"},
				},
			},
			{
				URN: "2",
				AzureDevOpsConnection: &schema.AzureDevOpsConnection{
					Orgs:     []string{"solarsystem", "milkyway"},
					Projects: []string{"solar/system", "milky/way"},
				},
			},
		})

		if len(result.Providers) == 0 {
			t.Fatal("No providers found, expected one")
		}

		p := result.Providers[0]

		serverInvokedCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverInvokedCount += 1
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
			} else if strings.HasPrefix(r.URL.Path, "/solar/system") {
				response.Value = append(response.Value, azuredevops.Repository{
					ID:   "2",
					Name: "two",
					Project: azuredevops.Project{
						ID:   "2",
						Name: "venus",
					},
				})
			} else if strings.HasPrefix(r.URL.Path, "/milkyway") {
				response.Value = append(response.Value, azuredevops.Repository{
					ID:   "3",
					Name: "three",
					Project: azuredevops.Project{
						ID:   "3",
						Name: "earth",
					},
				})
			} else if strings.HasPrefix(r.URL.Path, "/milky/way") {
				response.Value = append(response.Value, azuredevops.Repository{
					ID:   "4",
					Name: "four",
					Project: azuredevops.Project{
						ID:   "4",
						Name: "mars",
					},
				})
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}))

		mockServerURL = mockServer.URL

		// In the provider initialisation code, we put stuff in a map to deduplicate orgs /
		// projects, before putting them back into a slice. As a result the ordering is no longer
		// guaranteed.
		//
		// So we need to put the expected permissions in a map to be able to cmp.Diff against it.
		wantPermissions := map[extsvc.RepoID]struct{}{
			"1": {},
			"2": {},
			"3": {},
			"4": {},
		}

		permissions, err := p.FetchUserPerms(
			context.Background(),
			account,
			authz.FetchPermsOptions{},
		)

		if err != nil {
			t.Fatalf("Unexpected error, (-want, +got)\n%s", err)
		}

		if serverInvokedCount != 4 {
			t.Fatalf("External list reops API should have been called only 4 times, but got called %d times", serverInvokedCount)
		}

		gotPermissions := map[extsvc.RepoID]struct{}{}
		for _, id := range permissions.Exacts {
			gotPermissions[id] = struct{}{}
		}

		if diff := cmp.Diff(wantPermissions, gotPermissions); diff != "" {
			t.Errorf("Mismatched perms, (-want, +got)\n%s", diff)
		}
	})
}

func Test_ValidateConnection(t *testing.T) {
	licensing.MockCheckFeature = func(_ licensing.Feature) error { return nil }
	rcache.SetupForTest(t)

	db := database.NewMockDB()
	result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
		{
			URN: "1",
			AzureDevOpsConnection: &schema.AzureDevOpsConnection{
				Orgs:     []string{"solarsystem"},
				Projects: []string{"solar/system"},
			},
		},
		{
			URN: "2",
			AzureDevOpsConnection: &schema.AzureDevOpsConnection{
				Orgs:     []string{"solarsystem", "milkyway"},
				Projects: []string{"solar/system", "milky/way"},
			},
		},
	})

	if len(result.Providers) == 0 {
		fmt.Println(result)
		t.Fatal("No providers found, expected one")
	}

	p := result.Providers[0]

	t.Run("expected errors", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		azuredevops.MockVisualStudioAppURL = mockServer.URL

		err := p.ValidateConnection(context.Background())
		if err == nil {
			t.Fatalf("Expected errors but got nil")
		}
	})

	t.Run("expected no errors", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := azuredevops.Profile{
				ID:          "1",
				DisplayName: "foo",
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}))
		azuredevops.MockVisualStudioAppURL = mockServer.URL

		err := p.ValidateConnection(context.Background())
		if err != nil {
			t.Fatalf("Expected no errors but got: %v", err)
		}
	})
}
