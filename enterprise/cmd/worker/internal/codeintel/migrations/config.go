package migrations

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	DiagnosticsCountMigrationBatchSize        int
	DiagnosticsCountMigrationBatchInterval    time.Duration
	DefinitionsCountMigrationBatchSize        int
	DefinitionsCountMigrationBatchInterval    time.Duration
	ReferencesCountMigrationBatchSize         int
	ReferencesCountMigrationBatchInterval     time.Duration
	DocumentColumnSplitMigrationBatchSize     int
	DocumentColumnSplitMigrationBatchInterval time.Duration
	APIDocsSearchMigrationBatchSize           int
	APIDocsSearchMigrationBatchInterval       time.Duration
	CommittedAtMigrationBatchSize             int
	CommittedAtMigrationBatchInterval         time.Duration
	ReferenceCountMigrationBatchSize          int
	ReferenceCountMigrationBatchInterval      time.Duration
}

var config = &Config{}

func init() {
	config.DiagnosticsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of document records to migrate at a time.")
	config.DiagnosticsCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DIAGNOSTICS_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.DefinitionsCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DEFINITIONS_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of definition records to migrate at once.")
	config.DefinitionsCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DEFINITIONS_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.ReferencesCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_REFERENCES_COUNT_MIGRATION_BATCH_SIZE", "1000", "The maximum number of reference records to migrate at a time.")
	config.ReferencesCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_REFERENCES_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.DocumentColumnSplitMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_DOCUMENT_COLUMN_SPLIT_MIGRATION_BATCH_SIZE", "100", "The maximum number of document records to migrate at a time.")
	config.DocumentColumnSplitMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_DOCUMENT_COLUMN_SPLIT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.APIDocsSearchMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_API_DOCS_SEARCH_MIGRATION_BATCH_SIZE", "1", "The maximum number of bundles to migrate at a time.")
	config.APIDocsSearchMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_API_DOCS_SEARCH_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.CommittedAtMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_COMMITTED_AT_MIGRATION_BATCH_SIZE", "100", "The maximum number of upload records to migrate at a time.")
	config.CommittedAtMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_COMMITTED_AT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
	config.ReferenceCountMigrationBatchSize = config.GetInt("PRECISE_CODE_INTEL_REFERENCE_COUNT_MIGRATION_BATCH_SIZE", "100", "The maximum number of upload records to migrate at a time.")
	config.ReferenceCountMigrationBatchInterval = config.GetInterval("PRECISE_CODE_INTEL_REFERENCE_COUNT_MIGRATION_BATCH_INTERVAL", "1s", "The timeout between processing migration batches.")
}
