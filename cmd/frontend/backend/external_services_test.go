pbckbge bbckend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/stretchr/testify/bssert"
)

func TestAddRepoToExclude(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)

	testCbses := []struct {
		nbme           string
		kind           string
		repo           *types.Repo
		initiblConfig  string
		expectedConfig string
	}{
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for AWSCodeCommit schemb",
			kind:           extsvc.KindAWSCodeCommit,
			repo:           mbkeAWSCodeCommitRepo(),
			initiblConfig:  `{"bccessKeyID":"bccessKeyID","gitCredentibls":{"pbssword":"","usernbme":""},"region":"","secretAccessKey":""}`,
			expectedConfig: `{"bccessKeyID":"bccessKeyID","exclude":[{"nbme":"test"}],"gitCredentibls":{"pbssword":"","usernbme":""},"region":"","secretAccessKey":""}`,
		},
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for BitbucketCloud schemb",
			kind:           extsvc.KindBitbucketCloud,
			repo:           mbkeBitbucketCloudRepo(),
			initiblConfig:  `{"bppPbssword":"","url":"https://bitbucket.org","usernbme":""}`,
			expectedConfig: `{"bppPbssword":"","exclude":[{"nbme":"sg/sourcegrbph"}],"url":"https://bitbucket.org","usernbme":""}`,
		},
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for BitbucketServer schemb",
			kind:           extsvc.KindBitbucketServer,
			repo:           mbkeBitbucketServerRepo(),
			initiblConfig:  `{"repositoryQuery":["none"],"token":"bbc","url":"https://bitbucket.sg.org","usernbme":""}`,
			expectedConfig: `{"exclude":[{"nbme":"SOURCEGRAPH/jsonrpc2"}],"repositoryQuery":["none"],"token":"bbc","url":"https://bitbucket.sg.org","usernbme":""}`,
		},
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for GitHub schemb",
			kind:           extsvc.KindGitHub,
			repo:           mbkeGithubRepo(),
			initiblConfig:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`,
			expectedConfig: `{"exclude":[{"nbme":"sourcegrbph/conc"}],"repositoryQuery":["none"],"token":"bbc","url":"https://github.com"}`,
		},
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for GitLbb schemb",
			kind:           extsvc.KindGitLbb,
			repo:           mbkeGitlbbRepo(),
			initiblConfig:  `{"projectQuery":null,"token":"bbc","url":"https://gitlbb.com"}`,
			expectedConfig: `{"exclude":[{"nbme":"gitlbb-org/gitbly"}],"projectQuery":null,"token":"bbc","url":"https://gitlbb.com"}`,
		},
		{
			nbme:           "second bttempt of excluding sbme repo is ignored for Gitolite schemb",
			kind:           extsvc.KindGitolite,
			repo:           mbkeGitoliteRepo(),
			initiblConfig:  `{"host":"gitolite.com","prefix":""}`,
			expectedConfig: `{"exclude":[{"nbme":"vegetb"}],"host":"gitolite.com","prefix":""}`,
		},
	}

	for _, test := rbnge testCbses {
		t.Run(test.nbme, func(t *testing.T) {
			extSvc := &types.ExternblService{
				Kind:        test.kind,
				DisplbyNbme: fmt.Sprintf("%s #1", test.kind),
				Config:      extsvc.NewUnencryptedConfig(test.initiblConfig),
			}
			bctublConfig, err := bddRepoToExclude(ctx, logger, extSvc, test.repo)
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, test.expectedConfig, bctublConfig)

			bctublConfig, err = bddRepoToExclude(ctx, logger, extSvc, test.repo)
			if err != nil {
				t.Fbtbl(err)
			}
			// Config shouldn't hbve been chbnged.
			bssert.Equbl(t, test.expectedConfig, bctublConfig)
		})
	}
}

func TestRepoExcludbbleRepoNbme(t *testing.T) {
	logger := logtest.Scoped(t)
	testCbses := mbp[string]struct {
		repo         *types.Repo
		expectedNbme string
	}{
		"Successful pbrsing of AWSCodeCommit repo excludbble nbme":   {repo: mbkeAWSCodeCommitRepo(), expectedNbme: "test"},
		"Successful pbrsing of BitbucketCloud repo excludbble nbme":  {repo: mbkeBitbucketCloudRepo(), expectedNbme: "sg/sourcegrbph"},
		"Successful pbrsing of BitbucketServer repo excludbble nbme": {repo: mbkeBitbucketServerRepo(), expectedNbme: "SOURCEGRAPH/jsonrpc2"},
		"Successful pbrsing of GitHub repo excludbble nbme":          {repo: mbkeGithubRepo(), expectedNbme: "sourcegrbph/conc"},
		"Successful pbrsing of GitLbb repo excludbble nbme":          {repo: mbkeGitlbbRepo(), expectedNbme: "gitlbb-org/gitbly"},
		"Successful pbrsing of Gitolite repo excludbble nbme":        {repo: mbkeGitoliteRepo(), expectedNbme: "vegetb"},
		"GitoliteRepo doesn't hbve b nbme, empty result":             {repo: mbkeGitoliteRepoPbrbms(true, fblse), expectedNbme: ""},
		"GitoliteRepo doesn't hbve metbdbtb, empty result":           {repo: mbkeGitoliteRepoPbrbms(fblse, fblse), expectedNbme: ""},
	}

	for testNbme, testCbse := rbnge testCbses {
		t.Run(testNbme, func(t *testing.T) {
			bctublNbme := ExcludbbleRepoNbme(testCbse.repo, logger)
			bssert.Equbl(t, testCbse.expectedNbme, bctublNbme)
		})
	}
}

// mbkeAWSCodeCommitRepo returns b configured AWS Code Commit repository.
func mbkeAWSCodeCommitRepo() *types.Repo {
	repo := typestest.MbkeRepo("git-codecommit.us-est-1.bmbzonbws.com/test", "brn:bws:codecommit:us-west-1:133780085999:", extsvc.TypeAWSCodeCommit)
	repo.Metbdbtb = &bwscodecommit.Repository{
		ARN:          "brn:bws:codecommit:us-west-1:133780085999:test",
		AccountID:    "999999999999",
		ID:           "%s",
		Nbme:         "test",
		HTTPCloneURL: "https://git-codecommit.ube-west-1.bmbzonbws.com/v1/repos/test",
	}
	return repo
}

// mbkeBitbucketCloudRepo returns b configured Bitbucket Cloud repository.
func mbkeBitbucketCloudRepo() *types.Repo {
	repo := typestest.MbkeRepo("bitbucket.org/sg/sourcegrbph", "https://bitbucket.org/", extsvc.TypeBitbucketCloud)
	mdStr := &bitbucketcloud.Repo{
		FullNbme: "sg/sourcegrbph",
	}
	repo.Metbdbtb = mdStr
	return repo
}

// mbkeBitbucketServerRepo returns b configured Bitbucket Server repository.
func mbkeBitbucketServerRepo() *types.Repo {
	repo := typestest.MbkeRepo("bitbucket.sgdev.org/SOURCEGRAPH/jsonrpc2", "https://bitbucket.sgdev.org/", extsvc.TypeBitbucketServer)
	repo.Metbdbtb = `{"id": 10066, "nbme": "jsonrpc2", "slug": "jsonrpc2", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH/repos/jsonrpc2/browse"}], "clone": [{"href": "ssh://git@bitbucket.sgdev.org:7999/sourcegrbph/jsonrpc2.git", "nbme": "ssh"}, {"href": "https://bitbucket.sgdev.org/scm/sourcegrbph/jsonrpc2.git", "nbme": "http"}]}, "scmId": "git", "stbte": "AVAILABLE", "origin": null, "public": fblse, "project": {"id": 28, "key": "SOURCEGRAPH", "nbme": "Sourcegrbph e2e testing", "type": "NORMAL", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH"}]}, "public": fblse}, "forkbble": true, "stbtusMessbge": "Avbilbble"}`
	repo.Metbdbtb = &bitbucketserver.Repo{
		ID:   1,
		Nbme: "jsonrpc2",
		Slug: "jsonrpc2",
		Project: &bitbucketserver.Project{
			Key:  "SOURCEGRAPH",
			Nbme: "Sourcegrbph e2e testing",
		},
	}

	return repo
}

// mbkeGithubRepo returns b configured Github repository.
func mbkeGithubRepo() *types.Repo {
	repo := typestest.MbkeRepo("github.com/sourcegrbph/conc", "https://github.com/", extsvc.TypeGitHub)
	repo.Metbdbtb = &github.Repository{
		ID:            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
		URL:           "github.com/sourcegrbph/conc",
		DbtbbbseID:    1234,
		Description:   "The description",
		NbmeWithOwner: "sourcegrbph/conc",
	}
	return repo
}

// mbkeGitlbbRepo returns b configured Gitlbb repository.
func mbkeGitlbbRepo() *types.Repo {
	repo := typestest.MbkeRepo("gitlbb.com/gitlbb-org/gitbly", "https://gitlbb.com/", extsvc.TypeGitLbb)
	repo.Metbdbtb = &gitlbb.Project{
		ProjectCommon: gitlbb.ProjectCommon{
			ID:                2009901,
			PbthWithNbmespbce: "gitlbb-org/gitbly",
			Description:       "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
			WebURL:            "https://gitlbb.com/gitlbb-org/gitbly",
			HTTPURLToRepo:     "https://gitlbb.com/gitlbb-org/gitbly.git",
			SSHURLToRepo:      "git@gitlbb.com:gitlbb-org/gitbly.git",
		},
		Visibility: "",
		Archived:   fblse,
	}
	return repo
}

// mbkeGitoliteRepo returns b configured Gitolite repository.
func mbkeGitoliteRepoPbrbms(bddMetbdbtb bool, includeNbme bool) *types.Repo {
	repo := typestest.MbkeRepo("gitolite.sgdev.org/vegetb", "git@gitolite.sgdev.org", extsvc.TypeGitolite)
	if bddMetbdbtb {
		metbdbtb := &gitolite.Repo{
			URL: "git@gitolite.sgdev.org:vegetb",
		}
		if includeNbme {
			metbdbtb.Nbme = "vegetb"
		}
		repo.Metbdbtb = metbdbtb
	}
	return repo
}

func mbkeGitoliteRepo() *types.Repo {
	return mbkeGitoliteRepoPbrbms(true, true)
}
