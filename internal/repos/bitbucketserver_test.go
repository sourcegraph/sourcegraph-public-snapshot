package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

type Client struct {
	s   *BitbucketServerSource
	err error
}

func New(serverUrl string) (*Client, error) {
	simpleConfig := &schema.BitbucketServerConnection{
		Url:   serverUrl,
		Token: "secret",
	}
	fmt.Println("Done creating config")

	svc := types.ExternalService{
		ID:   1,
		Kind: extsvc.KindBitbucketServer,
	}
	fmt.Println("Done creating svc")

	source, err := newBitbucketServerSource(&svc, simpleConfig, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Done creating source")

	return &Client{
		s:   source,
		err: err,
	}, nil
}

var (
	mux    *http.ServeMux
	server *httptest.Server
	// client *Client
)

func setup() *Client {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client, _ := New(server.URL)
	fmt.Println("Done creating client")

	return client
}

func TestListRepos(t *testing.T) {
	client := setup()
	client.s.config.RepositoryQuery = []string{"?projectname=\"foo\""}

	mux.HandleFunc("/rest/api/1.0", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, fixture("repos.json"))
	})

	mux.HandleFunc("/rest/api/1.0/labels/archived/labeled", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("REACHED ARCHIVED")
	})

	mux.HandleFunc("/rest/api/1.0/repos?projectname=foo", func(w http.ResponseWriter, r *http.Request) {

	})

	fmt.Println("Making results...")
	results := make(chan SourceResult)
	client.s.ListRepos(context.Background(), results)

	repoNameMap := map[string]struct{}{
		"python-langserver-fork": {},
		"python-langserver":      {},
		"golang-langserver":      {},
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	for i := 0; i < len(repoNameMap); i++ {
		select {
		case r := <-results:
			//verify result is in repoNameMap
		case <-ctx.Done():
			//fail test
			//break
		}
	}

	server.Close()
}

func fixture(path string) string {
	b, err := ioutil.ReadFile("testdata" + path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// func TestListReposv2(t *testing.T) {
// 	// make the bitbucket server source
// 	// connect the source
// 	// add some repos from test file
// 	// get back the repos

// 	// config := map[string]*schema.BitbucketServerConnection{
// 	// 	"simple": {
// 	// 		Url:   "https://bitbucket.example.com",
// 	// 		Token: "secret",
// 	// 	},
// 	// }
// 	// fmt.Println("Done creating config")

// 	simpleConfig := &schema.BitbucketServerConnection{
// 		Url:   "https://bitbucket.example.com",
// 		Token: "secret",
// 	}

// 	svc := types.ExternalService{
// 		ID:   1,
// 		Kind: extsvc.KindBitbucketServer,
// 	}
// 	fmt.Println("Done creating svc")

// 	s, err := newBitbucketServerSource(&svc, simpleConfig, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println("Done creating BitbucketServerSource")

// 	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var repos []*bitbucketserver.Repo
// 	if err := json.Unmarshal(b, &repos); err != nil {
// 		t.Fatal(err)
// 	}

// 	for _, repo := range repos {
// 		s.config.Repos = append(s.config.Repos, s.makeRepo(repo, false))
// 	}
// }

func TestListReposv1(t *testing.T) {
	fmt.Println("TestListRepos called...")

	config := map[string]*schema.BitbucketServerConnection{
		"simple": {
			Url:   "https:/example.com/",
			Token: "secret",
		},
	}
	fmt.Println("Done with config")

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketServer}
	fmt.Println("Done with svc")

	s, err := newBitbucketServerSource(&svc, config["simple"], nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("simple", func(t *testing.T) {
		results := make(chan SourceResult)
		fmt.Println("Done making results channel")
		s.ListRepos(context.Background(), results)
		fmt.Println("Done ListRepos, printing results")

		r := <-results
		fmt.Println(r)
		fmt.Println("Done printing r")
	})

}
