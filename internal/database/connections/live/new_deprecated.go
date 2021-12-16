package connections

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// DEPRECATED: Use (Raw|Ensure|Migrate)FrontendDB instead.
func NewFrontendDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.Frontend
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "frontend", schema, true, observationContext)
}

// DEPRECATED: Use (Raw|Ensure|Migrate)CodeIntelDB instead.
func NewCodeIntelDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeIntel
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "codeintel", schema, true, observationContext)
}

// DEPRECATED: Use (Raw|Ensure|Migrate)CodeInsightsDB instead.
func NewCodeInsightsDB(dsn, appName string, migrate bool, observationContext *observation.Context) (*sql.DB, error) {
	schema := schemas.CodeInsights
	if !migrate {
		schema = nil
	}

	return connect(dsn, appName, "codeinsight", schema, true, observationContext)
}
