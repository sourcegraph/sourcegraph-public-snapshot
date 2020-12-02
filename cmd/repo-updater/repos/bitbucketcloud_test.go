package repos

import (
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
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
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

			svc := &types.ExternalService{
				Kind:   extsvc.KindBitbucketCloud,
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

func TestBitbucketCloudSource_makeRepo(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketcloud.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketCloud}

	tests := []struct {
		name   string
		schmea *schema.BitbucketCloudConnection
	}{
		{
			name: "simple",
			schmea: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    "alice",
				AppPassword: "secret",
			},
		}, {
			name: "ssh",
			schmea: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    "alice",
				AppPassword: "secret",
				GitURLType:  "ssh",
			},
		}, {
			name: "path-pattern",
			schmea: &schema.BitbucketCloudConnection{
				Url:                   "https://bitbucket.org",
				Username:              "alice",
				AppPassword:           "secret",
				RepositoryPathPattern: "bb/{nameWithOwner}",
			},
		},
	}
	for _, test := range tests {
		test.name = "BitbucketCloudSource_makeRepo_" + test.name
		t.Run(test.name, func(t *testing.T) {
			s, err := newBitbucketCloudSource(&svc, test.schmea, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}

			testutil.AssertGolden(t, "testdata/golden/"+test.name, update(test.name), got)
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

	svc := types.ExternalService{ID: 1, Kind: extsvc.KindBitbucketCloud}

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

			path := filepath.Join("testdata", "bitbucketcloud-repos-exclude-"+name+".golden")
			testutil.AssertGolden(t, path, update(name), got)
		})
	}
}
