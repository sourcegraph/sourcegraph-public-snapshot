package batches

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type externalServiceWebhookMigrator struct {
	logger    log.Logger
	store     *basestore.Store
	key       encryption.Key
	batchSize int
}

var _ oobmigration.Migrator = &externalServiceWebhookMigrator{}

func NewExternalServiceWebhookMigratorWithDB(store *basestore.Store, key encryption.Key, batchSize int) *externalServiceWebhookMigrator {
	return &externalServiceWebhookMigrator{
		logger:    log.Scoped("ExternalServiceWebhookMigrator"),
		store:     store,
		batchSize: batchSize,
		key:       key,
	}
}

func (m *externalServiceWebhookMigrator) ID() int                 { return 13 }
func (m *externalServiceWebhookMigrator) Interval() time.Duration { return time.Second * 3 }

// Progress returns the percentage (ranged [0, 1]) of external services with a
// populated has_webhooks column.
func (m *externalServiceWebhookMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(externalServiceWebhookMigratorProgressQuery)))
	return progress, err
}

const externalServiceWebhookMigratorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS float) / CAST(c2.count AS float)
	END
FROM
	(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM external_services WHERE deleted_at IS NULL) c2
`

// Up loads a set of external services without a populated has_webhooks column and
// updates that value by looking at that external service's configuration values.
func (m *externalServiceWebhookMigrator) Up(ctx context.Context) (err error) {
	var parseErrs error

	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Commit transaction with non-parse errors. If we include parse errors in
		// this set prior to the tx.Done call, then we will always rollback the tx
		// and lose progress on the batch
		err = tx.Done(err)

		// Add non-"fatal" errors for callers
		err = errors.CombineErrors(err, parseErrs)
	}()

	type svc struct {
		ID           int
		Kind, Config string
	}
	svcs, err := func() (svcs []svc, err error) {
		rows, err := tx.Query(ctx, sqlf.Sprintf(externalServiceWebhookMigratorSelectQuery, m.batchSize))
		if err != nil {
			return nil, err
		}
		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			var id int
			var kind, config, keyID string
			if err := rows.Scan(&id, &kind, &config, &keyID); err != nil {
				return nil, err
			}
			config, err = encryption.MaybeDecrypt(ctx, m.key, config, keyID)
			if err != nil {
				return nil, err
			}

			svcs = append(svcs, svc{ID: id, Kind: kind, Config: config})
		}

		return svcs, nil
	}()
	if err != nil {
		return err
	}

	type jsonWebhook struct {
		Secret string `json:"secret"`
	}
	type jsonGitHubGitLabConfig struct {
		Webhooks []*jsonWebhook `json:"webhooks"`
	}
	type jsonPlugin struct {
		Webhooks *jsonWebhook `json:"webhooks"`
	}
	type jsonBitBucketConfig struct {
		Webhooks *jsonWebhook `json:"webhooks"`
		Plugin   *jsonPlugin  `json:"plugin"`
	}

	hasWebhooks := func(kind, rawConfig string) (bool, error) {
		switch strings.ToUpper(kind) {
		case "GITHUB":
			fallthrough
		case "GITLAB":
			var config jsonGitHubGitLabConfig
			if err := jsonc.Unmarshal(rawConfig, &config); err != nil {
				return false, err
			}

			return len(config.Webhooks) > 0, nil

		case "BITBUCKETSERVER":
			var config jsonBitBucketConfig
			if err := jsonc.Unmarshal(rawConfig, &config); err != nil {
				return false, err
			}

			hasWebhooks := config.Webhooks != nil && config.Webhooks.Secret != ""
			hasPluginWebhooks := config.Plugin != nil && config.Plugin.Webhooks != nil && config.Plugin.Webhooks.Secret != ""
			return hasWebhooks || hasPluginWebhooks, nil
		}

		return false, nil
	}

	for _, svc := range svcs {
		if ok, err := hasWebhooks(svc.Kind, svc.Config); err != nil {
			// do not fail-fast on parse errors, make progress on the batch
			parseErrs = errors.CombineErrors(parseErrs, err)
		} else {
			if err := tx.Exec(ctx, sqlf.Sprintf(externalServiceWebhookMigratorUpdateQuery, ok, svc.ID)); err != nil {
				return err
			}
		}
	}

	return nil
}

const externalServiceWebhookMigratorSelectQuery = `
SELECT id, kind, config, encryption_key_id FROM external_services WHERE deleted_at IS NULL AND has_webhooks IS NULL ORDER BY id LIMIT %s FOR UPDATE
`

const externalServiceWebhookMigratorUpdateQuery = `
UPDATE external_services SET has_webhooks = %s WHERE id = %s
`

func (*externalServiceWebhookMigrator) Down(context.Context) error {
	// non-destructive
	return nil
}
