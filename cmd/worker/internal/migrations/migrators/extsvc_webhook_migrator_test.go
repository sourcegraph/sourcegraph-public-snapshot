package migrators

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalServiceWebhookMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	createExternalServices := func(t *testing.T, ctx context.Context, db database.DB) []*types.ExternalService {
		t.Helper()
		var svcs []*types.ExternalService

		es := db.ExternalServices()
		basestore := basestore.NewWithDB(db, sql.TxOptions{})

		// Create a trivial external service of each kind, as well as duplicate
		// services for the external service kinds that support webhooks.
		for _, svc := range []struct {
			kind string
			cfg  any
		}{
			{kind: extsvc.KindAWSCodeCommit, cfg: schema.AWSCodeCommitConnection{}},
			{kind: extsvc.KindBitbucketServer, cfg: schema.BitbucketServerConnection{}},
			{kind: extsvc.KindBitbucketCloud, cfg: schema.BitbucketCloudConnection{}},
			{kind: extsvc.KindGitHub, cfg: schema.GitHubConnection{}},
			{kind: extsvc.KindGitLab, cfg: schema.GitLabConnection{}},
			{kind: extsvc.KindGitolite, cfg: schema.GitoliteConnection{}},
			{kind: extsvc.KindPerforce, cfg: schema.PerforceConnection{}},
			{kind: extsvc.KindPhabricator, cfg: schema.PhabricatorConnection{}},
			{kind: extsvc.KindJVMPackages, cfg: schema.JVMPackagesConnection{}},
			{kind: extsvc.KindOther, cfg: schema.OtherExternalServiceConnection{}},

			{kind: extsvc.KindBitbucketServer, cfg: schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						DisableSync: false,
						Secret:      "this is a secret",
					},
				},
			}},
			{kind: extsvc.KindGitHub, cfg: schema.GitHubConnection{
				Webhooks: []*schema.GitHubWebhook{
					{
						Org:    "org",
						Secret: "this is also a secret",
					},
				},
			}},
			{kind: extsvc.KindGitLab, cfg: schema.GitLabConnection{
				Webhooks: []*schema.GitLabWebhook{
					{Secret: "this is yet another secret"},
				},
			}},
		} {
			buf, err := json.MarshalIndent(svc.cfg, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			svcs = append(svcs, &types.ExternalService{
				Kind:        svc.kind,
				DisplayName: svc.kind,
				Config:      string(buf),
			})
		}

		if err := es.Upsert(ctx, svcs...); err != nil {
			t.Fatal(err)
		}

		// Add one more external service with invalid JSON, which was once
		// possible and may still exist in databases in the wild.  We don't want
		// the migrator to error in that case!
		//
		// We'll have to do this the old fashioned way, since Create now
		// actually checks the validity of the configuration.
		row := basestore.QueryRow(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO
					external_services
					(
						kind,
						display_name,
						config,
						has_webhooks
					)
					VALUES (%s, %s, %s, %s)
				RETURNING
					id
				`,
				extsvc.KindOther,
				"other",
				"invalid JSON",
				false,
			),
		)
		var id int64
		if err := row.Scan(&id); err != nil {
			t.Fatal(err)
		}
		svc, err := es.GetByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		svcs = append(svcs, svc)

		// We'll also add another external service that is deleted, and
		// shouldn't count.
		if err := basestore.Exec(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO
					external_services
					(
						kind,
						display_name,
						config,
						deleted_at
					)
					VALUES (%s, %s, %s, NOW())
				`,
				extsvc.KindOther,
				"deleted",
				"{}",
			),
		); err != nil {
			t.Fatal(err)
		}

		return svcs
	}

	clearHasWebhooks := func(t *testing.T, ctx context.Context, db dbutil.DB) {
		t.Helper()

		basestore := basestore.NewWithDB(db, sql.TxOptions{})
		if err := basestore.Exec(
			ctx,
			sqlf.Sprintf("UPDATE external_services SET has_webhooks = NULL"),
		); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Progress", func(t *testing.T) {
		db := database.NewDB(dbtest.NewDB(t))
		createExternalServices(t, ctx, db)

		m := NewExternalServiceWebhookMigratorWithDB(db)

		// By default, all the external services should have non-NULL
		// has_webhooks.
		progress, err := m.Progress(ctx)
		assert.Nil(t, err)
		assert.EqualValues(t, 1., progress)

		// Now we'll clear that flag and ensure the progress drops to zero.
		clearHasWebhooks(t, ctx, db)
		progress, err = m.Progress(ctx)
		assert.Nil(t, err)
		assert.EqualValues(t, 0., progress)
	})

	t.Run("Up", func(t *testing.T) {
		db := database.NewDB(dbtest.NewDB(t))
		initSvcs := createExternalServices(t, ctx, db)
		es := db.ExternalServices()

		m := NewExternalServiceWebhookMigratorWithDB(db)
		// Ensure that we have to run two Ups.
		m.BatchSize = len(initSvcs) - 1

		// To start with, there should be nothing to do, as Upsert will have set
		// has_webhooks already. Let's make sure nothing happens successfully.
		assert.Nil(t, m.Up(ctx))

		// Now we'll clear out the has_webhooks flags and re-run Up. This should
		// update all but one of the external services.
		clearHasWebhooks(t, ctx, db)
		assert.Nil(t, m.Up(ctx))

		// Do we really have one external service left?
		after, err := es.List(ctx, database.ExternalServicesListOptions{
			NoCachedWebhooks: true,
		})
		assert.Nil(t, err)
		assert.EqualValues(t, 1, len(after))

		// Now we'll do the last one.
		assert.Nil(t, m.Up(ctx))
		after, err = es.List(ctx, database.ExternalServicesListOptions{
			NoCachedWebhooks: true,
		})
		assert.Nil(t, err)
		assert.EqualValues(t, 0, len(after))

		// Finally, let's make sure we have the expected number of each: we
		// should have three records with has_webhooks = true, and the rest
		// should be has_webhooks = false.
		svcs, err := es.List(ctx, database.ExternalServicesListOptions{})
		assert.Nil(t, err)

		hasWebhooks := 0
		noWebhooks := 0
		for _, svc := range svcs {
			assert.NotNil(t, svc.HasWebhooks)
			if *svc.HasWebhooks {
				hasWebhooks += 1
			} else {
				noWebhooks += 1
			}
		}

		assert.EqualValues(t, 3, hasWebhooks)
		assert.EqualValues(t, len(initSvcs)-3, noWebhooks)
	})
}
