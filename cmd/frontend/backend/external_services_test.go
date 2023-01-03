package backend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestTrimHostFromRepoName(t *testing.T) {
	testCases := map[string]struct {
		inputRepoName       string
		expectedTrimmedName string
	}{
		"repo has a hostname":               {inputRepoName: "github.com/sourcegraph/sourcegraph", expectedTrimmedName: "sourcegraph/sourcegraph"},
		"repo doesn't have a hostname":      {inputRepoName: "sourcegraph/horsegraph", expectedTrimmedName: "sourcegraph/horsegraph"},
		"non-hostname prefix isn't trimmed": {inputRepoName: "source/graph/horsegraph", expectedTrimmedName: "source/graph/horsegraph"},
		"GitLab nested project is trimmed":  {inputRepoName: "gitlab.com/source/graph/horsegraph", expectedTrimmedName: "source/graph/horsegraph"},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			actualTrimmedName := trimHostFromRepoName(testCase.inputRepoName)
			assert.Equal(t, testCase.expectedTrimmedName, actualTrimmedName)
		})
	}
}

func TestAddRepoToExclude(t *testing.T) {
	ctx := context.Background()

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
			expectedConfig: `{"accessKeyID":"accessKeyID","exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"gitCredentials":{"password":"","username":""},"region":"","secretAccessKey":""}`,
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
			expectedConfig: `{"exclude":[{"id":1,"name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://bitbucket.sg.org","username":""}`,
		},
		{
			name:           "second attempt of excluding same repo is ignored for GitHub schema",
			kind:           extsvc.KindGitHub,
			initialConfig:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			expectedConfig: `{"exclude":[{"id":"1","name":"sourcegraph/sourcegraph"}],"repositoryQuery":["none"],"token":"abc","url":"https://github.com"}`,
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
			actualConfig, err := addRepoToExclude(ctx, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.expectedConfig, actualConfig)

			actualConfig, err = addRepoToExclude(ctx, extSvc, &types.Repo{ID: api.RepoID(1), Name: "sourcegraph/sourcegraph"})
			if err != nil {
				t.Fatal(err)
			}
			// Config shouldn't have been changed.
			assert.Equal(t, test.expectedConfig, actualConfig)
		})
	}
}
