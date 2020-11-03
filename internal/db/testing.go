package db

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

func assertJSONEqual(t *testing.T, want, got interface{}) {
	want_j := asJSON(t, want)
	got_j := asJSON(t, got)
	if want_j != got_j {
		t.Errorf("Wanted %s, but got %s", want_j, got_j)
	}
}

func jsonEqual(t *testing.T, a, b interface{}) bool {
	return asJSON(t, a) == asJSON(t, b)
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// MakeGithubRepo returns a configured Github repository.
func MakeGithubRepo(services ...*types.ExternalService) *types.Repo {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	repo := types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "http://github.com",
		},
		Name: "github.com/foo/bar",
		RepoFields: &types.RepoFields{
			URI:         "github.com/foo/bar",
			Description: "The description",
			CreatedAt:   now,
			Sources:     make(map[string]*types.SourceInfo),
			Metadata:    new(github.Repository),
		},
	}

	for _, svc := range services {
		repo.Sources[svc.URN()] = &types.SourceInfo{
			ID: svc.URN(),
		}
	}

	return &repo
}

// MakeGithubRepo returns a configured Gitlab repository.
func MakeGitlabRepo(services ...*types.ExternalService) *types.Repo {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	repo := types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAA==",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "http://gitlab.com",
		},
		Name: "gitlab.com/foo/bar",
		RepoFields: &types.RepoFields{
			URI:         "gitlab.com/foo/bar",
			Description: "The description",
			CreatedAt:   now,
			Sources:     make(map[string]*types.SourceInfo),
			Metadata:    new(gitlab.Project),
		},
	}

	for _, svc := range services {
		repo.Sources[svc.URN()] = &types.SourceInfo{
			ID: svc.URN(),
		}
	}

	return &repo
}

// GenerateRepos takes a list of base repos and generates n ones with different names.
func GenerateRepos(n int, base ...*types.Repo) types.Repos {
	if len(base) == 0 {
		return nil
	}

	rs := make(types.Repos, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base[i%len(base)].Clone()
		r.Name += api.RepoName(id)
		r.ExternalRepo.ID += id
		rs = append(rs, r)
	}
	return rs
}

// MakeExternalServices creates one configured external service per kind and returns the list.
func MakeExternalServices() types.ExternalServices {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	githubSvc := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"]}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gitlabSvc := types.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "GitLab - Test",
		Config:      `{"url": "https://gitlab.com", "token": "abc", "projectQuery": ["projects?membership=true&archived=no"]}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	bitbucketServerSvc := types.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server - Test",
		Config:      `{"url": "https://bitbucket.com", "username": "foo", "token": "abc", "repositoryQuery": ["none"]}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	bitbucketCloudSvc := types.ExternalService{
		Kind:        extsvc.KindBitbucketCloud,
		DisplayName: "Bitbucket Cloud - Test",
		Config:      `{"url": "https://bitbucket.com", "username": "foo", "appPassword": "abc"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	awsSvc := types.ExternalService{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplayName: "AWS Code - Test",
		Config:      `{"region": "eu-west-1", "accessKeyID": "key", "secretAccessKey": "secret", "gitCredentials": {"username": "foo", "password": "bar"}}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	otherSvc := types.ExternalService{
		Kind:        extsvc.KindOther,
		DisplayName: "Other - Test",
		Config:      `{"url": "https://other.com", "repos": ["none"]}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	gitoliteSvc := types.ExternalService{
		Kind:        extsvc.KindGitolite,
		DisplayName: "Gitolite - Test",
		Config:      `{"prefix": "foo", "host": "bar"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return []*types.ExternalService{
		&githubSvc,
		&gitlabSvc,
		&bitbucketServerSvc,
		&bitbucketCloudSvc,
		&awsSvc,
		&otherSvc,
		&gitoliteSvc,
	}
}

// GenerateExternalServices takes a list of base external services and generates n ones with different names.
func GenerateExternalServices(n int, base ...*types.ExternalService) types.ExternalServices {
	if len(base) == 0 {
		return nil
	}
	es := make(types.ExternalServices, 0, n)
	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		r := base[i%len(base)].Clone()
		r.DisplayName += id
		es = append(es, r)
	}
	return es
}
