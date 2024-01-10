package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCreation(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	p, err := db.Phabricator().Create(ctx, "callsign", "repo", "url")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &types.PhabricatorRepo{
		ID:       1,
		Name:     "repo",
		URL:      "url",
		Callsign: "callsign",
	}, p)

	p, err = db.Phabricator().CreateOrUpdate(ctx, "callsign2", "repo", "url2")
	if err != nil {
		t.Fatal(err)
	}
	// Assert the ID is still the same
	assert.Equal(t, &types.PhabricatorRepo{
		ID:       1,
		Name:     "repo",
		URL:      "url2",
		Callsign: "callsign2",
	}, p)
}

func TestCreateIfNotExistsAndGetByName(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	config := &schema.PhabricatorConnection{
		Repos: []*schema.Repos{
			{
				Callsign: "callsign",
				Path:     "repo",
			},
		},
		Token: "deadbeef",
		Url:   "url",
	}
	marshalled, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.ExternalServices().Create(ctx, func() *conf.Unified {
		return &conf.Unified{}
	}, &types.ExternalService{
		ID:          0,
		Kind:        extsvc.KindPhabricator,
		DisplayName: "Phab",
		Config:      extsvc.NewUnencryptedConfig(string(marshalled)),
	}); err != nil {
		t.Fatal(err)
	}

	_, err = db.Phabricator().CreateIfNotExists(ctx, "callsign", "repo", "url")
	if err != nil {
		t.Fatal(err)
	}

	// It should exist
	repo, err := db.Phabricator().GetByName(ctx, "repo")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, repo)
}
