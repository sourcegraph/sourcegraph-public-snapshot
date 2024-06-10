package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// ModelConfigurationStore provides access to the `model_configurations` table.
type ModelConfigurationStore interface {
	// GetLatest returns the latest model configuration for the Sourcegraph instance.
	//
	// The very first time this function is called, before any configuration rows are
	// found in the database, this will as a side-effect create the initial row, persist
	// it, and return that. So it will always appear that there is a "latest"
	// configuration.
	GetLatest(context.Context) (*ModelConfiguration, error)

	// TODO(PRIME-292): Provide a way to persist changes to the model configuration.
	// ApplyChange(context.Context, ModelConfigurationUpdateOptions) error

	// TODO: PRIME(PRIME-324): Provide a way to list previous versions of the model
	// configuration, and replace the current settings with an earlier version.
	// List(context.Context, ListModelConfigurationOptions) ([]*types.ModelConfiguration, err error)
}

type modelConfigurationStore struct {
	*basestore.Store
	logger log.Logger
}

var _ ModelConfigurationStore = (*modelConfigurationStore)(nil)

// ModelConfiguration is a row in the `model_configurations` table.
//
// Conceptually, every row in the table is a complete set of the Sourcegraph instance's
// LLM model configuration data. Sorting these by time then, maintains a history of all
// changes made to the Sourcegraph instance's configuration.
type ModelConfiguration struct {
	// Primary Key, the "most recently created" is the current version.
	CreatedAt time.Time

	// User account who made the saved the change that lead to this version. Will be
	// nil if the user has been deleted.
	//
	// The very first model configuration will be attributed to user ID 1, systemadmin.
	CreatedBy *int32

	// BaseConfiguration is a JSON blob containing the "officially supported"
	// LLM models list, published by Sourcegraph. If NULL, the Sourcegraph instance
	// should just use the static data that is embedded in the current binary
	// from the `sourcegraph/llm-models` repo. (i.e. the set of LLM models that
	// were known when we shipped this particular Sourcegraph instance.)
	//
	// If this Sourcegraph instance has fetched a more recent version of the doc,
	// it will be persisted in this column.
	BaseConfigurationJSON *string

	// ConfigurationPatchJSON is a JSON Patch document to be applied to the base
	// configuration object. i.e. this will describe all of the customization
	// that the Sourcegraph instance administrator has added to the base document.
	//
	// For example, changing the max token counts for LLM input/output. Or
	// adding any client-specific or server-specific data.
	//
	// The "supported LLM models" document returned from this instance will be
	// the result of applying this JSON Patch to the base configuration.
	//
	// ## Handling Secrets
	//
	// Secrets may be embedded in the document, such as access keys for configuring
	// BYOK mode. For this reason we persist the data in two forms:
	// - RedactedConfigurationPatchJSON, is the JSON Patch document, with any
	//   secrets replaced with REDACTED. (With the determinination of which
	//   fields should be redacted or not determined by the LLM Models SDK.)
	// - EncryptedConfigurationPatchJSON, which is the full JSON Patch document,
	//   encrypted via the EncryptionKeyID field.
	RedactedConfigurationPatchJSON string
	// TODO(PRIME-327): Actually implement this. Currently these columns
	// are just ignored.
	EncryptedConfigurationPatchJSON string
	EncryptionKeyID                 string

	// Flags is a set of bit fields used to manage additional settings,
	// such as whether or not to check for updates or if "beta/experimental"
	// models should be allowed.
	// TODO(PRIME-326): Actually read/write this data.
	Flags int64
}

func ModelConfigurationWith(logger log.Logger, other basestore.ShareableStore) ModelConfigurationStore {
	store := basestore.NewWithHandle(other.Handle())
	return &modelConfigurationStore{
		logger: logger,
		Store:  store,
	}
}

var modelConfigurationColumns = []*sqlf.Query{
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("created_by"),
	sqlf.Sprintf("base_configuration_json"),
	sqlf.Sprintf("redacted_configuration_patch_json"),
	sqlf.Sprintf("encrypted_configuration_patch_json"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("flags"),
}

func scanModelConfiguraitonRow(scanner dbutil.Scanner) (*ModelConfiguration, error) {
	var mc ModelConfiguration
	err := scanner.Scan(
		&mc.CreatedAt,
		&mc.CreatedBy,
		&mc.BaseConfigurationJSON,
		&mc.RedactedConfigurationPatchJSON,
		&mc.EncryptedConfigurationPatchJSON,
		&mc.EncryptionKeyID,
		&mc.Flags,
	)
	return &mc, err
}

func (mcs *modelConfigurationStore) GetLatest(ctx context.Context) (*ModelConfiguration, error) {
	q := sqlf.Sprintf(`
		SELECT %s
		FROM model_configurations
		ORDER BY created_at DESC
		LIMIT 1
		`, sqlf.Join(modelConfigurationColumns, ","))
	row := mcs.QueryRow(ctx, q)
	latestConfig, err := scanModelConfiguraitonRow(row)

	// Loaded the row OK.
	if err == nil {
		return latestConfig, nil
	}
	// Generic IO error.
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// There are no rows available, i.e. no model configuration has ever been saved.
	// In this case, we persist the first row and just return that.
	defaultConf := ModelConfiguration{
		CreatedAt: time.Now(),
		// User ID 1, "systemadmin".
		CreatedBy: pointers.Ptr(int32(1)),

		// BaseConfigurationJSON is nil, so that callers will know to replace it with
		// whatever the static configuration shipped with this Sourcegraph instance.
		BaseConfigurationJSON: nil,

		RedactedConfigurationPatchJSON: "{}",
		// TODO(PRIME-327): Wire in the system-wide encryption key to use for
		// this table. And actually encrypt the "{}", storing the cipher text.
		EncryptedConfigurationPatchJSON: "{}",
		EncryptionKeyID:                 "",

		// TODO(PRIME-326): Actually read/write this data.
		Flags: 0,
	}

	insertQuery := sqlf.Sprintf(`
		INSERT INTO model_configurations (%s)
		VALUES (%s, %s, %s, %s, %s, %s, %s)
		`,
		sqlf.Join(modelConfigurationColumns, ","),
		// Actual values for the row, needing to be in the
		// same order as modelConfigurationColumns.
		defaultConf.CreatedAt,
		defaultConf.CreatedBy,
		defaultConf.BaseConfigurationJSON,
		defaultConf.RedactedConfigurationPatchJSON,
		defaultConf.EncryptedConfigurationPatchJSON,
		defaultConf.EncryptionKeyID,
		defaultConf.Flags)
	if err = mcs.Exec(ctx, insertQuery); err != nil {
		return nil, errors.Wrap(err, "inserting first row")
	}

	// Now that we've persisted the first row, we expect to be able to now fetch it.
	return mcs.GetLatest(ctx)
}
