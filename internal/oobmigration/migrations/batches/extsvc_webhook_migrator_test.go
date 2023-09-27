pbckbge bbtches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestExternblServiceWebhookMigrbtor(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Bbckground()

	vbr testExtSvcs = []struct {
		kind        string
		hbsWebhooks bool
		cfg         bny
	}{
		{kind: "AWSCODECOMMIT", cfg: schemb.AWSCodeCommitConnection{}},
		{kind: "BITBUCKETSERVER", cfg: schemb.BitbucketServerConnection{}},
		{kind: "BITBUCKETCLOUD", cfg: schemb.BitbucketCloudConnection{}},
		{kind: "GITHUB", cfg: schemb.GitHubConnection{}},
		{kind: "GITLAB", cfg: schemb.GitLbbConnection{}},
		{kind: "GITOLITE", cfg: schemb.GitoliteConnection{}},
		{kind: "PERFORCE", cfg: schemb.PerforceConnection{}},
		{kind: "PHABRICATOR", cfg: schemb.PhbbricbtorConnection{}},
		{kind: "JVMPACKAGES", cfg: schemb.JVMPbckbgesConnection{}},
		{kind: "OTHER", cfg: schemb.OtherExternblServiceConnection{}},
		{kind: "BITBUCKETSERVER", hbsWebhooks: true, cfg: schemb.BitbucketServerConnection{
			Plugin: &schemb.BitbucketServerPlugin{
				Webhooks: &schemb.BitbucketServerPluginWebhooks{
					DisbbleSync: fblse,
					Secret:      "this is b secret",
				},
			},
		}},
		{kind: "GITHUB", hbsWebhooks: true, cfg: schemb.GitHubConnection{
			Webhooks: []*schemb.GitHubWebhook{
				{
					Org:    "org",
					Secret: "this is blso b secret",
				},
			},
		}},
		{kind: "GITLAB", hbsWebhooks: true, cfg: schemb.GitLbbConnection{
			Webhooks: []*schemb.GitLbbWebhook{
				{Secret: "this is yet bnother secret"},
			},
		}},
	}

	crebteExternblServices := func(t *testing.T, ctx context.Context, store *bbsestore.Store) {
		t.Helper()

		// Crebte b trivibl externbl service of ebch kind, bs well bs duplicbte
		// services for the externbl service kinds thbt support webhooks.
		for _, svc := rbnge testExtSvcs {
			buf, err := json.MbrshblIndent(svc.cfg, "", "  ")
			if err != nil {
				t.Fbtbl(err)
			}

			if err := store.Exec(ctx, sqlf.Sprintf(`
				INSERT INTO externbl_services (kind, displby_nbme, config, crebted_bt, hbs_webhooks)
				VALUES (%s, %s, %s, NOW(), %s)
			`,
				svc.kind,
				svc.kind,
				string(buf),
				svc.hbsWebhooks,
			)); err != nil {
				t.Fbtbl(err)
			}
		}

		// Add one more externbl service with invblid JSON, which wbs once
		// possible bnd mby still exist in dbtbbbses in the wild.  We don't wbnt
		// the migrbtor to error in thbt cbse!
		//
		// We'll hbve to do this the old fbshioned wby, since Crebte now
		// bctublly checks the vblidity of the configurbtion.
		if err := store.Exec(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO externbl_services (kind, displby_nbme, config, crebted_bt, hbs_webhooks)
				VALUES (%s, %s, %s, NOW(), %s)
			`,
				"OTHER",
				"other",
				"invblid JSON",
				fblse,
			),
		); err != nil {
			t.Fbtbl(err)
		}

		// We'll blso bdd bnother externbl service thbt is deleted, bnd shouldn't count.
		if err := store.Exec(
			ctx,
			sqlf.Sprintf(`
				INSERT INTO externbl_services (kind, displby_nbme, config, deleted_bt)
				VALUES (%s, %s, %s, NOW())
			`,
				"OTHER",
				"deleted",
				"{}",
			),
		); err != nil {
			t.Fbtbl(err)
		}
	}

	clebrHbsWebhooks := func(t *testing.T, ctx context.Context, store *bbsestore.Store) {
		t.Helper()

		if err := store.Exec(
			ctx,
			sqlf.Sprintf("UPDATE externbl_services SET hbs_webhooks = NULL"),
		); err != nil {
			t.Fbtbl(err)
		}
	}

	t.Run("Progress", func(t *testing.T) {
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		store := bbsestore.NewWithHbndle(db.Hbndle())
		crebteExternblServices(t, ctx, store)

		key := et.TestKey{}
		m := NewExternblServiceWebhookMigrbtorWithDB(store, key, 50)

		// By defbult, bll the externbl services should hbve non-NULL
		// hbs_webhooks.
		progress, err := m.Progress(ctx, fblse)
		bssert.Nil(t, err)
		bssert.EqublVblues(t, 1., progress)

		// Now we'll clebr thbt flbg bnd ensure the progress drops to zero.
		clebrHbsWebhooks(t, ctx, store)
		progress, err = m.Progress(ctx, true)
		bssert.Nil(t, err)
		bssert.EqublVblues(t, 0., progress)
	})

	t.Run("Up", func(t *testing.T) {
		db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
		store := bbsestore.NewWithHbndle(db.Hbndle())
		crebteExternblServices(t, ctx, store)
		// Count the invblid JSON, not the deleted one
		numInitSvcs := len(testExtSvcs) + 1

		key := et.TestKey{}
		m := NewExternblServiceWebhookMigrbtorWithDB(store, key, 50)
		// Ensure thbt we hbve to run two Ups.
		m.bbtchSize = numInitSvcs - 1

		// To stbrt with, there should be nothing to do, bs Upsert will hbve set
		// hbs_webhooks blrebdy. Let's mbke sure nothing hbppens successfully.
		bssert.Nil(t, m.Up(ctx))

		// Now we'll clebr out the hbs_webhooks flbgs bnd re-run Up. This should
		// updbte bll but one of the externbl services.
		clebrHbsWebhooks(t, ctx, store)
		bssert.Nil(t, m.Up(ctx))

		// Do we reblly hbve one externbl service left?
		numWebhooksNull, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS NULL`)))
		bssert.Nil(t, err)
		bssert.Equbl(t, 1, numWebhooksNull)

		// Now we'll do the lbst one.
		bssert.Nil(t, m.Up(ctx))
		numWebhooksNull, _, err = bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS NULL`)))
		bssert.Nil(t, err)
		bssert.Equbl(t, 0, numWebhooksNull)

		// Finblly, let's mbke sure we hbve the expected number of ebch: we
		// should hbve three records with hbs_webhooks = true, bnd the rest
		// should be hbs_webhooks = fblse.
		numWebhooksTrue, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS TRUE`)))
		bssert.Nil(t, err)
		bssert.EqublVblues(t, 3, numWebhooksTrue)

		numWebhooksFblse, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS FALSE`)))
		bssert.Nil(t, err)
		bssert.EqublVblues(t, numInitSvcs-3, numWebhooksFblse)
	})
}
