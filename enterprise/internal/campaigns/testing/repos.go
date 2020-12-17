package testing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	idb "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepo(t *testing.T, store *idb.ExternalServiceStore, serviceKind string) *types.Repo {
	t.Helper()

	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	svc := types.ExternalService{
		Kind:        serviceKind,
		DisplayName: serviceKind + " - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := store.Upsert(context.Background(), &svc); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	return &types.Repo{
		Name:    api.RepoName(fmt.Sprintf("repo-%d", svc.ID)),
		URI:     fmt.Sprintf("repo-%d", svc.ID),
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", svc.ID),
			ServiceType: extsvc.KindToType(serviceKind),
			ServiceID:   fmt.Sprintf("https://%s.com/", strings.ToLower(serviceKind)),
		},
		Sources: map[string]*types.SourceInfo{
			svc.URN(): {
				ID:       svc.URN(),
				CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}
}

func CreateTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*types.Repo, *types.ExternalService) {
	t.Helper()

	repoStore := idb.NewRepoStoreWithDB(db)
	esStore := idb.NewExternalServicesStoreWithDB(db)

	ext := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: MarshalJSON(t, &schema.GitHubConnection{
			Url:             "https://github.com",
			Token:           "SECRETTOKEN",
			RepositoryQuery: []string{"none"},
			// This field is needed to enforce permissions
			Authorization: &schema.GitHubAuthorization{},
		}),
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	if err := esStore.Create(ctx, confGet, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, esStore, extsvc.KindGitHub)
		r.Sources = map[string]*types.SourceInfo{ext.URN(): {
			ID:       ext.URN(),
			CloneURL: "https://secrettoken@github.com/" + string(r.Name),
		}}

		rs = append(rs, r)
	}

	err := repoStore.Create(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}

func CreateGitlabTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*types.Repo, *types.ExternalService) {
	t.Helper()

	repoStore := idb.NewRepoStoreWithDB(db)
	esStore := idb.NewExternalServicesStoreWithDB(db)

	ext := &types.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "GitLab",
		Config: MarshalJSON(t, &schema.GitLabConnection{
			Url:   "https://gitlab.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, esStore, extsvc.KindGitLab)
		r.Sources = map[string]*types.SourceInfo{ext.URN(): {
			ID:       ext.URN(),
			CloneURL: "https://git:gitlab-token@gitlab.com/" + string(r.Name),
		}}

		rs = append(rs, r)
	}

	err := repoStore.Create(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}

func CreateBbsTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*types.Repo, *types.ExternalService) {
	t.Helper()

	repoStore := idb.NewRepoStoreWithDB(db)
	esStore := idb.NewExternalServicesStoreWithDB(db)

	ext := &types.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server",
		Config: MarshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bitbucket.sourcegraph.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := esStore.Upsert(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*types.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, esStore, extsvc.KindBitbucketServer)
		r.Sources = map[string]*types.SourceInfo{
			ext.URN(): {
				ID:       ext.URN(),
				CloneURL: "https://bbs-user:bbs-token@bitbucket.sourcegraph.com/scm/" + string(r.Name),
			},
		}

		rs = append(rs, r)
	}

	err := repoStore.Create(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}
