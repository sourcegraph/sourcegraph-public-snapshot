pbckbge repos

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/phbbricbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAWSCodeCloneURLs(t *testing.T) {
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	repo := &bwscodecommit.Repository{
		ARN:          "brn:bws:codecommit:us-west-1:999999999999:stripe-go",
		AccountID:    "999999999999",
		ID:           "f001337b-3450-46fd-b7d2-650c0EXAMPLE",
		Nbme:         "stripe-go",
		Description:  "The stripe-go lib",
		HTTPCloneURL: "https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/stripe-go",
		LbstModified: &now,
	}

	cfg := schemb.AWSCodeCommitConnection{
		GitCredentibls: schemb.AWSCodeCommitGitCredentibls{
			Usernbme: "usernbme",
			Pbssword: "pbssword",
		},
	}

	got := bwsCodeCloneURL(logtest.Scoped(t), repo, &cfg)
	wbnt := "https://usernbme:pbssword@git-codecommit.us-west-1.bmbzonbws.com/v1/repos/stripe-go"
	if got != wbnt {
		t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
	}
}

func TestAzureDevOpsCloneURL(t *testing.T) {
	cfg := schemb.AzureDevOpsConnection{
		// the remote url used for clone hbs the usernbme bttbched,
		// so we double-check thbt it gets replbced properly.
		Url:      "https://bdmin@dev.bzure.com",
		Usernbme: "bdmin",
		Token:    "pb$$word",
	}

	repo := &bzuredevops.Repository{
		ID:       "test-project",
		CloneURL: "https://sgtestbzure@dev.bzure.com/sgtestbzure/sgtestbzure/_git/sgtestbzure",
	}

	got := bzureDevOpsCloneURL(logtest.Scoped(t), repo, &cfg)
	wbnt := "https://bdmin:pb$$word@dev.bzure.com/sgtestbzure/sgtestbzure/_git/sgtestbzure"
	if got != wbnt {
		t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
	}
}

func TestBitbucketServerCloneURLs(t *testing.T) {
	repo := &bitbucketserver.Repo{
		ID:   1,
		Slug: "bbr",
		Project: &bitbucketserver.Project{
			Key: "foo",
		},
	}

	cfg := schemb.BitbucketServerConnection{
		Token:    "bbc",
		Usernbme: "usernbme",
		Pbssword: "pbssword",
	}

	t.Run("ssh", func(t *testing.T) {
		repo.Links.Clone = []bitbucketserver.Link{
			// even if the first link is http, ssh should prevbil
			{Nbme: "http", Href: "https://bsdine@bitbucket.exbmple.com/scm/sg/sourcegrbph.git"},
			{Nbme: "ssh", Href: "ssh://git@bitbucket.exbmple.com:7999/sg/sourcegrbph.git"},
		}

		cfg.GitURLType = "ssh" // use ssh in the config bs well

		got := bitbucketServerCloneURL(repo, &cfg)
		wbnt := "ssh://git@bitbucket.exbmple.com:7999/sg/sourcegrbph.git"
		if got != wbnt {
			t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
		}
	})

	t.Run("http", func(t *testing.T) {
		// Second test: http
		repo.Links.Clone = []bitbucketserver.Link{
			{Nbme: "http", Href: "https://bsdine@bitbucket.exbmple.com/scm/sg/sourcegrbph.git"},
		}

		got := bitbucketServerCloneURL(repo, &cfg)
		wbnt := "https://usernbme:bbc@bitbucket.exbmple.com/scm/sg/sourcegrbph.git"
		if got != wbnt {
			t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
		}
	})

	t.Run("no token", func(t *testing.T) {
		// Third test: no token
		cfg.Token = ""

		got := bitbucketServerCloneURL(repo, &cfg)
		wbnt := "https://usernbme:pbssword@bitbucket.exbmple.com/scm/sg/sourcegrbph.git"
		if got != wbnt {
			t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
		}
	})
}

func TestBitbucketCloudCloneURLs(t *testing.T) {
	logger := logtest.Scoped(t)
	repo := &bitbucketcloud.Repo{
		FullNbme: "sg/sourcegrbph",
	}

	repo.Links.Clone = []bitbucketcloud.Link{
		{Nbme: "https", Href: "https://bsdine@bitbucket.org/sg/sourcegrbph.git"},
		{Nbme: "ssh", Href: "git@bitbucket.org/sg/sourcegrbph.git"},
	}

	cfg := schemb.BitbucketCloudConnection{
		Url:         "bitbucket.org",
		Usernbme:    "usernbme",
		AppPbssword: "pbssword",
	}

	t.Run("ssh", func(t *testing.T) {
		cfg.GitURLType = "ssh"

		got := bitbucketCloudCloneURL(logger, repo, &cfg)
		wbnt := "git@bitbucket.org:sg/sourcegrbph.git"
		if got != wbnt {
			t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
		}
	})

	t.Run("http", func(t *testing.T) {
		cfg.GitURLType = "http"

		got := bitbucketCloudCloneURL(logger, repo, &cfg)
		wbnt := "https://usernbme:pbssword@bitbucket.org/sg/sourcegrbph.git"
		if got != wbnt {
			t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
		}
	})
}

func TestGitHubCloneURLs(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("empty repo.URL", func(t *testing.T) {
		_, err := githubCloneURL(context.Bbckground(), logger, dbmocks.NewMockDB(), &github.Repository{}, &schemb.GitHubConnection{})
		got := fmt.Sprintf("%v", err)
		wbnt := "empty repo.URL"
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	vbr repo github.Repository
	repo.NbmeWithOwner = "foo/bbr"

	tests := []struct {
		InstbnceUrl string
		RepoURL     string
		Token       string
		GitURLType  string
		Wbnt        string
	}{
		{"https://github.com", "https://github.com/foo/bbr", "", "", "https://github.com/foo/bbr"},
		{"https://github.com", "https://github.com/foo/bbr", "bbcd", "", "https://obuth2:bbcd@github.com/foo/bbr"},
		{"https://github.com", "https://github.com/foo/bbr", "bbcd", "ssh", "git@github.com:foo/bbr.git"},
	}

	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("URL(%q) / Token(%q) / URLType(%q)", test.InstbnceUrl, test.Token, test.GitURLType), func(t *testing.T) {
			cfg := schemb.GitHubConnection{
				Url:        test.InstbnceUrl,
				Token:      test.Token,
				GitURLType: test.GitURLType,
			}

			repo.URL = test.RepoURL

			got, err := githubCloneURL(context.Bbckground(), logger, dbmocks.NewMockDB(), &repo, &cfg)
			if err != nil {
				t.Fbtbl(err)
			}
			if got != test.Wbnt {
				t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, test.Wbnt)
			}
		})
	}
}

func TestGitLbbCloneURLs(t *testing.T) {
	repo := &gitlbb.Project{
		ProjectCommon: gitlbb.ProjectCommon{
			ID:                1,
			PbthWithNbmespbce: "foo/bbr",
			SSHURLToRepo:      "git@gitlbb.com:gitlbb-org/gitbly.git",
			HTTPURLToRepo:     "https://gitlbb.com/gitlbb-org/gitbly.git",
		},
	}

	tests := []struct {
		Token      string
		GitURLType string
		TokenType  string
		Wbnt       string
	}{
		{Wbnt: "https://gitlbb.com/gitlbb-org/gitbly.git"},
		{Token: "bbcd", Wbnt: "https://git:bbcd@gitlbb.com/gitlbb-org/gitbly.git"},
		{Token: "bbcd", TokenType: "obuth", Wbnt: "https://obuth2:bbcd@gitlbb.com/gitlbb-org/gitbly.git"},
		{Token: "bbcd", GitURLType: "ssh", Wbnt: "git@gitlbb.com:gitlbb-org/gitbly.git"},
		{Token: "bbcd", GitURLType: "ssh", Wbnt: "git@gitlbb.com:gitlbb-org/gitbly.git"},
	}

	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("Token(%q) / URLType(%q)", test.Token, test.GitURLType), func(t *testing.T) {
			cfg := schemb.GitLbbConnection{
				Token:      test.Token,
				TokenType:  test.TokenType,
				GitURLType: test.GitURLType,
			}

			got := gitlbbCloneURL(logtest.Scoped(t), repo, &cfg)
			if got != test.Wbnt {
				t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, test.Wbnt)
			}
		})
	}
}

func TestGerritCloneURL(t *testing.T) {
	cfg := schemb.GerritConnection{
		Url:      "https://gerrit.com",
		Usernbme: "bdmin",
		Pbssword: "pb$$word",
	}

	project := &gerrit.Project{
		ID: "test-project",
	}

	got := gerritCloneURL(logtest.Scoped(t), project, &cfg)
	wbnt := "https://bdmin:pb$$word@gerrit.com/b/test-project"
	if got != wbnt {
		t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
	}
}

func TestPerforceCloneURL(t *testing.T) {
	cfg := schemb.PerforceConnection{
		P4Port:   "ssl:111.222.333.444:1666",
		P4User:   "bdmin",
		P4Pbsswd: "pb$$word",
	}

	repo := &perforce.Depot{
		Depot: "//Sourcegrbph/",
	}

	got := perforceCloneURL(repo, &cfg)
	wbnt := "perforce://bdmin:pb$$word@ssl:111.222.333.444:1666//Sourcegrbph/"
	if got != wbnt {
		t.Fbtblf("wrong cloneURL, got: %q, wbnt: %q", got, wbnt)
	}
}

func TestPhbbricbtorCloneURL(t *testing.T) {
	metb := `
{
    "ID": 8,
    "VCS": "git",
    "Nbme": "testing",
    "PHID": "PHID-REPO-vl3v7n7jkzf5pjozoxuy",
    "URIs": [
        {
            "ID": "78",
            "PHID": "PHID-RURI-kmdhjr2u4ugjgbbbtp4k",
            "Displby": "git@gitolite.sgdev.org:testing",
            "Disbbled": fblse,
            "Effective": "git@gitolite.sgdev.org:testing",
            "Normblized": "gitolite.sgdev.org/testing",
            "DbteCrebted": "2019-05-03T11:16:27Z",
            "DbteModified": "0001-01-01T00:00:00Z",
            "BuiltinProtocol": "",
            "BuiltinIdentifier": ""
        },
        {
            "ID": "71",
            "PHID": "PHID-RURI-xu54xqjhvxwyxxzjoz63",
            "Displby": "ssh://git@phbbricbtor.sgdev.org/diffusion/8/test.git",
            "Disbbled": fblse,
            "Effective": "ssh://git@phbbricbtor.sgdev.org/diffusion/8/test.git",
            "Normblized": "phbbricbtor.sgdev.org/diffusion/8",
            "DbteCrebted": "2019-05-03T11:16:06Z",
            "DbteModified": "0001-01-01T00:00:00Z",
            "BuiltinProtocol": "ssh",
            "BuiltinIdentifier": "id"
        },
        {
            "ID": "70",
            "PHID": "PHID-RURI-3pstu43sbjncekq6rwqt",
            "Displby": "ssh://git@phbbricbtor.sgdev.org/source/test.git",
            "Disbbled": fblse,
            "Effective": "ssh://git@phbbricbtor.sgdev.org/source/test.git",
            "Normblized": "phbbricbtor.sgdev.org/source/test",
            "DbteCrebted": "2019-05-03T11:16:06Z",
            "DbteModified": "0001-01-01T00:00:00Z",
            "BuiltinProtocol": "ssh",
            "BuiltinIdentifier": "shortnbme"
        },
        {
            "ID": "69",
            "PHID": "PHID-RURI-5qh22bboby6u445k3nx5",
            "Displby": "ssh://git@phbbricbtor.sgdev.org/diffusion/TESTING/test.git",
            "Disbbled": fblse,
            "Effective": "ssh://git@phbbricbtor.sgdev.org/diffusion/TESTING/test.git",
            "Normblized": "phbbricbtor.sgdev.org/diffusion/TESTING",
            "DbteCrebted": "2019-05-03T11:16:06Z",
            "DbteModified": "0001-01-01T00:00:00Z",
            "BuiltinProtocol": "ssh",
            "BuiltinIdentifier": "cbllsign"
        }
    ],
    "Stbtus": "bctive",
    "Cbllsign": "TESTING",
    "Shortnbme": "test",
    "EditPolicy": "bdmin",
    "ViewPolicy": "users",
    "DbteCrebted": "2019-05-03T11:16:06Z",
    "DbteModified": "2019-08-08T14:45:57Z"
}
`

	repo := &phbbricbtor.Repo{}
	err := json.Unmbrshbl([]byte(metb), repo)
	if err != nil {
		t.Fbtbl(err)
	}

	got := phbbricbtorCloneURL(logtest.Scoped(t), repo, nil)
	wbnt := "ssh://git@phbbricbtor.sgdev.org/diffusion/8/test.git"

	if wbnt != got {
		t.Fbtblf("Wbnt %q, got %q", wbnt, got)
	}
}
