package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalServiceWebhookMigrator(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	var testExtSvcs = []struct {
		kind        string
		hasWebhooks bool
		cfg         any
	}{
		{kind: "AWSCODECOMMIT", cfg: schema.AWSCodeCommitConnection{}},
		{kind: "BITBUCKETSERVER", cfg: schema.BitbucketServerConnection{}},
		{kind: "BITBUCKETCLOUD", cfg: schema.BitbucketCloudConnection{}},
		{kind: "GITHUB", cfg: schema.GitHubConnection{}},
		{kind: "GITLAB", cfg: schema.GitLabConnection{}},
		{kind: "GITOLITE", cfg: schema.GitoliteConnection{}},
		{kind: "PERFORCE", cfg: schema.PerforceConnection{}},
		{kind: "PHABRICATOR", cfg: schema.PhabricatorConnection{}},
		{kind: "JVMPACKAGES", cfg: schema.JVMPackagesConnection{}},
		{kind: "OTHER", cfg: schema.OtherExternalServiceConnection{}},
		{kind: "BITBUCKETSERVER", hasWebhooks: true, cfg: schema.BitbucketServerConnection{
			Plugin: &schema.BitbucketServerPlugin{
				Webhooks: &schema.BitbucketServerPluginWebhooks{
					DisableSync: false,
					Secret:      "this is a secret",
				},
			},
		}},
		{kind: "GITHUB", hasWebhooks: true, cfg: schema.GitHubConnection{
			Webhooks: []*schema.GitHubWebhook{
				{
					Org:    "org",
					Secret: "this is also a secret",
				},
			},
		}},
		{kind: "GITLAB", hasWebhooks: true, cfg: schema.GitLabConnection{
			Webhooks: []*schema.GitLabWebhook{
				{Secret: "this is yet another secret"},
			},
		}},
	}

	createExternalServices := func(t *testing.T, ctx context.Context, store *basestore.Store) {
		t.Helper()

		// Create a trivial external service of each kind, as well as duplicate
		// services for the external service kinds that support webhooks.
		for _, svc := range testExtSvcs {
			buf, err := json.MarshalIndent(svc.cfg, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			if err := store.Exec(ctx, sqlf.Sprintf(`
				INSERT INTO external_services (kind, display_name, config, created_at, has_webhooks)
				VALUES (%s, %s, %s, NOW(), %s)
			`,
				svc.kind,
				svc.kind,
				string(buf),
				svc.hasWebhooks,
			)); err != nil {
				t.Fatal(err)
			}
		}

		// Add one more external service with invalid JSON, which was once
		// possible and may still exist in databases in the wild.  We don't want
		// the migrator to error in that case!
		//
		// We'll have to do this the old fashioned way, since Create now
		// actually checks the validity of the configuration.
		if err := store.Exec(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO external_services (kind, display_name, config, created_at, has_webhooks)
				VALUES (%s, %s, %s, NOW(), %s)
			`,
				"OTHER",
				"other",
				"invalid JSON",
				false,
			),
		); err != nil {
			t.Fatal(err)
		}

		// We'll also add another external service that is deleted, and shouldn't count.
		if err := store.Exec(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO external_services (kind, display_name, config, deleted_at)
				VALUES (%s, %s, %s, NOW())
			`,
				"OTHER",
				"deleted",
				"{}",
			),
		); err != nil {
			t.Fatal(err)
		}
	}

	clearHasWebhooks := func(t *testing.T, ctx context.Context, store *basestore.Store) {
		t.Helper()

		if err := store.Exec(
			ctx,
			sqlf.Sprintf("UPDATE external_services SET has_webhooks = NULL"),
		); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Progress", func(t *testing.T) {
		db := database.NewDB(logger, dbtest.NewDB(logger, t))
		store := basestore.NewWithHandle(db.Handle())
		createExternalServices(t, ctx, store)

		key := et.TestKey{}
		m := NewExternalServiceWebhookMigratorWithDB(store, key, 50)

		// By default, all the external services should have non-NULL
		// has_webhooks.
		progress, err := m.Progress(ctx, false)
		assert.Nil(t, err)
		assert.EqualValues(t, 1., progress)

		// Now we'll clear that flag and ensure the progress drops to zero.
		clearHasWebhooks(t, ctx, store)
		progress, err = m.Progress(ctx, true)
		assert.Nil(t, err)
		assert.EqualValues(t, 0., progress)
	})

	t.Run("Up", func(t *testing.T) {
		db := database.NewDB(logger, dbtest.NewDB(logger, t))
		store := basestore.NewWithHandle(db.Handle())
		createExternalServices(t, ctx, store)
		// Count the invalid JSON, not the deleted one
		numInitSvcs := len(testExtSvcs) + 1

		key := et.TestKey{}
		m := NewExternalServiceWebhookMigratorWithDB(store, key, 50)
		// Ensure that we have to run two Ups.
		m.batchSize = numInitSvcs - 1

		// To start with, there should be nothing to do, as Upsert will have set
		// has_webhooks already. Let's make sure nothing happens successfully.
		assert.Nil(t, m.Up(ctx))

		// Now we'll clear out the has_webhooks flags and re-run Up. This should
		// update all but one of the external services.
		clearHasWebhooks(t, ctx, store)
		assert.Nil(t, m.Up(ctx))

		// Do we really have one external service left?
		numWebhooksNull, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NULL`)))
		assert.Nil(t, err)
		assert.Equal(t, 1, numWebhooksNull)

		// Now we'll do the last one.
		assert.Nil(t, m.Up(ctx))
		numWebhooksNull, _, err = basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NULL`)))
		assert.Nil(t, err)
		assert.Equal(t, 0, numWebhooksNull)

		// Finally, let's make sure we have the expected number of each: we
		// should have three records with has_webhooks = true, and the rest
		// should be has_webhooks = false.
		numWebhooksTrue, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS TRUE`)))
		assert.Nil(t, err)
		assert.EqualValues(t, 3, numWebhooksTrue)

		numWebhooksFalse, _, err := basestore.ScanFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS FALSE`)))
		assert.Nil(t, err)
		assert.EqualValues(t, numInitSvcs-3, numWebhooksFalse)
	})
}
