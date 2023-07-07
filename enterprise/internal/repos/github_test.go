package repos

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
	ghaauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TestGithubSource_ListRepos_GitHubApp tests the ListRepos function for GitHub
// Apps specifically. We have a separate test case for this so that the VCR
// tests for GitHub App and non-GitHub App connections can be updated separately,
// as setting up credentials for a GitHub App VCR test is significantly more effort.
func TestGithubSource_ListRepos_GitHubApp(t *testing.T) {
	// This private key is no longer valid. If this VCR test needs to be updated,
	// a new GitHub App with new keys and secrets will have to be created
	// and deleted afterwards.
	const ghAppPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAqHG1k8V0pCUAh+U5+thGPHutM0R8rIVmAlPCVw7VzqtxyMf3
5pK4uc7IrIy29w5seyJRDLtY7PnsqU+lvXaAL8k3J0CtRi7doZEfUX1lGOqpomsg
fyJeBH988ZSK+b8DUk7GAj0+Vgy6L70Q3ZdRJt2Ili3Zwtlv14vNyuAxUhgP04Ag
1rczMjNc5LJpvw7gFPk7paYgV41LLrTr1c66ZycXbqFk/a/er6QW4Nnojn1jjJNb
mq6xU7XZlx65BglW8iKJORmo2Or88H178/vFSNnxW0eUarw3FDKsVBubTdr0vLRV
hw5EIsQ7nfrUBvTjMmouLEennYEIStYWNKfuAQIDAQABAoIBAHxIYeQlJZnTH2Al
drEpkDEiQ7n3B1I3nvuKl3KqpIC3qN2vBa8fhKK7+v6tWHZTMyFrQYf2V3eKM978
wFpZq90WRtZ0dyS4gZirPgNfVQ+cXQtUpYaIcfw5oJOSuTPqhuXc72ZJj8vn2hxN
ELue4SafAB9mtyx4SHguU+ojnuBlZA8w2SllddWfJXnmSymrQUCOKvyL/NKSLRqf
Vws4T01Sn5vsJp//lQtLhIDRTFk6qSeX007gNMNi/TiHka+HgulX4R5cxptXq4Xf
xgH9Us2v87UbRRfPygptDk1YZ+g+zpqjX6bbZN8TsMceMkV6eN9txFo9YQlzPxUP
zsP5M5ECgYEA1A3uATaRR/eDj/ziGGsJdxP6lWqmfozw2edQEmIaKBUTj2FOSKc5
vZKQlw54sTtW5tN+9wkiiavCpq5wWRddPfxA0S2hwCnp3IrAanfrD6mjK1oSczf/
lX4c5kZoSIuiJfImToJa6NMoGYdG7btT6wBuqc6NOST55AobBwoQK80CgYEAy1oi
8v/pRdgOaCg1Qu78HS/covyUkNzt0NRL0KUQ//cJuhxkpbycjInU3W0n9sfa694b
dK+D3br1GKRJaeKFZQyW7PV2B5ckXuBdtHOHgFdc14BtQJDWELGthE7rx3BdZYpl
Dz0vF/okm3Vv2J3zBwT733fjYWqQzlOjBPBuXwUCgYEAxGCyDQWPvWoGuI2khKB7
f39NDJpb3c6ALgv9J0kamAwMtTeT28yhuGHG7V1FgDxH2jP63KPlDEG4Xcwl1xvA
CetVy2HK7b7jCI6mavLrCPI8XaVoeLNfSf4knUyOvsAxRZrexs4JipwiAqI4mWhl
6rfXxAG43zbTBNAm/3neR/ECgYBns16xRxoh2Q13xlFrAc6l37uHjoEA4vmQDkNf
cl4Z+lQGieY1stquvLdF+B1yNvcIY6ritYLstyO4Xkdl7POT1Xi9/GslcclFbOu8
U1Ide+/HoiGU1Iel2cYf+9M3ULEAUDQ7Mjtq4dB7Sscv01SVFtCPZGcbTans3i/7
G9VdNQKBgQC3p4CuoJZ0dWizgCuClOPH879RcBfE16xrxxQ+CbQTkYtyqTbaf+Et
x0BN4L+7v8OqXKSX0opjSVT7lg+RhAoZ8Efv+CsJn6SKz9RmFfNGkiqmwjmFg9k2
EyAO2RYQG7mSE6w6CtTFiCjjmELpvdD2s1ygvPdCO1MJlCX264E3og==
-----END RSA PRIVATE KEY-----
`
	assertAllReposListed := func(want []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			have := rs.Names()
			sort.Strings(have)
			sort.Strings(want)

			if !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		}
	}

	testCases := []struct {
		name   string
		assert typestest.ReposAssertion
		mw     httpcli.Middleware
		conf   *schema.GitHubConnection
		err    string
	}{
		{
			name: "github app",
			assert: assertAllReposListed([]string{
				"github.com/pjlast/ygoza",
			}),
			conf: &schema.GitHubConnection{
				Url: "https://github.com/",
				GitHubAppDetails: &schema.GitHubAppDetails{
					InstallationID:       38844262,
					AppID:                350528,
					BaseURL:              "https://github.com/",
					CloneAllRepositories: true,
				},
			},
			err: "<nil>",
		},
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ghAppsStore := db.GitHubApps().WithEncryptionKey(keyring.Default().GitHubAppKey)
	_, err := ghAppsStore.Create(context.Background(), &ghtypes.GitHubApp{
		AppID:        350528,
		BaseURL:      "https://github.com/",
		Name:         "SourcegraphForPetriWoop",
		Slug:         "sourcegraphforpetriwoop",
		PrivateKey:   ghAppPrivateKey,
		ClientID:     "Iv1.4e78f8613134c221",
		ClientSecret: "0e1540fbcea7c59ddae70dc6eb0ae4f1f52255c9",
		Domain:       types.ReposGitHubAppDomain,
		Logo:         "logo.png",
		AppURL:       "https://github.com/appurl",
	})
	require.NoError(t, err)
	auth.FromConnection = ghaauth.CreateEnterpriseFromConnection(ghAppsStore, keyring.Default().GitHubAppKey)

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITHUB-LIST-REPOS/" + tc.name
		t.Run(tc.name, func(t *testing.T) {
			// The GitHubSource uses the github.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			var (
				cf   *httpcli.Factory
				save func(testing.TB)
			)
			if tc.mw != nil {
				cf, save = repos.NewClientFactory(t, tc.name, tc.mw)
			} else {
				cf, save = repos.NewClientFactory(t, tc.name)
			}

			defer save(t)

			svc := &types.ExternalService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(repos.MarshalJSON(t, tc.conf)),
			}

			ctx := context.Background()
			githubSrc, err := repos.NewGitHubSource(ctx, logtest.Scoped(t), svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := repos.ListAll(context.Background(), githubSrc)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}
