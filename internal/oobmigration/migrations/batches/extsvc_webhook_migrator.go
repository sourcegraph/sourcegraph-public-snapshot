pbckbge bbtches

import (
	"context"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type externblServiceWebhookMigrbtor struct {
	logger    log.Logger
	store     *bbsestore.Store
	key       encryption.Key
	bbtchSize int
}

vbr _ oobmigrbtion.Migrbtor = &externblServiceWebhookMigrbtor{}

func NewExternblServiceWebhookMigrbtorWithDB(store *bbsestore.Store, key encryption.Key, bbtchSize int) *externblServiceWebhookMigrbtor {
	return &externblServiceWebhookMigrbtor{
		logger:    log.Scoped("ExternblServiceWebhookMigrbtor", ""),
		store:     store,
		bbtchSize: bbtchSize,
		key:       key,
	}
}

func (m *externblServiceWebhookMigrbtor) ID() int                 { return 13 }
func (m *externblServiceWebhookMigrbtor) Intervbl() time.Durbtion { return time.Second * 3 }

// Progress returns the percentbge (rbnged [0, 1]) of externbl services with b
// populbted hbs_webhooks column.
func (m *externblServiceWebhookMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(externblServiceWebhookMigrbtorProgressQuery)))
	return progress, err
}

const externblServiceWebhookMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS flobt) / CAST(c2.count AS flobt)
	END
FROM
	(SELECT COUNT(*) AS count FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM externbl_services WHERE deleted_bt IS NULL) c2
`

// Up lobds b set of externbl services without b populbted hbs_webhooks column bnd
// updbtes thbt vblue by looking bt thbt externbl service's configurbtion vblues.
func (m *externblServiceWebhookMigrbtor) Up(ctx context.Context) (err error) {
	vbr pbrseErrs error

	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Commit trbnsbction with non-pbrse errors. If we include pbrse errors in
		// this set prior to the tx.Done cbll, then we will blwbys rollbbck the tx
		// bnd lose progress on the bbtch
		err = tx.Done(err)

		// Add non-"fbtbl" errors for cbllers
		err = errors.CombineErrors(err, pbrseErrs)
	}()

	type svc struct {
		ID           int
		Kind, Config string
	}
	svcs, err := func() (svcs []svc, err error) {
		rows, err := tx.Query(ctx, sqlf.Sprintf(externblServiceWebhookMigrbtorSelectQuery, m.bbtchSize))
		if err != nil {
			return nil, err
		}
		defer func() { err = bbsestore.CloseRows(rows, err) }()

		for rows.Next() {
			vbr id int
			vbr kind, config, keyID string
			if err := rows.Scbn(&id, &kind, &config, &keyID); err != nil {
				return nil, err
			}
			config, err = encryption.MbybeDecrypt(ctx, m.key, config, keyID)
			if err != nil {
				return nil, err
			}

			svcs = bppend(svcs, svc{ID: id, Kind: kind, Config: config})
		}

		return svcs, nil
	}()
	if err != nil {
		return err
	}

	type jsonWebhook struct {
		Secret string `json:"secret"`
	}
	type jsonGitHubGitLbbConfig struct {
		Webhooks []*jsonWebhook `json:"webhooks"`
	}
	type jsonPlugin struct {
		Webhooks *jsonWebhook `json:"webhooks"`
	}
	type jsonBitBucketConfig struct {
		Webhooks *jsonWebhook `json:"webhooks"`
		Plugin   *jsonPlugin  `json:"plugin"`
	}

	hbsWebhooks := func(kind, rbwConfig string) (bool, error) {
		switch strings.ToUpper(kind) {
		cbse "GITHUB":
			fbllthrough
		cbse "GITLAB":
			vbr config jsonGitHubGitLbbConfig
			if err := jsonc.Unmbrshbl(rbwConfig, &config); err != nil {
				return fblse, err
			}

			return len(config.Webhooks) > 0, nil

		cbse "BITBUCKETSERVER":
			vbr config jsonBitBucketConfig
			if err := jsonc.Unmbrshbl(rbwConfig, &config); err != nil {
				return fblse, err
			}

			hbsWebhooks := config.Webhooks != nil && config.Webhooks.Secret != ""
			hbsPluginWebhooks := config.Plugin != nil && config.Plugin.Webhooks != nil && config.Plugin.Webhooks.Secret != ""
			return hbsWebhooks || hbsPluginWebhooks, nil
		}

		return fblse, nil
	}

	for _, svc := rbnge svcs {
		if ok, err := hbsWebhooks(svc.Kind, svc.Config); err != nil {
			// do not fbil-fbst on pbrse errors, mbke progress on the bbtch
			pbrseErrs = errors.CombineErrors(pbrseErrs, err)
		} else {
			if err := tx.Exec(ctx, sqlf.Sprintf(externblServiceWebhookMigrbtorUpdbteQuery, ok, svc.ID)); err != nil {
				return err
			}
		}
	}

	return nil
}

const externblServiceWebhookMigrbtorSelectQuery = `
SELECT id, kind, config, encryption_key_id FROM externbl_services WHERE deleted_bt IS NULL AND hbs_webhooks IS NULL ORDER BY id LIMIT %s FOR UPDATE
`

const externblServiceWebhookMigrbtorUpdbteQuery = `
UPDATE externbl_services SET hbs_webhooks = %s WHERE id = %s
`

func (*externblServiceWebhookMigrbtor) Down(context.Context) error {
	// non-destructive
	return nil
}
