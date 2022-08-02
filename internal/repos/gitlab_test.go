package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestProjectQueryToURL(t *testing.T) {
	tests := []struct {
		projectQuery string
		perPage      int
		expURL       string
		expErr       error
	}{{
		projectQuery: "?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "projects?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "groups/groupID/projects",
		perPage:      100,
		expURL:       "groups/groupID/projects?per_page=100",
	}, {
		projectQuery: "groups/groupID/projects?foo=bar",
		perPage:      100,
		expURL:       "groups/groupID/projects?foo=bar&per_page=100",
	}, {
		projectQuery: "",
		perPage:      100,
		expURL:       "projects?per_page=100",
	}, {
		projectQuery: "https://somethingelse.com/foo/bar",
		perPage:      100,
		expErr:       schemeOrHostNotEmptyErr,
	}}

	for _, test := range tests {
		t.Logf("Test case %+v", test)
		url, err := projectQueryToURL(test.projectQuery, test.perPage)
		if url != test.expURL {
			t.Errorf("expected %v, got %v", test.expURL, url)
		}
		if !errors.Is(err, test.expErr) {
			t.Errorf("expected err %v, got %v", test.expErr, err)
		}
	}
}

func TestGitLabSource_GetRepo(t *testing.T) {
	testCases := []struct {
		name                 string
		projectWithNamespace string
		assert               func(*testing.T, *types.Repo)
		err                  string
	}{
		{
			name:                 "not found",
			projectWithNamespace: "foobarfoobarfoobar/please-let-this-not-exist",
			err:                  "GitLab project \"foobarfoobarfoobar/please-let-this-not-exist\" not found",
		},
		{
			name:                 "found",
			projectWithNamespace: "gitlab-org/gitaly",
			assert: func(t *testing.T, have *types.Repo) {
				t.Helper()

				want := &types.Repo{
					Name:        "gitlab.com/gitlab-org/gitaly",
					Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
					URI:         "gitlab.com/gitlab-org/gitaly",
					Stars:       168,
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "2009901",
						ServiceType: "gitlab",
						ServiceID:   "https://gitlab.com/",
					},
					Sources: map[string]*types.SourceInfo{
						"extsvc:gitlab:0": {
							ID:       "extsvc:gitlab:0",
							CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
						},
					},
					Metadata: &gitlab.Project{
						ProjectCommon: gitlab.ProjectCommon{
							ID:                2009901,
							PathWithNamespace: "gitlab-org/gitaly",
							Description:       "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
							WebURL:            "https://gitlab.com/gitlab-org/gitaly",
							HTTPURLToRepo:     "https://gitlab.com/gitlab-org/gitaly.git",
							SSHURLToRepo:      "git@gitlab.com:gitlab-org/gitaly.git",
						},
						Visibility: "",
						Archived:   false,
						StarCount:  168,
						ForksCount: 76,
					},
				}

				if !reflect.DeepEqual(have, want) {
					t.Errorf("response: %s", cmp.Diff(have, want))
				}
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITLAB-DOT-COM/" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GitLabSource uses the gitlab.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			svc := &types.ExternalService{
				Kind: extsvc.KindGitLab,
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url: "https://gitlab.com",
				}),
			}

			ctx := context.Background()
			db := database.NewMockDB()
			gitlabSrc, err := NewGitLabSource(ctx, logtest.Scoped(t), db, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repo, err := gitlabSrc.GetRepo(context.Background(), tc.projectWithNamespace)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repo)
			}
		})
	}
}

func TestGitLabSource_makeRepo(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "gitlab-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*gitlab.Project
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindGitLab}

	tests := []struct {
		name   string
		schema *schema.GitLabConnection
	}{
		{
			name: "simple",
			schema: &schema.GitLabConnection{
				Url: "https://gitlab.com",
			},
		}, {
			name: "ssh",
			schema: &schema.GitLabConnection{
				Url:        "https://gitlab.com",
				GitURLType: "ssh",
			},
		}, {
			name: "path-pattern",
			schema: &schema.GitLabConnection{
				Url:                   "https://gitlab.com",
				RepositoryPathPattern: "gl/{pathWithNamespace}",
			},
		},
	}
	for _, test := range tests {
		test.name = "GitLabSource_makeRepo_" + test.name
		t.Run(test.name, func(t *testing.T) {

			ctx := context.Background()
			db := database.NewMockDB()
			s, err := newGitLabSource(ctx, logtest.Scoped(t), db, &svc, test.schema, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*types.Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}

			testutil.AssertGolden(t, "testdata/golden/"+test.name, update(test.name), got)
		})
	}
}

func TestGitLabSource_WithAuthenticator(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("supported", func(t *testing.T) {
		var src Source

		ctx := context.Background()
		db := database.NewMockDB()
		src, err := newGitLabSource(ctx, logger, db, &types.ExternalService{}, &schema.GitLabConnection{}, nil)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		src, err = src.(UserSource).WithAuthenticator(&auth.OAuthBearerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitLabSource); !ok {
			t.Error("cannot coerce Source into GitLabSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"BasicAuth":   &auth.BasicAuth{},
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				var src Source

				ctx := context.Background()
				db := database.NewMockDB()
				src, err := newGitLabSource(ctx, logger, db, &types.ExternalService{}, &schema.GitLabConnection{}, nil)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				src, err = src.(UserSource).WithAuthenticator(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HasType(err, UnsupportedAuthenticatorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func Test_maybeRefreshGitLabOAuthTokenFromCodeHost(t *testing.T) {
	tests := []struct {
		name    string
		expired bool
	}{
		{
			name:    "Expired token should be updated",
			expired: true,
		},
		{
			name:    "Not expired token should not be updated",
			expired: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var databaseHit bool
			var httpServerHit bool
			var newToken string

			// perms syncer mocking
			db := database.NewMockDB()
			externalServices := database.NewMockExternalServiceStore()
			externalServices.UpsertFunc.SetDefaultHook(func(ctx context.Context, services ...*types.ExternalService) error {
				databaseHit = true
				svc := services[0]
				parsed, err := extsvc.ParseConfig(extsvc.KindGitLab, svc.Config)
				if err != nil {
					t.Fatal(err)
				}
				config := parsed.(*schema.GitLabConnection)
				newToken = config.Token

				return nil
			})
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)

			// http mocking
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/oauth/token" {
					t.Errorf("Expected to request '/oauth/token', got: %s", r.URL.Path)
				}
				httpServerHit = true
				w.Header().Set("Content-Type", "application/json")
				refreshedToken := json.RawMessage(fmt.Sprintf(`
		{
			"access_token":"cafebabea66306277915a6919a90ac7972853317d9df385a828b17d9200b7d4c",
			"token_type":"Bearer",
			"refresh_token":"cafebabe251f4c2295494ee29b6b66f7011dad92251ab988a376a23ef12ad041",
			"expiry":"%s"
		}`,
					time.Now().Add(2*time.Hour).Format(time.RFC3339)))
				w.Write(refreshedToken)
			}))
			t.Cleanup(func() { server.Close() })

			// conf mocking
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				AuthProviders: []schema.AuthProviders{
					{
						Gitlab: &schema.GitLabAuthProvider{
							ClientID:     "clientId",
							ClientSecret: "clientSecret",
							Url:          fmt.Sprintf("%s/", server.URL),
						},
					},
				},
			}})
			t.Cleanup(func() { conf.Mock(nil) })

			// test data mocking
			expiryDate := time.Now().Add(1 * time.Hour)
			if test.expired {
				expiryDate = expiryDate.Add(-2 * time.Hour)
			}

			svc := &types.ExternalService{
				ID:   1,
				Kind: extsvc.KindGitLab,
				Config: fmt.Sprintf(`{
   "url": "%s",
   "token": "af865c51fb0ac7f7b6714ce25d837ad42f13f57006b651a592c810ac93d2e2cc",
   "token.type": "oauth",
   "token.oauth.refresh": "b84b0f06e306d1747ee9e87ef310aaa8784cc688d5a41590bee585634374f0c3",
   "token.oauth.expiry": %d,
   "projectQuery": [
     "projects?id_before=0"
   ],
   "authorization": {
     "identityProvider": {
       "type": "oauth"
     }
   }
 }`, server.URL, expiryDate.Unix()),
			}

			refreshed, err := maybeRefreshGitLabOAuthTokenFromCodeHost(context.Background(), logtest.Scoped(t), db, svc)
			if err != nil {
				t.Error(err)
			}

			// When token is expired, DB and HTTP server should be hit (for token update)
			want := test.expired
			if want != databaseHit {
				t.Errorf("Database hit:\ngot: %v\nwant: %v", databaseHit, want)
			}
			if want != httpServerHit {
				t.Errorf("HTTP Server hit:\ngot: %v\nwant: %v", httpServerHit, want)
			}
			if test.expired {
				wantToken := "cafebabea66306277915a6919a90ac7972853317d9df385a828b17d9200b7d4c"
				assert.Equal(t, wantToken, newToken)
				assert.Equal(t, wantToken, refreshed)
			}
		})
	}
}
