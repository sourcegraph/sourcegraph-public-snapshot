package codeintel

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	DiagnosticsCountMigrationBatchSize    int
	DefinitionsCountMigrationBatchSize    int
	ReferencesCountMigrationBatchSize     int
	DocumentColumnSplitMigrationBatchSize int
}

var config = &Config{}

func init() {
	config.DiagnosticsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of document records to migrate at a time.")
	config.DefinitionsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DEFINITIONS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of definition records to migrate at once.")
	config.ReferencesCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_REFERENCES_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of reference records to migrate at a time.")
	config.DocumentColumnSplitMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DOCUMENT_COLUMN_SPLIT_MIGRATION_BATCH_SIZE", "100", "The maximum number of document records to migrate at a time.")
}
