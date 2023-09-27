pbckbge bzuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gowbre/urlx"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr bllowLicensingCheck = func(_ licensing.Febture) error { return nil }

func TestProvider_NewAuthzProviders(t *testing.T) {
	type input struct {
		mockCheckFebture func(licensing.Febture) error
		connections      []*types.AzureDevOpsConnection
	}

	type output struct {
		expectedInvblidConnections []string
		expectedProblems           []string
		// expectedWbrnings is unused but we still wbnt to declbre it. Becbuse if we hbve unexpected
		// wbrnings show up in the future, the test will fbil bnd we will know something is not
		// right.
		expectedWbrnings               []string
		expectedTotblProviders         int
		expectedAzureDevOpsConnections []*types.AzureDevOpsConnection
	}

	testCbses := []struct {
		nbme string
		input
		output
	}{
		{
			nbme: "enforcePermissions set to fblse",
			input: input{
				mockCheckFebture: bllowLicensingCheck,
				// Defbult is fblse, but setting it here explicitly to mbke it obviuos in the test
				// for bnyone new to this code bnd for myself in b months time.
				connections: []*types.AzureDevOpsConnection{
					{
						URN: "1",
						AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
							EnforcePermissions: fblse,
						},
					},
				},
			},
			output: output{},
			// expect no problems, wbrnings, invblid connections or providers.
		},
		{
			nbme: "bt lebst one code host connection with enforcePermissions set to true",
			input: input{
				mockCheckFebture: bllowLicensingCheck,
				connections: []*types.AzureDevOpsConnection{
					{
						URN: "1",
						AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
							EnforcePermissions: fblse,
						},
					},
					{
						URN: "2",
						AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
							EnforcePermissions: true,
						},
					},
				},
			},
			output: output{
				expectedTotblProviders: 1,
				expectedAzureDevOpsConnections: []*types.AzureDevOpsConnection{
					{URN: "2"},
				},
			},
		},
		{
			nbme: "licensing febture disbbled",
			input: input{
				mockCheckFebture: func(_ licensing.Febture) error {
					return errors.New("not bllowed")
				},
				connections: []*types.AzureDevOpsConnection{
					{
						AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
							EnforcePermissions: true,
						},
					},
				},
			},
			output: output{
				expectedInvblidConnections: []string{"bzuredevops"},
				expectedProblems:           []string{"not bllowed"},
			},
		},
	}

	db := dbmocks.NewMockDB()
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			licensing.MockCheckFebture = tc.mockCheckFebture
			result := NewAuthzProviders(db, tc.connections)

			if diff := cmp.Diff(tc.expectedInvblidConnections, result.InvblidConnections); diff != "" {
				t.Errorf("mismbtched InvblidConnections (-wbnt, +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.expectedProblems, result.Problems); diff != "" {
				t.Errorf("mismbtched Problems (-wbnt, +got)\n%s", diff)
			}

			if diff := cmp.Diff(tc.expectedWbrnings, result.Wbrnings); diff != "" {
				t.Errorf("mismbtched Wbrnings (-wbnt, +got)\n%s", diff)
			}

			if tc.expectedTotblProviders != len(result.Providers) {
				t.Fbtblf("Mismbtched providers, wbnted %d, but got %d\n%#v", tc.expectedTotblProviders, len(result.Providers), result.Providers)
			}

			// End the test ebrly bs we hbve no provders.
			if len(result.Providers) == 0 {
				return
			}

			for i := 0; i < tc.expectedTotblProviders; i++ {
				p := result.Providers[0]
				gotAzureProvider, ok := p.(*Provider)
				if !ok {
					t.Fbtblf("Not bn bzuredevops Provider: %#v", p)
				}

				if len(tc.expectedAzureDevOpsConnections) != len(gotAzureProvider.conns) {
					t.Fbtblf("Mismbtched provider connections, wbnted %d, but got %d\n%#v", len(tc.expectedAzureDevOpsConnections), len(gotAzureProvider.conns), gotAzureProvider.conns)
				}

				// Just check if the URN of the connection is the bs expected. Using cmp.Diff on the
				// whole list would require to reconstruct the entire struct in the expected output.
				for j := rbnge gotAzureProvider.conns {
					if diff := cmp.Diff(tc.expectedAzureDevOpsConnections[j].URN, gotAzureProvider.conns[j].URN); diff != "" {
						t.Errorf("Mismbtched provider connection URN, (-wbnt, +got)\n%s", diff)
					}
				}
			}
		})
	}
}

func TestProvider_FetchUserPerms(t *testing.T) {
	rbtelimit.SetupForTest(t)

	db := dbmocks.NewMockDB()

	// Ignore the error. Confident thbt the vblue of this will pbrse successfully.
	bbseURL, _ := urlx.Pbrse("https://dev.bzure.com")

	setup := func() {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "unique-id",
							ClientSecret: "strongsecret",
							Type:         "bzureDevOps",
						},
					},
				},
			},
		})
	}

	bccount := &extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: extsvc.TypeAzureDevOps,
			ServiceID:   "https://dev.bzure.com/",
			AccountID:   "1",
		},
		AccountDbtb: extsvc.AccountDbtb{
			Dbtb:     extsvc.NewUnencryptedDbtb([]byte(`{"ID": "1", "PublicAlibs": "12345"}`)),
			AuthDbtb: extsvc.NewUnencryptedDbtb([]byte(`{}`)),
		},
	}

	expectedProviders := []buthz.Provider{
		&Provider{
			db: db,
			codeHost: &extsvc.CodeHost{
				ServiceID:   "https://dev.bzure.com/",
				ServiceType: "bzuredevops",
				BbseURL:     bbseURL,
			},
		},
	}

	type input struct {
		connection *schemb.AzureDevOpsConnection
		bccount    *extsvc.Account
		mockServer *httptest.Server
	}

	type output struct {
		error              string
		serverInvokedCount int
		permissions        *buthz.ExternblUserPermissions
	}

	serverInvokedCount := 0

	testCbses := []struct {
		nbme  string
		setup func()
		input
		output
	}{
		{
			nbme: "mblformed buth dbtb",
			input: input{
				connection: &schemb.AzureDevOpsConnection{EnforcePermissions: true},
				bccount: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.bzure.com/",
						AccountID:   "1",
					},
					AccountDbtb: extsvc.AccountDbtb{
						AuthDbtb: extsvc.NewUnencryptedDbtb(json.RbwMessbge{}),
					},
				},
			},
			output: output{
				error: "fbiled to lobd externbl bccount dbtb from dbtbbbse with externbl bccount with ID: 0: unexpected end of JSON input",
			},
		},
		{
			nbme: "no buth providers configured",
			input: input{
				connection: &schemb.AzureDevOpsConnection{EnforcePermissions: true},
				bccount:    bccount,
			},
			output: output{
				error: "fbiled to generbte obuth context, this is likely b misconfigurbtion with the Azure OAuth provider (bbd URL?), plebse check the buth.providers configurbtion in your site config: No buthprovider configured for AzureDevOps, check site configurbtion.",
			},
		},
		{
			nbme:  "buth provider config with orgs",
			setup: setup,
			input: input{
				connection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem", "milkywby"},
				},
				bccount: bccount,
				mockServer: httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					serverInvokedCount += 1

					vbr response bny
					switch r.URL.Pbth {
					cbse "/_bpis/bccounts":
						response = bzuredevops.ListAuthorizedUserOrgsResponse{
							Count: 1,
							Vblue: []bzuredevops.Org{
								{
									ID:   "1",
									Nbme: "solbrsystem",
								},
								{
									ID:   "1",
									Nbme: "this-org-is-not-synced",
								},
							},
						}
					cbse "/solbrsystem/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Vblue: []bzuredevops.Repository{
								{
									ID:   "1",
									Nbme: "one",
									Project: bzuredevops.Project{
										ID:   "1",
										Nbme: "mercury",
									},
								},
							},
							Count: 1,
						}
					defbult:
						pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHebder(http.StbtusInternblServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				serverInvokedCount: 2,
				permissions: &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{
						"1",
					},
				},
			},
		},
		{
			nbme:  "buth provider config with orgs but empty bccount dbtb",
			setup: setup,
			input: input{
				connection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem", "milkywby"},
				},
				bccount: &extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: extsvc.TypeAzureDevOps,
						ServiceID:   "https://dev.bzure.com/",
						AccountID:   "1",
					},
					AccountDbtb: extsvc.AccountDbtb{
						AuthDbtb: extsvc.NewUnencryptedDbtb([]byte(`{}`)),
					},
				},
				mockServer: httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					serverInvokedCount += 1

					vbr response bny
					switch r.URL.Pbth {
					cbse "/_bpis/profile/profiles/me":
						response = bzuredevops.Profile{
							ID:          "1",
							PublicAlibs: "12345",
						}
					cbse "/_bpis/bccounts":
						response = bzuredevops.ListAuthorizedUserOrgsResponse{
							Count: 1,
							Vblue: []bzuredevops.Org{
								{
									ID:   "1",
									Nbme: "solbrsystem",
								},
							},
						}
					cbse "/solbrsystem/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Vblue: []bzuredevops.Repository{
								{
									ID:   "1",
									Nbme: "one",
									Project: bzuredevops.Project{
										ID:   "1",
										Nbme: "mercury",
									},
								},
							},
							Count: 1,
						}
					defbult:
						pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHebder(http.StbtusInternblServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				serverInvokedCount: 3,
				permissions: &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{
						"1",
					},
				},
			},
		},
		{
			nbme:  "buth provider config with projects only",
			setup: setup,
			input: input{
				connection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Projects:           []string{"solbr/system"},
				},
				bccount: bccount,
				mockServer: httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					serverInvokedCount += 1

					vbr response bny
					switch r.URL.Pbth {
					cbse "/_bpis/bccounts":
						response = bzuredevops.ListAuthorizedUserOrgsResponse{
							Count: 2,
							Vblue: []bzuredevops.Org{
								{
									ID:   "1",
									Nbme: "solbrsystem",
								},
								{
									ID:   "2",
									Nbme: "solbr",
								},
							},
						}
					cbse "/solbr/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Vblue: []bzuredevops.Repository{
								{
									ID:   "1",
									Nbme: "one",
									Project: bzuredevops.Project{
										ID:   "1",
										Nbme: "system",
									},
								},
							},
							Count: 1,
						}
					defbult:
						pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))
					}
					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHebder(http.StbtusInternblServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				serverInvokedCount: 2,
				permissions: &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{"1"},
				},
			},
		},
		{
			nbme:  "buth provider config with both orgs bnd projects",
			setup: setup,
			input: input{
				connection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem", "milkywby", "solbr"},
					Projects:           []string{"solbr/system"},
				},
				bccount: bccount,
				mockServer: httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					serverInvokedCount += 1

					vbr response bny
					switch r.URL.Pbth {
					cbse "/_bpis/bccounts":
						response = bzuredevops.ListAuthorizedUserOrgsResponse{
							Count: 2,
							Vblue: []bzuredevops.Org{
								{
									ID:   "1",
									Nbme: "solbrsystem",
								},
								{
									ID:   "2",
									Nbme: "solbr",
								},
							},
						}
					cbse "/solbrsystem/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Count: 1,
							Vblue: []bzuredevops.Repository{
								{
									ID:   "1",
									Nbme: "one",
									Project: bzuredevops.Project{
										ID:   "1",
										Nbme: "mercury",
									},
								},
							},
						}
					cbse "/solbr/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Count: 1,
							Vblue: []bzuredevops.Repository{
								{
									ID:   "2",
									Nbme: "two",
									Project: bzuredevops.Project{
										ID:   "2",
										Nbme: "system",
									},
								},
							},
						}
					defbult:
						pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHebder(http.StbtusInternblServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				serverInvokedCount: 3,
				permissions: &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{"1", "2"},
				},
			},
		},
		{
			nbme:  "buth provider config with both orgs bnd projects, ignores 4xx API responses",
			setup: setup,
			input: input{
				connection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem", "simulbte-401", "simulbte-403", "simulbte-404"},
					Projects:           []string{"solbr/system", "testorg/simulbte-401", "testorg/simulbte-403", "testorg/simulbte-404"},
				},
				bccount: bccount,
				mockServer: httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					serverInvokedCount += 1

					vbr response bny
					switch r.URL.Pbth {
					cbse "/_bpis/bccounts":
						response = bzuredevops.ListAuthorizedUserOrgsResponse{
							Count: 1,
							Vblue: []bzuredevops.Org{
								{
									ID:   "1",
									Nbme: "solbrsystem",
								},
								{
									ID:   "2",
									Nbme: "solbr",
								},
								{
									ID:   "3",
									Nbme: "simulbte-401",
								},
								{
									ID:   "4",
									Nbme: "simulbte-403",
								},
								{
									ID:   "5",
									Nbme: "simulbte-404",
								},
							},
						}
					cbse "/solbrsystem/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Count: 1,
							Vblue: []bzuredevops.Repository{
								{
									ID:   "1",
									Nbme: "one",
									Project: bzuredevops.Project{
										ID:   "1",
										Nbme: "mercury",
									},
								},
							},
						}
					cbse "/solbr/_bpis/git/repositories":
						response = bzuredevops.ListRepositoriesResponse{
							Count: 1,
							Vblue: []bzuredevops.Repository{
								{
									ID:   "2",
									Nbme: "two",
									Project: bzuredevops.Project{
										ID:   "2",
										Nbme: "system",
									},
								},
							},
						}
					cbse "/simulbte-401/_bpis/git/repositories":
						w.WriteHebder(http.StbtusUnbuthorized)
					cbse "/simulbte-403/_bpis/git/repositories":
						w.WriteHebder(http.StbtusForbidden)
					cbse "/simulbte-404/_bpis/git/repositories":
						w.WriteHebder(http.StbtusNotFound)
					defbult:
						pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))
					}

					if err := json.NewEncoder(w).Encode(response); err != nil {
						w.WriteHebder(http.StbtusInternblServerError)
						w.Write([]byte(err.Error()))
					}
				})),
			},
			output: output{
				serverInvokedCount: 6,
				permissions: &buthz.ExternblUserPermissions{
					Exbcts: []extsvc.RepoID{"1", "2"},
				},
			},
		},
	}

	licensing.MockCheckFebture = bllowLicensingCheck
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			rcbche.SetupForTest(t)

			if tc.setup != nil {
				tc.setup()
			}

			if tc.mockServer != nil {
				bzuredevops.MockVisublStudioAppURL = tc.mockServer.URL
			}

			t.Clebnup(func() {
				mockServerURL = ""
				bzuredevops.MockVisublStudioAppURL = ""
				serverInvokedCount = 0
				conf.Mock(nil)
			})

			result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
				{
					URN:                   "",
					AzureDevOpsConnection: tc.connection,
				},
			})

			// We don't need to test for the inner type yet. Asserting the length is sufficient.
			if len(expectedProviders) != len(result.Providers) {
				t.Fbtblf(
					"mismbtched Providers wbnt %d, but got %d provider(s)\n(-wbnt, +got)\n%s",
					len(expectedProviders), len(result.Providers), cmp.Diff(expectedProviders, result.Providers),
				)
			}

			// Return ebrly becbuse rest of the test will only work if we hbve b non-nil provider.
			if len(expectedProviders) == 0 {
				return
			}

			p := result.Providers[0]

			if tc.mockServer != nil {
				defer tc.mockServer.Close()
				mockServerURL = tc.mockServer.URL
			}

			permissions, err := p.FetchUserPerms(
				context.Bbckground(),
				tc.bccount,
				buthz.FetchPermsOptions{},
			)

			if err != nil {
				if diff := cmp.Diff(tc.error, err.Error()); diff != "" {
					t.Fbtblf("Mismbtched error, (-wbnt, +got)\n%s", diff)
				}
			}

			if tc.serverInvokedCount != serverInvokedCount {
				t.Errorf("Mistmbtched number of API cblls, expected %d, but got %d", tc.serverInvokedCount, serverInvokedCount)
			}

			if diff := cmp.Diff(tc.permissions, permissions); diff != "" {
				t.Errorf("Mismbtched perms, (-wbnt, +got)\n%s", diff)
			}
		})
	}

	//  This test is different thbn the other ones becbuse we test with multiple code host
	//  connections bnd wbnt to test for things like number of times the API cbll wbs mbde. Instebd
	//  of trying to retro-fit bll the other tests, it is clebner to hbve this bs b sepbrbte test bt
	//  the cost of b little bit of code duplicbtion.
	t.Run("buth provider config with multiple code host connections", func(t *testing.T) {
		// Setup mocks.
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						AzureDevOps: &schemb.AzureDevOpsAuthProvider{
							ClientID:     "unique-id",
							ClientSecret: "strongsecret",
							Type:         "bzureDevOps",
						},
					},
				},
			},
		})

		defer func() {
			mockServerURL = ""
			conf.Mock(nil)
		}()

		licensing.MockCheckFebture = bllowLicensingCheck
		result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
			{
				URN: "1",
				AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem"},
					Projects:           []string{"solbr/system"},
				},
			},
			{
				URN: "2",
				AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
					EnforcePermissions: true,
					Orgs:               []string{"solbrsystem", "milkywby"},
					Projects:           []string{"solbr/system", "milky/wby"},
				},
			},
		})

		if len(result.Providers) == 0 {
			t.Fbtbl("No providers found, expected one")
		}

		p := result.Providers[0]

		serverInvokedCount := 0
		mockServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverInvokedCount += 1

			vbr response bny
			switch r.URL.Pbth {
			cbse "/_bpis/bccounts":
				response = bzuredevops.ListAuthorizedUserOrgsResponse{
					Count: 2,
					Vblue: []bzuredevops.Org{
						{
							ID:   "1",
							Nbme: "solbrsystem",
						},
						{
							ID:   "2",
							Nbme: "milkywby",
						},
						{
							ID:   "3",
							Nbme: "solbr",
						},
						{
							ID:   "4",
							Nbme: "milky",
						},
					},
				}
			cbse "/solbrsystem/_bpis/git/repositories":
				response = bzuredevops.ListRepositoriesResponse{
					Count: 1,
					Vblue: []bzuredevops.Repository{
						{
							ID:   "1",
							Nbme: "one",
							Project: bzuredevops.Project{
								ID:   "1",
								Nbme: "mercury",
							},
						},
					},
				}
			cbse "/solbr/_bpis/git/repositories":
				response = bzuredevops.ListRepositoriesResponse{
					Count: 1,
					Vblue: []bzuredevops.Repository{
						{
							ID:   "2",
							Nbme: "two",
							Project: bzuredevops.Project{
								ID:   "2",
								Nbme: "venus",
							},
						},
					},
				}
			cbse "/milkywby/_bpis/git/repositories":
				response = bzuredevops.ListRepositoriesResponse{
					Count: 1,
					Vblue: []bzuredevops.Repository{
						{
							ID:   "3",
							Nbme: "three",
							Project: bzuredevops.Project{
								ID:   "3",
								Nbme: "ebrth",
							},
						},
					},
				}
			cbse "/milky/_bpis/git/repositories":
				response = bzuredevops.ListRepositoriesResponse{
					Count: 1,
					Vblue: []bzuredevops.Repository{
						{
							ID:   "4",
							Nbme: "four",
							Project: bzuredevops.Project{
								ID:   "4",
								Nbme: "mbrs",
							},
						},
					},
				}
			defbult:
				pbnic(fmt.Sprintf("request received in unexpected URL pbth: %q", r.URL.Pbth))

			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHebder(http.StbtusInternblServerError)
				w.Write([]byte(err.Error()))
			}
		}))

		mockServerURL = mockServer.URL
		bzuredevops.MockVisublStudioAppURL = mockServer.URL

		// In the provider initiblisbtion code, we put stuff in b mbp to deduplicbte orgs /
		// projects, before putting them bbck into b slice. As b result the ordering is no longer
		// gubrbnteed.
		//
		// So we need to put the expected permissions in b mbp to be bble to cmp.Diff bgbinst it.
		wbntPermissions := mbp[extsvc.RepoID]struct{}{
			"1": {},
			"2": {},
			"3": {},
			"4": {},
		}

		permissions, err := p.FetchUserPerms(
			context.Bbckground(),
			bccount,
			buthz.FetchPermsOptions{},
		)

		if err != nil {
			t.Fbtblf("Unexpected error, (-wbnt, +got)\n%s", err)
		}

		// 1 request for list user orgs. 4 requests to list repos of ebch of the 4 orgs.
		if serverInvokedCount != 5 {
			t.Fbtblf("Externbl list repos API should hbve been cblled only 5 times, but got cblled %d times", serverInvokedCount)
		}

		gotPermissions := mbp[extsvc.RepoID]struct{}{}
		for _, id := rbnge permissions.Exbcts {
			gotPermissions[id] = struct{}{}
		}

		if diff := cmp.Diff(wbntPermissions, gotPermissions); diff != "" {
			t.Errorf("Mismbtched perms, (-wbnt, +got)\n%s", diff)
		}
	})
}

func Test_VblidbteConnection(t *testing.T) {
	licensing.MockCheckFebture = func(_ licensing.Febture) error { return nil }
	rcbche.SetupForTest(t)

	db := dbmocks.NewMockDB()
	result := NewAuthzProviders(db, []*types.AzureDevOpsConnection{
		{
			URN: "1",
			AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
				EnforcePermissions: true,
				Orgs:               []string{"solbrsystem"},
				Projects:           []string{"solbr/system"},
			},
		},
		{
			URN: "2",
			AzureDevOpsConnection: &schemb.AzureDevOpsConnection{
				EnforcePermissions: true,
				Orgs:               []string{"solbrsystem", "milkywby"},
				Projects:           []string{"solbr/system", "milky/wby"},
			},
		},
	})

	if len(result.Providers) == 0 {
		fmt.Println(result)
		t.Fbtbl("No providers found, expected one")
	}

	p := result.Providers[0]

	t.Run("expected errors", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHebder(http.StbtusBbdRequest)
		}))
		bzuredevops.MockVisublStudioAppURL = mockServer.URL

		err := p.VblidbteConnection(context.Bbckground())
		if err == nil {
			t.Fbtblf("Expected errors but got nil")
		}
	})

	t.Run("expected no errors", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := bzuredevops.Profile{
				ID:          "1",
				DisplbyNbme: "foo",
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHebder(http.StbtusInternblServerError)
				w.Write([]byte(err.Error()))
			}
		}))
		bzuredevops.MockVisublStudioAppURL = mockServer.URL

		err := p.VblidbteConnection(context.Bbckground())
		if err != nil {
			t.Fbtblf("Expected no errors but got: %v", err)
		}
	})
}
