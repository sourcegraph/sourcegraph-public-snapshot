package backend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	internalrepos "github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
)

func TestAddRepoToExclude(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	testCases := []struct {
		name           string
		kind           string
		repo           *types.Repo
		initialConfig  string
		expectedConfig string
	}{
		{
			name:           "second attempt of excluding same repo is ignored for AWSCodeCommit schema",
			kind:           extsvc.KindAWSCodeCommit,
			repo:           makeAWSCodeCommitRepo(),
			initialConfig:  `{"accessKeyID":"accessKeyID","gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
			expectedConfig: `{"accessKeyID":"accessKeyID","exclude":[{"name":"test"}],"gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for Azure DevOps schema",
			kind:           extsvc.KindAzureDevOps,
			repo:           makeAzureDevOpsRepo(),
			initialConfig:  `{"url":"https://dev.azure.com","username":"test","token":"test","orgs":["org"]}`,
			expectedConfig: `{"exclude":[{"name":"org/namespace/test"}],"orgs":["org"],"token":"test","url":"https://dev.azure.com","username":"test"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketCloud schema",
			kind:           extsvc.KindBitbucketCloud,
			repo:           makeBitbucketCloudRepo(),
			initialConfig:  `{"appPassword":"","url":"https://bitbucket.org","username":""}`,
			expectedConfig: `{"exclude":[{"name":"sg/sourcegraph"}],"url":"https://bitbucket.org"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for BitbucketServer schema",
			kind:           extsvc.KindBitbucketServer,
			repo:           makeBitbucketServerRepo(),
			initialConfig:  `{"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
			expectedConfig: `{"exclude":[{"name":"SOURCEGRAPH/jsonrpc2"}],"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for Gerrit schema",
			kind:           extsvc.KindGerrit,
			repo:           makeGerritRepo(),
			initialConfig:  `{"url": "https://gerrit.example.com/", "username": "test", "password": "test", "projects": ["test"]}`,
			expectedConfig: `{"exclude":[{"name":"test"}],"password":"test","projects":["test"],"url":"https://gerrit.example.com/","username":"test"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitHub schema",
			kind:           extsvc.KindGitHub,
			repo:           makeGithubRepo(),
			initialConfig:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			expectedConfig: `{"exclude":[{"name":"sourcegraph/conc"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitLab schema",
			kind:           extsvc.KindGitLab,
			repo:           makeGitlabRepo(),
			initialConfig:  `{"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
			expectedConfig: `{"exclude":[{"name":"gitlab-org/gitaly"}],"projectQuery":null,"token":"abc","url":"https://gitlab.com"}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for Gitolite schema",
			kind:           extsvc.KindGitolite,
			repo:           makeGitoliteRepo(),
			initialConfig:  `{"host":"gitolite.com","prefix":""}`,
			expectedConfig: `{"exclude":[{"name":"vegeta"}],"host":"gitolite.com","prefix":""}`,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			extSvc := &types.ExternalService{
				Kind:        test.kind,
				DisplayName: fmt.Sprintf("%s #1", test.kind),
				Config:      extsvc.NewUnencryptedConfig(test.initialConfig),
			}
			actualConfig, err := addRepoToExclude(ctx, logger, extSvc, test.repo)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.expectedConfig, actualConfig)

			actualConfig, err = addRepoToExclude(ctx, logger, extSvc, test.repo)
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
		"Successful parsing of AWSCodeCommit repo excludable name":   {repo: makeAWSCodeCommitRepo(), expectedName: "test"},
		"Successful parsing of BitbucketCloud repo excludable name":  {repo: makeBitbucketCloudRepo(), expectedName: "sg/sourcegraph"},
		"Successful parsing of BitbucketServer repo excludable name": {repo: makeBitbucketServerRepo(), expectedName: "SOURCEGRAPH/jsonrpc2"},
		"Successful parsing of GitHub repo excludable name":          {repo: makeGithubRepo(), expectedName: "sourcegraph/conc"},
		"Successful parsing of GitLab repo excludable name":          {repo: makeGitlabRepo(), expectedName: "gitlab-org/gitaly"},
		"Successful parsing of Gitolite repo excludable name":        {repo: makeGitoliteRepo(), expectedName: "vegeta"},
		"GitoliteRepo doesn't have a name, empty result":             {repo: makeGitoliteRepoParams(true, false), expectedName: ""},
		"GitoliteRepo doesn't have metadata, empty result":           {repo: makeGitoliteRepoParams(false, false), expectedName: ""},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actualName := ExcludableRepoName(testCase.repo, logger)
			assert.Equal(t, testCase.expectedName, actualName)
		})
	}
}

// makeAWSCodeCommitRepo returns a configured AWS Code Commit repository.
func makeAWSCodeCommitRepo() *types.Repo {
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

// makeAzureDevOpsRepo returns a configured Azure DevOps repository.
func makeAzureDevOpsRepo() *types.Repo {
	repo := typestest.MakeRepo("dev.azure.com/org/namespace/test", "https://dev.azure.com", extsvc.TypeAzureDevOps)
	repo.Metadata = &azuredevops.Repository{
		APIURL: "https://azure.devops.com/org/namespace/test",
		Name:   "test",
		Project: azuredevops.Project{
			Name: "namespace",
		},
	}
	return repo
}

// makeBitbucketCloudRepo returns a configured Bitbucket Cloud repository.
func makeBitbucketCloudRepo() *types.Repo {
	repo := typestest.MakeRepo("bitbucket.org/sg/sourcegraph", "https://bitbucket.org/", extsvc.TypeBitbucketCloud)
	mdStr := &bitbucketcloud.Repo{
		FullName: "sg/sourcegraph",
	}
	repo.Metadata = mdStr
	return repo
}

// makeBitbucketServerRepo returns a configured Bitbucket Server repository.
func makeBitbucketServerRepo() *types.Repo {
	repo := typestest.MakeRepo("bitbucket.sgdev.org/SOURCEGRAPH/jsonrpc2", "https://bitbucket.sgdev.org/", extsvc.TypeBitbucketServer)
	repo.Metadata = `{"id": 10066, "name": "jsonrpc2", "slug": "jsonrpc2", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH/repos/jsonrpc2/browse"}], "clone": [{"href": "ssh://git@bitbucket.sgdev.org:7999/sourcegraph/jsonrpc2.git", "name": "ssh"}, {"href": "https://bitbucket.sgdev.org/scm/sourcegraph/jsonrpc2.git", "name": "http"}]}, "scmId": "git", "state": "AVAILABLE", "origin": null, "public": false, "project": {"id": 28, "key": "SOURCEGRAPH", "name": "Sourcegraph e2e testing", "type": "NORMAL", "links": {"self": [{"href": "https://bitbucket.sgdev.org/projects/SOURCEGRAPH"}]}, "public": false}, "forkable": true, "statusMessage": "Available"}`
	repo.Metadata = &bitbucketserver.Repo{
		ID:   1,
		Name: "jsonrpc2",
		Slug: "jsonrpc2",
		Project: &bitbucketserver.Project{
			Key:  "SOURCEGRAPH",
			Name: "Sourcegraph e2e testing",
		},
	}

	return repo
}

func makeGerritRepo() *types.Repo {
	repo := typestest.MakeRepo("gerrit.com/test", "https://gerrit.com/", extsvc.TypeGerrit)
	repo.Metadata = &gerrit.Project{
		ID: "test",
	}
	return repo
}

// makeGithubRepo returns a configured Github repository.
func makeGithubRepo() *types.Repo {
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

// makeGitlabRepo returns a configured Gitlab repository.
func makeGitlabRepo() *types.Repo {
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

// makeGitoliteRepo returns a configured Gitolite repository.
func makeGitoliteRepoParams(addMetadata bool, includeName bool) *types.Repo {
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

func makeGitoliteRepo() *types.Repo {
	return makeGitoliteRepoParams(true, true)
}

func TestExternalServiceValidate(t *testing.T) {
	var (
		src    internalrepos.Source
		called bool
		ctx    = context.Background()
	)
	src = testSource{
		fn: func() error {
			called = true
			return nil
		},
	}
	err := externalServiceValidate(ctx, src)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if !called {
		t.Errorf("expected called, got not called")
	}
}

func TestExternalService_ListNamespaces(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternalService{
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlabConnection := `
	{
	   "url": "https://gitlab.com",
	   "token": "abc",
	}`

	gitlabSource := types.ExternalService{
		Kind:   extsvc.KindGitLab,
		Config: extsvc.NewUnencryptedConfig(gitlabConnection),
	}

	githubOrg := &types.ExternalServiceNamespace{
		ID:         1,
		Name:       "sourcegraph",
		ExternalID: "aaaaa",
	}

	githubExternalServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternalService := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(githubExternalServiceConfig),
	}

	gitlabExternalServiceConfig := `
	{
		"url": "https://gitlab.com",
		"token": "abc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlabExternalService := types.ExternalService{
		ID:     2,
		Kind:   extsvc.KindGitLab,
		Config: extsvc.NewUnencryptedConfig(gitlabExternalServiceConfig),
	}

	gitlabRepository := &types.Repo{
		Name:        "gitlab.com/gitlab-org/gitaly",
		Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
		URI:         "gitlab.com/gitlab-org/gitaly",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
		Sources: map[string]*types.SourceInfo{
			gitlabSource.URN(): {
				ID:       gitlabSource.URN(),
				CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
			},
		},
	}

	var idDoesNotExist int64 = 99

	testCases := []struct {
		name              string
		externalService   *types.ExternalService
		externalServiceID *int64
		kind              string
		config            string
		result            []*types.ExternalServiceNamespace
		src               internalrepos.Source
		wantErr           string
	}{
		{
			name:   "discoverable source - github",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			result: []*types.ExternalServiceNamespace{githubOrg},
		},
		{
			name:    "unavailable - github.com",
			kind:    extsvc.KindGitHub,
			config:  githubConnection,
			src:     internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, &types.Repo{}), true, githubOrg),
			wantErr: "fake source unavailable",
		},
		{
			name:   "discoverable source - github - empty namespaces result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, &types.Repo{}), false),
			result: []*types.ExternalServiceNamespace{},
		},
		{
			name:    "source does not implement discoverable source",
			kind:    extsvc.KindGitLab,
			config:  gitlabConnection,
			src:     internalrepos.NewFakeSource(&gitlabSource, nil, &types.Repo{}),
			wantErr: internalrepos.UnimplementedDiscoverySource,
		},
		{
			name:              "discoverable source - github - use existing external service",
			externalService:   &githubExternalService,
			externalServiceID: &githubExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			result:            []*types.ExternalServiceNamespace{githubOrg},
		},
		{
			name:              "external service for ID does not exist and other config parameters are not attempted",
			externalService:   &githubExternalService,
			externalServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			wantErr:           fmt.Sprintf("external service not found: %d", idDoesNotExist),
		},
		{
			name:              "source does not implement discoverable source - use existing external service",
			externalService:   &gitlabExternalService,
			externalServiceID: &gitlabExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			src:               internalrepos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			wantErr:           internalrepos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			logger := logtest.Scoped(t)

			db := database.NewDB(logger, dbtest.NewDB(t))

			var store internalrepos.Store
			if tc.externalService != nil {
				store = internalrepos.NewStore(logtest.Scoped(t), db)
				if err := store.ExternalServiceStore().Upsert(ctx, tc.externalService); err != nil {
					t.Fatal(err)
				}
			}

			if tc.wantErr == "" {
				tc.wantErr = "<nil>"
			}

			e := NewMockExternalServices(logger, db, internalrepos.NewFakeSourcer(nil, tc.src))

			res, err := e.ListNamespaces(ctx, tc.externalServiceID, tc.kind, tc.config)
			if have, want := fmt.Sprint(err), tc.wantErr; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(res, tc.result); diff != "" {
				t.Fatalf("response mismatch(-have, +want): %s", diff)
			}
		})
	}
}

func TestExternalService_DiscoverRepos(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlabConnection := `
	{
	   "url": "https://gitlab.com",
	   "token": "abc",
	}`

	gitlabSource := types.ExternalService{
		ID:     2,
		Kind:   extsvc.KindGitLab,
		Config: extsvc.NewUnencryptedConfig(gitlabConnection),
	}

	githubRepository := &types.Repo{
		Name:        "github.com/foo/bar",
		Description: "The description",
		Archived:    false,
		Fork:        false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			githubSource.URN(): {
				ID:       githubSource.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	}

	gitlabRepository := &types.Repo{
		Name:        "gitlab.com/gitlab-org/gitaly",
		Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
		URI:         "gitlab.com/gitlab-org/gitaly",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
		Sources: map[string]*types.SourceInfo{
			gitlabSource.URN(): {
				ID:       gitlabSource.URN(),
				CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
			},
		},
	}

	githubExternalServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternalService := types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(githubExternalServiceConfig),
	}

	gitlabExternalServiceConfig := `
	{
		"url": "https://gitlab.com",
		"token": "abc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlabExternalService := types.ExternalService{
		ID:     2,
		Kind:   extsvc.KindGitLab,
		Config: extsvc.NewUnencryptedConfig(gitlabExternalServiceConfig),
	}

	var idDoesNotExist int64 = 99

	testCases := []struct {
		name              string
		externalService   *types.ExternalService
		externalServiceID *int64
		kind              string
		config            string
		query             string
		first             int32
		excludeRepos      []string
		result            []*types.ExternalServiceRepository
		src               internalrepos.Source
		wantErr           string
	}{
		{
			name:         "discoverable source - github",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{},
			src:          internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()},
		},
		{
			name:         "discoverable source - github - non empty query string",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "myquerystring",
			first:        5,
			excludeRepos: []string{},
			src:          internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()},
		},
		{
			name:         "discoverable source - github - non empty excludeRepos",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{"org1/repo1", "owner2/repo2"},
			src:          internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()},
		},
		{
			name:    "unavailable - github.com",
			kind:    extsvc.KindGitHub,
			config:  githubConnection,
			src:     internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil, githubRepository), true),
			wantErr: "fake source unavailable",
		},
		{
			name:   "discoverable source - github - empty repositories result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubSource, nil), false),
			result: []*types.ExternalServiceRepository{},
		},
		{
			name:    "source does not implement discoverable source",
			kind:    extsvc.KindGitLab,
			config:  gitlabConnection,
			src:     internalrepos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			wantErr: internalrepos.UnimplementedDiscoverySource,
		},
		{
			name:              "discoverable source - github - use existing external service",
			externalService:   &githubExternalService,
			externalServiceID: &githubExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubExternalService, nil, githubRepository), false),
			result:            []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()},
		},
		{
			name:              "external service for ID does not exist and other config parameters are not attempted",
			externalService:   &githubExternalService,
			externalServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubExternalServiceConfig,
			query:             "myquerystring",
			first:             5,
			excludeRepos:      []string{},
			src:               internalrepos.NewFakeDiscoverableSource(internalrepos.NewFakeSource(&githubExternalService, nil, githubRepository), false),
			wantErr:           fmt.Sprintf("external service not found: %d", idDoesNotExist),
		},
		{
			name:              "source does not implement discoverable source - use existing external service",
			externalService:   &gitlabExternalService,
			externalServiceID: &gitlabExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               internalrepos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			wantErr:           internalrepos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			logger := logtest.Scoped(t)

			db := database.NewDB(logger, dbtest.NewDB(t))

			var store internalrepos.Store
			if tc.externalService != nil {
				store = internalrepos.NewStore(logtest.Scoped(t), db)
				if err := store.ExternalServiceStore().Upsert(ctx, tc.externalService); err != nil {
					t.Fatal(err)
				}
			}

			e := NewMockExternalServices(logger, db, internalrepos.NewFakeSourcer(nil, tc.src))

			if tc.wantErr == "" {
				tc.wantErr = "<nil>"
			}

			res, err := e.DiscoverRepos(ctx, tc.externalServiceID, tc.kind, tc.config, tc.first, tc.query, tc.excludeRepos)
			if have, want := fmt.Sprint(err), tc.wantErr; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(res, tc.result); diff != "" {
				t.Fatalf("response mismatch(-have, +want): %s", diff)
			}
		})
	}
}

type testSource struct {
	fn func() error
}

var (
	_ internalrepos.Source     = &testSource{}
	_ internalrepos.UserSource = &testSource{}
)

func (t testSource) ListRepos(_ context.Context, _ chan internalrepos.SourceResult) {
}

func (t testSource) ExternalServices() types.ExternalServices {
	return nil
}

func (t testSource) CheckConnection(_ context.Context) error {
	return nil
}

func (t testSource) WithAuthenticator(_ auth.Authenticator) (internalrepos.Source, error) {
	return t, nil
}

func (t testSource) ValidateAuthenticator(_ context.Context) error {
	return t.fn()
}
