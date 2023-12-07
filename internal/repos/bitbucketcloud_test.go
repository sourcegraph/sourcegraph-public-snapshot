package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	bbtest "github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud/testing"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketCloudSource_ListRepos(t *testing.T) {
	ratelimit.SetupForTest(t)

	assertAllReposListed := func(want []string) typestest.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			have := rs.Names()
			sort.Strings(have)
			sort.Strings(want)

			if diff := cmp.Diff(want, have); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		}
	}

	testCases := []struct {
		name   string
		assert typestest.ReposAssertion
		conf   *schema.BitbucketCloudConnection
		err    string
	}{
		{
			name: "found",
			assert: assertAllReposListed([]string{
				"/sourcegraph-testing/src-cli",
				"/sourcegraph-testing/sourcegraph",
			}),
			conf: &schema.BitbucketCloudConnection{
				Username:    bbtest.GetenvTestBitbucketCloudUsername(),
				AppPassword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
				Teams: []string{
					bbtest.GetenvTestBitbucketCloudUsername(),
				},
			},
			err: "<nil>",
		},
		{
			name: "with teams",
			assert: assertAllReposListed([]string{
				"/sglocal/go-langserver",
				"/sglocal/python-langserver",
				"/sourcegraph-testing/src-cli",
				"/sourcegraph-testing/sourcegraph",
			}),
			conf: &schema.BitbucketCloudConnection{
				Username:    bbtest.GetenvTestBitbucketCloudUsername(),
				AppPassword: os.Getenv("BITBUCKET_CLOUD_APP_PASSWORD"),
				Teams: []string{
					"sglocal",
					bbtest.GetenvTestBitbucketCloudUsername(),
				},
			},
			err: "<nil>",
		},
		{
			name: "with access token",
			assert: assertAllReposListed([]string{
				"/sourcegraph-source/src-cli",
				"/sourcegraph-source/source-test",
			}),
			conf: &schema.BitbucketCloudConnection{
				AccessToken: os.Getenv("BITBUCKET_CLOUD_ACCESS_TOKEN"),
				Teams: []string{
					"sourcegraph-source",
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BITBUCKETCLOUD-LIST-REPOS/" + tc.name
		t.Run(tc.name, func(t *testing.T) {
			cf, save := NewClientFactory(t, tc.name)
			defer save(t)

			svc := &types.ExternalService{
				Kind:   extsvc.KindBitbucketCloud,
				Config: extsvc.NewUnencryptedConfig(MarshalJSON(t, tc.conf)),
			}

			bbcSrc, err := newBitbucketCloudSource(logtest.Scoped(t), svc, tc.conf, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := ListAll(context.Background(), bbcSrc)

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
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketcloud.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	svc := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindBitbucketCloud,
		Config: extsvc.NewEmptyConfig(),
	}

	tests := []struct {
		name   string
		schema *schema.BitbucketCloudConnection
	}{
		{
			name: "simple",
			schema: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    "alice",
				AppPassword: "secret",
			},
		}, {
			name: "ssh",
			schema: &schema.BitbucketCloudConnection{
				Url:         "https://bitbucket.org",
				Username:    "alice",
				AppPassword: "secret",
				GitURLType:  "ssh",
			},
		}, {
			name: "path-pattern",
			schema: &schema.BitbucketCloudConnection{
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
			s, err := newBitbucketCloudSource(logtest.Scoped(t), &svc, test.schema, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*types.Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}

			testutil.AssertGolden(t, "testdata/golden/"+test.name, Update(test.name), got)
		})
	}
}

func TestBitbucketCloudSource_Exclude(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
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

	svc := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindBitbucketCloud,
		Config: extsvc.NewEmptyConfig(),
	}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketCloudSource(logtest.Scoped(t), &svc, config, nil)
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
			testutil.AssertGolden(t, path, Update(name), got)
		})
	}
}
