package backend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/stretchr/testify/assert"
)

func TestAddRepoToExclude(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	testCases := []struct {
		name           string
		kind           string
		initialConfig  string
		expectedConfig string
	}{
		{
			name:           "second attempt of excluding same repo is ignored for AWSCodeCommit schema",
			kind:           extsvc.KindAWSCodeCommit,
			initialConfig:  `{"accessKeyID":"accessKeyID","gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
			expectedConfig: `{"accessKeyID":"accessKeyID","exclude":[{"name":"sourcegraph/sourcegraph"}],"gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketCloud schema",
			kind:           extsvc.KindBitbucketCloud,
			initialConfig:  `{"appPassword":"","url":"https://bitbucket.org","username":""}`,
			expectedConfig: `{"appPassword":"","exclude":[{"name":"sourcegraph/sourcegraph"}],"url":"https://bitbucket.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketServer schema",
			kind:           extsvc.KindBitbucketServer,
			initialConfig:  `{"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitHub schema",
			kind:           extsvc.KindGitHub,
			initialConfig:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitLab schema",
			kind:           extsvc.KindGitLab,
			initialConfig:  `{"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for Gitolite schema",
			kind:           extsvc.KindGitolite,
			initialConfig:  `{"host":"gitolite.com","prefix":""}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/sourcegraph"}],"host":"gitolite.com","prefix":""}`,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			extSvc := &types.ExternalService{
				Kind:        test.kind,
				DisplayName: fmt.Sprintf("%s #1", test.kind),
				Config:      extsvc.NewUnencryptedConfig(test.initialConfig),
			}
			actualConfig, err := addRepoToExclude(ctx, logger, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.expectedConfig, actualConfig)

			actualConfig, err = addRepoToExclude(ctx, logger, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			// Config shouldn't have been changed.
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}

func TestRepoExcludableRepoName(t *testing.T) {
	logger := logtest.Scoped(t)
	testCases := map[string]struct {
		repo         *types.Repo
		expectedName string
	}{
		"Successful parsing of AWSCodeCommit repo excludable name":   {repo: MakeAWSCodeCommitRepo(), expectedName: "test"},
		"Successful parsing of BitbucketCloud repo excludable name":  {repo: MakeBitbucketCloudRepo(), expectedName: "sg/sourcegraph"},
		"Successful parsing of BitbucketServer repo excludable name": {repo: MakeBitbucketServerRepo(), expectedName: "SOURCEGRAPH/jsonrpc2"},
		"Successful parsing of GitHub repo excludable name":          {repo: MakeGithubRepo(), expectedName: "sourcegraph/conc"},
		"Successful parsing of GitLab repo excludable name":          {repo: MakeGitlabRepo(), expectedName: "gitlab-org/gitaly"},
		"Successful parsing of Gitolite repo excludable name":        {repo: MakeGitoliteRepo(true, true), expectedName: "vegeta"},
		"GitoliteRepo doesn't have a name, empty result":             {repo: MakeGitoliteRepo(true, false), expectedName: ""},
		"GitoliteRepo doesn't have metadata, empty result":           {repo: MakeGitoliteRepo(false, false), expectedName: ""},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actualName := ExcludableRepoName(testCase.repo, logger)
			assert.Equal(t, testCase.expectedName, actualName)
		})
	}
}

// MakeAWSCodeCommitRepo returns a configured AWS Code Commit repository.
func MakeAWSCodeCommitRepo() *types.Repo {
	repo := typestest.MakeRepo("git-codecommit.us-est-1.amazonaws.com/test", "arn:aws:codecommit:us-west-1:133780085999:", extsvc.TypeAWSCodeCommit)
	repo.Metadata = &awscodecommit.Repository{
		ARN:          "arn:aws:codecommit:us-west-1:133780085999:test",
		AccountID:    "999999999999",
		ID:           "%s",
		Name:         "test",
		HTTPCloneURL: "https://git-codecommit.uae-west-1.amazonaws.com/v1/repos/test",
	}
	return repo
}

// MakeBitbucketCloudRepo returns a configured Bitbucket Cloud repository.
func MakeBitbucketCloudRepo() *types.Repo {
	repo := typestest.MakeRepo("bitbucket.org/sg/sourcegraph", "https://bitbucket.org/", extsvc.TypeBitbucketCloud)
	mdStr := &bitbucketcloud.Repo{
		FullName: "sg/sourcegraph",
	}
	repo.Metadata = mdStr
	return repo
}

// MakeBitbucketServerRepo returns a configured Bitbucket Server repository.
func MakeBitbucketServerRepo() *types.Repo {
	repo := typestest.MakeRepo("bitbucket.sgdev.org/SOURCEGRAPH/jsonrpc2", "https://bitbucket.sgdev.org/", extsvc.TypeBitbucketServer)
	repo.Metadata = `{"id": 10066, "name": "jsonrpc2", "slug": "jsonrpc2", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH/repos/jsonrpc2/browse"}], "clone": [{"href": "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/jsonrpc2.git", "name": "ssh"}, {"href": "https://bitbucket.sgdev.org/scm/sourcegraph/jsonrpc2.git", "name": "http"}]}, "scmId": "git", "state": "AVAILABLE", "origin": null, "public": false, "project": {"id": 28, "key": "SOURCEGRAPH", "name": "Sourcegraph e2e testing", "type": "NORMAL", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH"}]}, "public": false}, "forkable": true, "statusMessage": "Available"}`
	repo.Metadata = &bitbucketserver.Repo{
		ID:   1,
		Name: "jsonrpc2",
		Slug: "jsonrpc2",
		Project: &bitbucketserver.Project{
			Key:  "SOURCEGRAPH",
			Name: "SOURCEGRAPH",
		},
	}

	return repo
}

// MakeGithubRepo returns a configured Github repository.
func MakeGithubRepo() *types.Repo {
	repo := typestest.MakeRepo("github.com/sourcegraph/conc", "https://github.com/", extsvc.TypeGitHub)
	repo.Metadata = &github.Repository{
		ID:            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
		URL:           "github.com/sourcegraph/conc",
		DatabaseID:    1234,
		Description:   "The description",
		NameWithOwner: "sourcegraph/conc",
	}
	return repo
}

// MakeGitlabRepo returns a configured Gitlab repository.
func MakeGitlabRepo() *types.Repo {
	repo := typestest.MakeRepo("gitlab.com/gitlab-org/gitaly", "https://gitlab.com/", extsvc.TypeGitLab)
	repo.Metadata = &gitlab.Project{
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
	}
	return repo
}

// MakeGitoliteRepo returns a configured Gitolite repository.
func MakeGitoliteRepo(addMetadata bool, includeName bool) *types.Repo {
	repo := typestest.MakeRepo("gitolite.sgdev.org/vegeta", "git@gitolite.sgdev.org", extsvc.TypeGitolite)
	if addMetadata {
		metadata := &gitolite.Repo{
			URL: "git@gitolite.sgdev.org:vegeta",
		}
		if includeName {
			metadata.Name = "vegeta"
		}
		repo.Metadata = metadata
	}
	return repo
}
