package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketServerSource_MakeRepo(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	fmt.Println("Printing repos...")
	for _, repo := range repos {
		fmt.Printf("%+v\n", repo)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"simple": {
			Url:   "bitbucket.example.com",
			Token: "secret",
		},
		"ssh": {
			Url:                         "https://bitbucket.example.com",
			Token:                       "secret",
			InitialRepositoryEnablement: true,
			GitURLType:                  "ssh",
		},
		"path-pattern": {
			Url:                   "https://bitbucket.example.com",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
		"username": {
			Url:                   "https://bitbucket.example.com",
			Username:              "foo",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			fmt.Println("Name:", name)
			// fmt.Printf("Config: %+v\n", config)
			s, err := newBitbucketServerSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*types.Repo
			fmt.Println("Repos:")
			for _, r := range repos {
				// fmt.Println("R:", r)
				got = append(got, s.makeRepo(r, false))
			}

			path := filepath.Join("testdata", "bitbucketserver-repos-"+name+".golden")
			testutil.AssertGolden(t, path, update(name), got)
		})
	}
}

func TestBitbucketServerSource_Exclude(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"none": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
		},
		"name": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Name: "SG/python-langserver-fork",
			}, {
				Name: "~KEEGAN/rgp",
			}},
		},
		"id": {
			Url:     "https://bitbucket.example.com",
			Token:   "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{Id: 4}},
		},
		"pattern": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Pattern: "SG/python.*",
			}, {
				Pattern: "~KEEGAN/.*",
			}},
		},
		"both": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			// We match on the bitbucket server repo name, not the repository path pattern.
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Id: 1,
			}, {
				Name: "~KEEGAN/rgp",
			}, {
				Pattern: ".*-fork",
			}},
		},
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketServerSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			var got output
			for _, r := range repos {
				name := r.Slug
				if r.Project != nil {
					name = r.Project.Key + "/" + name
				}
				if s.excludes(r) {
					got.Exclude = append(got.Exclude, name)
				} else {
					got.Include = append(got.Include, name)
				}
			}

			path := filepath.Join("testdata", "bitbucketserver-repos-exclude-"+name+".golden")
			testutil.AssertGolden(t, path, update(name), got)
		})
	}
}

func TestBitbucketServerSource_WithAuthenticator(t *testing.T) {
	svc := &types.ExternalService{
		Kind: extsvc.KindBitbucketServer,
		Config: marshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bitbucket.sgdev.org",
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		}),
	}

	bbsSrc, err := NewBitbucketServerSource(svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("supported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"BasicAuth":           &auth.BasicAuth{},
			"OAuthBearerToken":    &auth.OAuthBearerToken{},
			"SudoableOAuthClient": &bitbucketserver.SudoableOAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				if gs, ok := src.(*BitbucketServerSource); !ok {
					t.Error("cannot coerce Source into bbsSource")
				} else if gs == nil {
					t.Error("unexpected nil Source")
				}
			})
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
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

func TestListRepos(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}

	var repos []bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/rest/api/1.0/labels/archived/labeled", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("LABELED ENDPOINT HIT")
	})

	mux.HandleFunc("/rest/api/1.0/repos", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("REPOS ENDPOINT HIT")

		projectName := r.URL.Query().Get("projectName")
		for _, repo := range repos {
			repoProjectName := "/" + repo.Project.Key + "/" + repo.Slug
			if projectName == repoProjectName {
				// fmt.Println("===== MATCH =====, Repo:", repo.Name)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(struct {
					PageToken bitbucketserver.PageToken `json:"pageToken"`
					Values    any                       `json:"values"`
				}{
					PageToken: bitbucketserver.PageToken{
						Size:          1,
						Limit:         1000,
						IsLastPage:    true,
						Start:         1,
						NextPageStart: 1,
					},
					Values: []bitbucketserver.Repo{repo},
				})
			}
		}
	})

	mux.HandleFunc("/rest/api/1.0/projects/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("PROJECTS ENDPOINT HIT")

		pathArr := strings.Split(r.URL.Path, "/")
		projectKey := pathArr[5]
		repoSlug := pathArr[7]

		for _, repo := range repos {
			if repo.Project.Key == projectKey && repo.Slug == repoSlug {
				// fmt.Println("==== MATCH ====, Repo:", repo.Name)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(repo)
			}
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	simpleConfig := &schema.BitbucketServerConnection{
		Url:   server.URL,
		Token: "secret",
	}

	svc := types.ExternalService{
		ID:   1,
		Kind: extsvc.KindBitbucketServer,
	}

	s, err := newBitbucketServerSource(&svc, simpleConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	s.config.Repos = []string{
		"SG/go-langserver",
		"SG/python-langserver",
		"SG/python-langserver/fork",
		"~KEEGAN/rgp",
		"~KEEGAN/rgp-unavailable",
	}

	s.config.RepositoryQuery = []string{
		"?projectName=/SG/go-langserver",
		"?projectName=/SG/python-langserver",
		"?projectName=/SG/python-langserver-fork",
		"?projectName=/~KEEGAN/rgp",
		"?projectName=/~KEEGAN/rgp-unavailable",
		"none",
		// "all",
	}

	ctxWithTimeout, cancelFunction := context.WithTimeout(context.Background(), 5*time.Second)
	fmt.Println(cancelFunction)

	results := make(chan SourceResult, 5)
	defer close(results)

	s.ListRepos(ctxWithTimeout, results)

	// go func() {
	// 	fmt.Println("Printing", len(results), "results...")
	// 	for res := range results {
	// 		fmt.Printf("%+v\n", *(res.Repo))
	// 	}
	// }()

	repoNameMap := map[string]struct{}{
		"/SG/go-langserver":          {},
		"/SG/python-langserver":      {},
		"/SG/python-langserver-fork": {},
		"/~KEEGAN/rgp":               {},
		"/~KEEGAN/rgp-unavailable":   {},
	}

	for {
		select {
		case res := <-results:
			repoNameArr := strings.Split(string(res.Repo.Name), "/")
			repoName := "/" + repoNameArr[1] + "/" + repoNameArr[2]
			fmt.Println("Repo:", repoName)
			if _, ok := repoNameMap[repoName]; ok {
				fmt.Println("Verified")
			}
		case <-ctxWithTimeout.Done():
			fmt.Println("Timeout")
			// cancelFunction()
			t.Fatal(err)
		default:
			return
		}
	}

}

// func TestListReposv1(t *testing.T) {

// 	fmt.Println("Making results...")
// 	results := make(chan SourceResult)
// 	client.s.ListRepos(context.Background(), results)

// 	repoNameMap := map[string]struct{}{
// 		"python-langserver-fork": {},
// 		"python-langserver":      {},
// 		"golang-langserver":      {},
// 	}

// 	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

// 	for i := 0; i < len(repoNameMap); i++ {
// 		select {
// 		case r := <-results:
// 			//verify result is in repoNameMap
// 		case <-ctx.Done():
// 			//fail test
// 			//break
// 		}
// 	}
// }
