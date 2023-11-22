package repos

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

// TestSources_ListRepos_YieldExistingRepos is the main, happy-path test for
// listing repositories.
func TestSources_ListRepos_YieldExistingRepos(t *testing.T) {
	ratelimit.SetupForTest(t)
	rcache.SetupForTest(t)

	tests := []struct {
		svc       *types.ExternalService
		wantNames []string
	}{
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitHub, &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				Repos: []string{
					"sourcegraph/Sourcegraph",
					"tsenart/Vegeta",
					"tsenart/vegeta-missing",
				},
			}),
			wantNames: []string{
				"github.com/sourcegraph/sourcegraph",
				"github.com/tsenart/vegeta",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitLab, &schema.GitLabConnection{
				Url:          "https://gitlab.com",
				Token:        os.Getenv("GITLAB_ACCESS_TOKEN"),
				ProjectQuery: []string{"none"},
				Projects: []*schema.GitLabProject{
					{Name: "gnachman/iterm2"},
					{Name: "gnachman/iterm2-missing"},
					{Id: 13083}, // https://gitlab.com/gitlab-org/gitlab-ce
				},
			}),
			wantNames: []string{
				"gitlab.com/gitlab-org/gitlab-ce",
				"gitlab.com/gnachman/iterm2",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantBitbucketServer, &schema.BitbucketServerConnection{
				Url:             "https://bitbucket.sgdev.org",
				Token:           os.Getenv("BITBUCKET_SERVER_TOKEN"),
				RepositoryQuery: []string{"none"},
				Repos: []string{
					"Sour/vegetA",
					"sour/sourcegraph",
				},
			}),
			wantNames: []string{
				"bitbucket.sgdev.org/SOUR/sourcegraph",
				"bitbucket.sgdev.org/SOUR/vegeta",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantAWSCodeCommit, &schema.AWSCodeCommitConnection{
				AccessKeyID:     getAWSEnv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: getAWSEnv("AWS_SECRET_ACCESS_KEY"),
				Region:          "us-west-1",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "git-username",
					Password: "git-password",
				},
			}),
			wantNames: []string{
				"__WARNING_DO_NOT_PUT_ANY_PRIVATE_CODE_IN_HERE",
				"empty-repo",
				"stripe-go",
				"test",
				"test2",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantOther, &schema.OtherExternalServiceConnection{
				Url: "https://github.com",
				Repos: []string{
					"google/go-cmp",
				},
			}),
			wantNames: []string{
				"github.com/google/go-cmp",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.svc.Kind, func(t *testing.T) {
			name := tc.svc.Kind + "/included-repos-that-exist-are-yielded"

			cf, save := NewClientFactory(t, name)
			defer save(t)

			repos := listRepos(t, cf, tc.svc)

			var haveNames []string
			for _, r := range repos {
				haveNames = append(haveNames, string(r.Name))
			}
			sort.Strings(haveNames)

			if !reflect.DeepEqual(haveNames, tc.wantNames) {
				t.Error(cmp.Diff(haveNames, tc.wantNames))
			}
		})
	}
}

func TestSources_ListRepos_Excluded(t *testing.T) {
	conf.Mock(&conf.Unified{
		ServiceConnectionConfig: conftypes.ServiceConnections{
			GitServers: []string{"127.0.0.1:3178"},
		}, SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: boolPointer(false),
			},
		},
	})
	defer conf.Mock(nil)

	rcache.SetupForTest(t)
	ratelimit.SetupForTest(t)

	tests := []struct {
		svc       *types.ExternalService
		wantNames []string
	}{
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitHub, &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_ACCESS_TOKEN"),
				RepositoryQuery: []string{
					"user:tsenart in:name patrol", // yields only the tsenart/patrol repo
				},
				Repos: []string{
					"sourcegraph/sourcegraph",
					"keegancsmith/sqlf",
					"tsenart/VEGETA",
					"tsenart/go-tsz",     // fork
					"sourcegraph/about",  // has >500MB and < 200 stars
					"facebook/react",     // has 215k stars as of now
					"torvalds/linux",     // has ~4GB
					"avelino/awesome-go", // has < 20 MB and > 100k stars
				},
				Exclude: []*schema.ExcludedGitHubRepo{
					{Name: "tsenart/Vegeta"},
					{Id: "MDEwOlJlcG9zaXRvcnkxNTM2NTcyNDU="}, // tsenart/patrol ID
					{Pattern: "^keegancsmith/.*"},
					{Forks: true},
					{Stars: "> 215000"},                  // exclude facebook/react
					{Size: "> 3GB"},                      // exclude torvalds/linux
					{Size: ">= 500MB", Stars: "< 200"},   // exclude about repo
					{Size: "<= 20MB", Stars: "> 100000"}, // exclude awesome-go
				},
			}),
			wantNames: []string{
				"github.com/sourcegraph/sourcegraph",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitLab, &schema.GitLabConnection{
				Url:   "https://gitlab.com",
				Token: os.Getenv("GITLAB_ACCESS_TOKEN"),
				ProjectQuery: []string{
					"?search=gokulkarthick",
					"?search=dotfiles-vegetableman",
				},
				Exclude: []*schema.ExcludedGitLabProject{
					{Name: "gokulkarthick/gokulkarthick"},
					{Id: 7789240},
				},
			}),
			wantNames: []string{
				"gitlab.com/guld/dotfiles-vegetableman",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantBitbucketServer, &schema.BitbucketServerConnection{
				Url:   "https://bitbucket.sgdev.org",
				Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				Repos: []string{
					"SOUR/vegeta",
					"sour/sourcegraph",
				},
				RepositoryQuery: []string{
					"?visibility=private",
				},
				Exclude: []*schema.ExcludedBitbucketServerRepo{
					{Name: "SOUR/Vegeta"},      // test case insensitivity
					{Id: 10067},                // sourcegraph repo id
					{Pattern: ".*/automation"}, // only matches automation-testing repo
					{Pattern: ".*public-repo.*"},
					{Pattern: ".*secret-repo.*"},
					{Pattern: ".*private-repo.*"},
					{Pattern: ".*SGDEMO.*"},
				},
			}),
			wantNames: []string{
				"bitbucket.sgdev.org/IJ/ijt-repo-testing-sg-3.6",
				"bitbucket.sgdev.org/K8S/zoekt",
				"bitbucket.sgdev.org/PUBLIC/archived-repo",
				"bitbucket.sgdev.org/SOUR/sd",
				"bitbucket.sgdev.org/SOURCEGRAPH/jsonrpc2",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantAWSCodeCommit, &schema.AWSCodeCommitConnection{
				AccessKeyID:     getAWSEnv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: getAWSEnv("AWS_SECRET_ACCESS_KEY"),
				Region:          "us-west-1",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "git-username",
					Password: "git-password",
				},
				Exclude: []*schema.ExcludedAWSCodeCommitRepo{
					{Name: "stRIPE-gO"},
					{Id: "020a4751-0f46-4e19-82bf-07d0989b67dd"},                // ID of `test`
					{Name: "test2", Id: "2686d63d-bff4-4a3e-a94f-3e6df904238d"}, // ID of `test2`
				},
			}),
			wantNames: []string{
				"__WARNING_DO_NOT_PUT_ANY_PRIVATE_CODE_IN_HERE",
				"empty-repo",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitolite, &schema.GitoliteConnection{
				Prefix: "gitolite.mycorp.com/",
				Host:   "ssh://git@127.0.0.1:2222",
				Exclude: []*schema.ExcludedGitoliteRepo{
					{Name: "bar"},
					{Pattern: "gitolite-ad.*"},
				},
			}),
			wantNames: []string{
				"gitolite.mycorp.com/baz",
				"gitolite.mycorp.com/foo",
				"gitolite.mycorp.com/testing",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.svc.Kind, func(t *testing.T) {
			name := tc.svc.Kind + "/excluded-repos-are-never-yielded"

			cf, save := NewClientFactory(t, name)
			defer save(t)

			repos := listRepos(t, cf, tc.svc)

			var haveNames []string
			for _, r := range repos {
				haveNames = append(haveNames, string(r.Name))
			}
			sort.Strings(haveNames)

			if !reflect.DeepEqual(haveNames, tc.wantNames) {
				t.Error(cmp.Diff(haveNames, tc.wantNames))
			}
		})
	}
}

func TestSources_ListRepos_RepositoryPathPattern(t *testing.T) {
	// conf mock is required for gitolite
	conf.Mock(&conf.Unified{
		ServiceConnectionConfig: conftypes.ServiceConnections{
			GitServers: []string{"127.0.0.1:3178"},
		}, SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				EnableGRPC: boolPointer(false),
			},
		},
	})
	defer conf.Mock(nil)

	ratelimit.SetupForTest(t)
	rcache.SetupForTest(t)

	tests := []struct {
		svc       *types.ExternalService
		wantNames []string
		wantURIs  []string
	}{
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitHub, &schema.GitHubConnection{
				Url:                   "https://github.com",
				Token:                 os.Getenv("GITHUB_ACCESS_TOKEN"),
				RepositoryPathPattern: "{host}/a/b/c/{nameWithOwner}",
				RepositoryQuery:       []string{"none"},
				Repos:                 []string{"tsenart/vegeta"},
			}),
			wantNames: []string{"github.com/a/b/c/tsenart/vegeta"},
			wantURIs:  []string{"github.com/tsenart/vegeta"},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitLab, &schema.GitLabConnection{
				Url:                   "https://gitlab.com",
				Token:                 os.Getenv("GITLAB_ACCESS_TOKEN"),
				RepositoryPathPattern: "{host}/a/b/c/{pathWithNamespace}",
				ProjectQuery:          []string{"none"},
				Projects: []*schema.GitLabProject{
					{Name: "gnachman/iterm2"},
				},
			}),
			wantNames: []string{"gitlab.com/a/b/c/gnachman/iterm2"},
			wantURIs:  []string{"gitlab.com/gnachman/iterm2"},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantBitbucketServer, &schema.BitbucketServerConnection{
				Url:                   "https://bitbucket.sgdev.org",
				Token:                 os.Getenv("BITBUCKET_SERVER_TOKEN"),
				RepositoryPathPattern: "{host}/a/b/c/{projectKey}/{repositorySlug}",
				RepositoryQuery:       []string{"none"},
				Repos:                 []string{"sour/vegeta"},
			}),
			wantNames: []string{"bitbucket.sgdev.org/a/b/c/SOUR/vegeta"},
			wantURIs:  []string{"bitbucket.sgdev.org/SOUR/vegeta"},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantAWSCodeCommit, &schema.AWSCodeCommitConnection{
				AccessKeyID:     getAWSEnv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: getAWSEnv("AWS_SECRET_ACCESS_KEY"),
				Region:          "us-west-1",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "git-username",
					Password: "git-password",
				},
				RepositoryPathPattern: "a/b/c/{name}",
			}),
			wantNames: []string{
				"a/b/c/empty-repo",
				"a/b/c/stripe-go",
				"a/b/c/test2",
				"a/b/c/__WARNING_DO_NOT_PUT_ANY_PRIVATE_CODE_IN_HERE",
				"a/b/c/test",
			},
			wantURIs: []string{
				"empty-repo",
				"stripe-go",
				"test2",
				"__WARNING_DO_NOT_PUT_ANY_PRIVATE_CODE_IN_HERE",
				"test",
			},
		},
		{
			svc: typestest.MakeExternalService(t, extsvc.VariantGitolite, &schema.GitoliteConnection{
				// Prefix serves as a sort of repositoryPathPattern for Gitolite
				Prefix: "gitolite.mycorp.com/",
				Host:   "ssh://git@127.0.0.1:2222",
			}),
			wantNames: []string{
				"gitolite.mycorp.com/bar",
				"gitolite.mycorp.com/baz",
				"gitolite.mycorp.com/foo",
				"gitolite.mycorp.com/gitolite-admin",
				"gitolite.mycorp.com/testing",
			},
			wantURIs: []string{
				"gitolite.mycorp.com/bar",
				"gitolite.mycorp.com/baz",
				"gitolite.mycorp.com/foo",
				"gitolite.mycorp.com/gitolite-admin",
				"gitolite.mycorp.com/testing",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.svc.Kind, func(t *testing.T) {
			name := tc.svc.Kind + "/repositoryPathPattern-determines-the-repo-name"

			cf, save := NewClientFactory(t, name)
			defer save(t)

			repos := listRepos(t, cf, tc.svc)

			var haveURIs, haveNames []string
			for _, r := range repos {
				haveURIs = append(haveURIs, r.URI)
				haveNames = append(haveNames, string(r.Name))
			}

			if !reflect.DeepEqual(haveNames, tc.wantNames) {
				t.Error(cmp.Diff(haveNames, tc.wantNames))
			}
			if !reflect.DeepEqual(haveURIs, tc.wantURIs) {
				t.Error(cmp.Diff(haveURIs, tc.wantURIs))
			}
		})
	}
}

func TestSources_Phabricator(t *testing.T) {
	cf, save := NewClientFactory(t, "PHABRICATOR/phabricator")
	defer save(t)

	svc := typestest.MakeExternalService(t, extsvc.VariantPhabricator, &schema.PhabricatorConnection{
		Url:   "https://secure.phabricator.com",
		Token: os.Getenv("PHABRICATOR_TOKEN"),
	})

	repos := listRepos(t, cf, svc)

	if len(repos) == 0 {
		t.Fatalf("no repos yielded")
	}

	for _, r := range repos {
		repo := r.Metadata.(*phabricator.Repo)
		if repo.VCS != "git" {
			t.Fatalf("non git repo yielded: %+v", repo)
		}

		if repo.Status == "inactive" {
			t.Fatalf("inactive repo yielded: %+v", repo)
		}

		if repo.Name == "" {
			t.Fatalf("empty repo name: %+v", repo)
		}

		ext := api.ExternalRepoSpec{
			ID:          repo.PHID,
			ServiceType: extsvc.TypePhabricator,
			ServiceID:   "https://secure.phabricator.com",
		}

		if have, want := r.ExternalRepo, ext; have != want {
			t.Fatal(cmp.Diff(have, want))
		}
	}
}

func TestSources_ListRepos_GitLab_NameTransformations(t *testing.T) {
	ratelimit.SetupForTest(t)
	rcache.SetupForTest(t)

	cf, save := NewClientFactory(t, "GITLAB/nameTransformations-updates-the-repo-name")
	defer save(t)

	svc := typestest.MakeExternalService(t, extsvc.VariantGitLab, &schema.GitLabConnection{
		Url:                   "https://gitlab.com",
		Token:                 os.Getenv("GITLAB_ACCESS_TOKEN"),
		RepositoryPathPattern: "{host}/{pathWithNamespace}",
		ProjectQuery:          []string{"none"},
		Projects: []*schema.GitLabProject{
			{Name: "sg-test.d/repo-git"},
			{Name: "sg-test.d/repo-gitrepo"},
		},
		NameTransformations: []*schema.GitLabNameTransformation{
			{
				Regex:       "\\.d/",
				Replacement: "/",
			},
			{
				Regex:       "-git$",
				Replacement: "",
			},
		},
	})

	repos := listRepos(t, cf, svc)
	haveNames := types.Repos(repos).Names()
	sort.Strings(haveNames)

	wantNames := []string{
		"gitlab.com/sg-test/repo",
		"gitlab.com/sg-test/repo-gitrepo",
	}

	if !reflect.DeepEqual(haveNames, wantNames) {
		t.Error(cmp.Diff(haveNames, wantNames))
	}
}

func TestSources_ListRepos_BitbucketServer_Archived(t *testing.T) {
	ratelimit.SetupForTest(t)
	rcache.SetupForTest(t)

	cf, save := NewClientFactory(t, "BITBUCKETSERVER/bitbucketserver-archived")
	defer save(t)

	svc := typestest.MakeExternalService(t, extsvc.VariantBitbucketServer, &schema.BitbucketServerConnection{
		Url:                   "https://bitbucket.sgdev.org",
		Token:                 os.Getenv("BITBUCKET_SERVER_TOKEN"),
		RepositoryPathPattern: "{repositorySlug}",
		RepositoryQuery:       []string{"none"},
		Repos:                 []string{"sour/vegeta", "PUBLIC/archived-repo"},
	})

	repos := listRepos(t, cf, svc)

	wantArchived := map[string]bool{
		"vegeta":        false,
		"archived-repo": true,
	}

	got := map[string]bool{}
	for _, r := range repos {
		got[string(r.Name)] = r.Archived
	}

	if !reflect.DeepEqual(got, wantArchived) {
		t.Error("mismatch archived state (-want +got):\n", cmp.Diff(wantArchived, got))
	}
}

func listRepos(t *testing.T, cf *httpcli.Factory, svc *types.ExternalService) []*types.Repo {
	t.Helper()

	ctx := context.Background()

	logger := logtest.NoOp(t)
	sourcer := NewSourcer(logger, dbmocks.NewMockDB(), cf)

	src, err := sourcer(ctx, svc)
	if err != nil {
		t.Fatal(err)
	}

	repos, err := ListAll(ctx, src)
	if err != nil {
		t.Errorf("error listing repos: %s", err)
	}

	return repos
}

func newClientFactoryWithOpt(t testing.TB, name string, opt httpcli.Opt) (*httpcli.Factory, func(testing.TB)) {
	mw, rec := TestClientFactorySetup(t, name)
	return httpcli.NewFactory(mw, opt, httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { Save(t, rec) }
}

func getAWSEnv(envVar string) string {
	s := os.Getenv(envVar)
	if s == "" {
		s = fmt.Sprintf("BOGUS-%s", envVar)
	}
	return s
}

func boolPointer(b bool) *bool {
	return &b
}
