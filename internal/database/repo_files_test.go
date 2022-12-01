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

func TestRepoFiles_Insert(t *testing.T) {
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
	vv := types.RepoVersion{
		RepoID:     repo.ID,
		ExternalID: "pretend this is a git sha",
		PathCoverage: types.RepoVersionPathCoverage{
			PathColor: 1,
			PathIndex: 1,
		},
		Reachability: map[int]int{1: 1},
	}

	v, err := db.RepoVersions().CreateIfNotExists(ctx, vv)
	if err != nil {
		t.Fatal(err)
	}

	d, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "dir")
	if err != nil {
		t.Fatal(err)
	}

	cID, err := db.RepoFileContents().Create(ctx, "content")
	if err != nil {
		t.Fatal(err)
	}

	ff := types.RepoFile{
		DirectoryID:      d.ID,
		VersionID:        v.ID,
		TopologicalOrder: 1, // we need to compute this
		BaseName:         "file",
		ContentID:        cID,
	}
	_, err = db.RepoFiles().CreateIfNotExists(ctx, ff)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRepoFiles_Conflict(t *testing.T) {
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
	vv := types.RepoVersion{
		RepoID:     repo.ID,
		ExternalID: "pretend this is a git sha",
		PathCoverage: types.RepoVersionPathCoverage{
			PathColor: 1,
			PathIndex: 1,
		},
		Reachability: map[int]int{1: 1},
	}

	v, err := db.RepoVersions().CreateIfNotExists(ctx, vv)
	if err != nil {
		t.Fatal(err)
	}

	d, err := db.RepoDirectories().CreateIfNotExists(ctx, repo.ID, "dir2")
	if err != nil {
		t.Fatal(err)
	}

	cID, err := db.RepoFileContents().Create(ctx, "content")
	if err != nil {
		t.Fatal(err)
	}

	ff := types.RepoFile{
		DirectoryID:      d.ID,
		VersionID:        v.ID,
		TopologicalOrder: 1, // we need to compute this
		BaseName:         "file2",
		ContentID:        cID,
	}
	f1, err := db.RepoFiles().CreateIfNotExists(ctx, ff)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, 0, f1.ID)
	f2, err := db.RepoFiles().CreateIfNotExists(ctx, ff)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, f1.ID, f2.ID)
}
