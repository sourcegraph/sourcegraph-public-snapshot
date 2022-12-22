package database

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

func TestCodeowners_putAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())
	r := makeRepo(ctx, t, db)
	codeowners := db.Codeowners()
	want := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "/internal/database/*.go",
				Owner:   []*codeownerspb.Owner{{Handle: "sourcegraphers"}},
			},
		},
	}
	err := codeowners.PutHead(ctx, r.Name, want)
	require.NoError(t, err)
	got, err := codeowners.GetHead(ctx, r.Name)
	require.NoError(t, err)
	assert.Equal(t, want.Repr(), got.Repr())
}

func TestCodeowners_getNoData(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())
	r := makeRepo(ctx, t, db)
	codeowners := db.Codeowners()
	got, err := codeowners.GetHead(ctx, r.Name)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestCodeowners_overwrite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())
	r := makeRepo(ctx, t, db)
	codeowners := db.Codeowners()
	first := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "/internal/database/*.go",
				Owner:   []*codeownerspb.Owner{{Handle: "sourcegraphers"}},
			},
		},
	}
	err := codeowners.PutHead(ctx, r.Name, first)
	require.NoError(t, err)
	second := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "/internal/database/*.go",
				Owner: []*codeownerspb.Owner{
					{Handle: "sourcegraphers"},
					{Handle: "another-owner"},
				},
			},
		},
	}
	err = codeowners.PutHead(ctx, r.Name, second)
	require.NoError(t, err)
	got, err := codeowners.GetHead(ctx, r.Name)
	require.NoError(t, err)
	assert.Equal(t, second.Repr(), got.Repr())
}

func TestCodeowners_noRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())
	codeowners := db.Codeowners()
	err := codeowners.PutHead(ctx, "I just made up this repo name", &codeownerspb.File{})
	require.Error(t, err)
}

func makeRepo(ctx context.Context, t *testing.T, db DB) *types.Repo {
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

	return mustCreate(ctx, t, db, &types.Repo{
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
}
