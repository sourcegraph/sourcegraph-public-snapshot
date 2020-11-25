package testing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRepo(t *testing.T, store repos.Store, serviceKind string) *repos.Repo {
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

	if err := store.UpsertExternalServices(context.Background(), &svc); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	return &repos.Repo{
		Name:    fmt.Sprintf("repo-%d", svc.ID),
		URI:     fmt.Sprintf("repo-%d", svc.ID),
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          fmt.Sprintf("external-id-%d", svc.ID),
			ServiceType: extsvc.KindToType(serviceKind),
			ServiceID:   fmt.Sprintf("https://%s.com/", strings.ToLower(serviceKind)),
		},
		Sources: map[string]*repos.SourceInfo{
			svc.URN(): {
				ID:       svc.URN(),
				CloneURL: "https://secrettoken@github.com/sourcegraph/sourcegraph",
			},
		},
	}
}

func CreateTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*repos.Repo, *types.ExternalService) {
	t.Helper()

	rstore := repos.NewDBStore(db, sql.TxOptions{})

	ext := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: MarshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := rstore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, rstore, extsvc.KindGitHub)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {
			ID:       ext.URN(),
			CloneURL: "https://secrettoken@github.com/" + r.Name,
		}}

		rs = append(rs, r)
	}

	err := rstore.InsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}

func CreateGitlabTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*repos.Repo, *types.ExternalService) {
	t.Helper()

	rstore := repos.NewDBStore(db, sql.TxOptions{})

	ext := &types.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "GitLab",
		Config: MarshalJSON(t, &schema.GitLabConnection{
			Url:   "https://gitlab.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := rstore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, rstore, extsvc.KindGitLab)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {
			ID:       ext.URN(),
			CloneURL: "https://git:gitlab-token@gitlab.com/" + r.Name,
		}}

		rs = append(rs, r)
	}

	err := rstore.InsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}

func CreateBbsTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*repos.Repo, *types.ExternalService) {
	t.Helper()

	rstore := repos.NewDBStore(db, sql.TxOptions{})

	ext := &types.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket Server",
		Config: MarshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bitbucket.sourcegraph.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := rstore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < count; i++ {
		r := TestRepo(t, rstore, extsvc.KindBitbucketServer)
		r.Sources = map[string]*repos.SourceInfo{
			ext.URN(): {
				ID:       ext.URN(),
				CloneURL: "https://bbs-user:bbs-token@bitbucket.sourcegraph.com/scm/" + r.Name,
			},
		}

		rs = append(rs, r)
	}

	err := rstore.InsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}
