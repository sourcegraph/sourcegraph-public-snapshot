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
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
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

func MakeRepo(name, serviceID, serviceType string, services ...*types.ExternalService) *types.Repo {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	repo := types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "1234",
			ServiceType: serviceType,
			ServiceID:   serviceID,
		},
		Name: api.RepoName(name),
		RepoFields: &types.RepoFields{
			URI:         name,
			Description: "The description",
			CreatedAt:   now,
			Sources:     make(map[string]*types.SourceInfo),
		},
	}

	for _, svc := range services {
		repo.Sources[svc.URN()] = &types.SourceInfo{
			ID: svc.URN(),
		}
	}

	return &repo
}

// MakeGithubRepo returns a configured Github repository.
func MakeGithubRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("github.com/foo/bar", "http://github.com", extsvc.TypeGitHub, services...)
	repo.RepoFields.Metadata = new(github.Repository)
	return repo
}

// MakeGitlabRepo returns a configured Gitlab repository.
func MakeGitlabRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("gitlab.com/foo/bar", "http://gitlab.com", extsvc.TypeGitLab, services...)
	repo.RepoFields.Metadata = new(gitlab.Project)
	return repo
}

// MakeBitbucketServerRepo returns a configured Bitbucket Server repository.
func MakeBitbucketServerRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("bitbucketserver.mycorp.com/foo/bar", "http://bitbucketserver.mycorp.com", extsvc.TypeBitbucketServer, services...)
	repo.RepoFields.Metadata = new(bitbucketserver.Repo)
	return repo
}

// MakeAWSCodeCommitRepo returns a configured AWS Code Commit repository.
func MakeAWSCodeCommitRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("git-codecommit.us-west-1.amazonaws.com/stripe-go", "arn:aws:codecommit:us-west-1:999999999999:", extsvc.KindAWSCodeCommit, services...)
	repo.RepoFields.Metadata = new(awscodecommit.Repository)
	return repo
}

// MakeOtherRepo returns a configured repository from a custom host.
func MakeOtherRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("git-host.com/org/foo", "https://git-host.com/", extsvc.KindOther, services...)
	return repo
}

// MakeGitoliteRepo returns a configured Gitolite repository.
func MakeGitoliteRepo(services ...*types.ExternalService) *types.Repo {
	repo := MakeRepo("gitolite.mycorp.com/bar", "git@gitolite.mycorp.com", extsvc.KindGitolite, services...)
	repo.RepoFields.Metadata = new(gitolite.Repo)
	return repo
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

// ExternalServicesToMap is a helper function that returns a map whose key is the external service kind.
// If two external services have the same kind, only the last one will be stored in the map.
func ExternalServicesToMap(es types.ExternalServices) map[string]*types.ExternalService {
	m := make(map[string]*types.ExternalService)

	for _, svc := range es {
		m[svc.Kind] = svc
	}

	return m
}
