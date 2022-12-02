package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoDirectories_InsertRoot(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})

	got, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "foo")
	if err != nil {
		t.Fatal(err)
	}
	want := &types.RepoDirectory{
		ID:           got.ID,
		RepoID:       repo.ID,
		AbsolutePath: "foo",
		ParentID:     0,
	}
	assert.Equal(t, want, got)
}

func TestRepoDirectories_InsertChild(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})

	parent, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "parent")
	if err != nil {
		t.Fatal(err)
	}
	if parent.ID == 0 {
		t.Fatal("want parent ID not 0")
	}
	got, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "parent/child")
	if err != nil {
		t.Fatal(err)
	}
	want := &types.RepoDirectory{
		ID:           got.ID,
		RepoID:       repo.ID,
		AbsolutePath: "parent/child",
		ParentID:     parent.ID,
	}
	assert.Equal(t, want, got)
}

func TestRepoDirectories_InsertDuplicate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	now := time.Now()

	service := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	err := db.ExternalServices().Create(ctx, confGet, &service)
	if err != nil {
		t.Fatal(err)
	}

	repo := mustCreate(ctx, t, db, &types.Repo{
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "r",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Name:        "name",
		Private:     true,
		URI:         "uri",
		Description: "description",
		Fork:        true,
		Archived:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    new(github.Repository),
		Sources: map[string]*types.SourceInfo{
			service.URN(): {
				ID:       service.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	})

	same1, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "same")
	if err != nil {
		t.Fatal(err)
	}
	same2, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "same")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, same1, same2)
}
