package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

func TestBitbucketCloudSource_ListRepos(t *testing.T) {
	assertAllReposListed := func(want []string) ReposAssertion {
		return func(t testing.TB, rs Repos) {
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
		assert ReposAssertion
		conf   *schema.BitbucketCloudConnection
		err    string
	}{
		{
			name: "found",
			assert: assertAllReposListed([]string{
				"bitbucket.org/Unknwon/boilerdb",
				"bitbucket.org/Unknwon/scripts",
				"bitbucket.org/Unknwon/wxvote",
			}),
			conf: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    bitbucketcloud.GetenvTestBitbucketCloudUsername(),
				AppPassword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
			},
			err: "<nil>",
		},
		{
			name: "with teams",
			assert: assertAllReposListed([]string{
				"bitbucket.org/Unknwon/boilerdb",
				"bitbucket.org/Unknwon/scripts",
				"bitbucket.org/Unknwon/wxvote",
				"bitbucket.org/sglocal/mux",
				"bitbucket.org/sglocal/go-langserver",
				"bitbucket.org/sglocal/python-langserver",
			}),
			conf: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    bitbucketcloud.GetenvTestBitbucketCloudUsername(),
				AppPassword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
				Teams: []string{
					"sglocal",
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BITBUCKETCLOUD-LIST-REPOS/" + tc.name
		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind:   "BITBUCKETCLOUD",
				Config: marshalJSON(t, tc.conf),
			}

			bbcSrc, err := newBitbucketCloudSource(svc, tc.conf, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := listAll(context.Background(), bbcSrc)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}

func TestBitbucketCloudSource_MakeRepo(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketcloud.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketCloudConnection{
		"simple": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
		},
		"ssh": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
			GitURLType:  "ssh",
		},
		"path-pattern": {
			Url:                   "https://bitbucket.org",
			Username:              "alice",
			AppPassword:           "secret",
			RepositoryPathPattern: "bb/{nameWithOwner}",
		},
	}

	svc := ExternalService{ID: 1, Kind: "BITBUCKETCLOUD"}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketCloudSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}
			actual, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", "bitbucketcloud-repos-"+name+".golden")
			if update(name) {
				err := ioutil.WriteFile(golden, actual, 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			expect, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expect) {
				d, err := testutil.Diff(string(actual), string(expect))
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
		})
	}
}

func TestBitbucketCloudSource_Exclude(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketcloud.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketCloudConnection{
		"none": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
		},
		"name": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
			Exclude: []*schema.ExcludedBitbucketCloudRepo{
				{Name: "SG/go-langserver"},
			},
		},
		"uuid": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
			Exclude: []*schema.ExcludedBitbucketCloudRepo{
				{Uuid: "{fceb73c7-cef6-4abe-956d-e471281126bd}"},
			},
		},
		"pattern": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
			Exclude: []*schema.ExcludedBitbucketCloudRepo{
				{Pattern: ".*-fork$"},
			},
		},
		"all": {
			Url:         "https://bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
			Exclude: []*schema.ExcludedBitbucketCloudRepo{
				{Name: "SG/go-LanGserVer"},
				{Uuid: "{fceb73c7-cef6-4abe-956d-e471281126bd}"},
				{Pattern: ".*-fork$"},
			},
		},
	}

	svc := ExternalService{ID: 1, Kind: "BITBUCKETCLOUD"}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketCloudSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			var got output
			for _, r := range repos {
				if s.excludes(r) {
					got.Exclude = append(got.Exclude, r.FullName)
				} else {
					got.Include = append(got.Include, r.FullName)
				}
			}
			actual, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", "bitbucketcloud-repos-exclude-"+name+".golden")
			if update(name) {
				err := ioutil.WriteFile(golden, actual, 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			expect, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expect) {
				d, err := testutil.Diff(string(actual), string(expect))
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
		})
	}
}
