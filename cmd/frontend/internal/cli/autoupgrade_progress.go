package cli

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

//go:embed templates/upgrade.html
var indexHTML string

type migrationStatus struct {
	Percentage string
	Applied    []int
	Pending    []int
	Failed     []int
}

func makeUpgradeProgressHandler(obsvCtx *observation.Context, sqlDB *sql.DB, db database.DB) (http.HandlerFunc, error) {
	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		return nil, err
	}
	codeintelDB, err := connections.RawNewCodeIntelDB(obsvCtx, dsns["codeintel"], appName)
	if err != nil {
		return nil, err
	}
	codeinsightsDB, err := connections.RawNewCodeInsightsDB(obsvCtx, dsns["codeinsights"], appName)
	if err != nil {
		return nil, err
	}

	store := migrationstore.NewWithDB(obsvCtx, sqlDB, schemas.Frontend.MigrationsTableName)
	codeintelStore := migrationstore.NewWithDB(obsvCtx, codeintelDB, schemas.CodeIntel.MigrationsTableName)
	codeinsightsStore := migrationstore.NewWithDB(obsvCtx, codeinsightsDB, schemas.CodeInsights.MigrationsTableName)

	ctx := context.Background()
	oobmigrationStore := oobmigration.NewStoreWithDB(db)
	if err := oobmigrationStore.SynchronizeMetadata(ctx); err != nil {
		return nil, err
	}

	handleTemplate := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		frontendApplied, frontendPending, frontendFailed, err := store.Versions(ctx)
		if err != nil {
			return err
		}
		codeintelApplied, codeintelPending, codeIntelFailed, err := codeintelStore.Versions(ctx)
		if err != nil {
			return err
		}
		codeinsightsApplied, codeinsightsPending, codeinsightsFailed, err := codeinsightsStore.Versions(ctx)
		if err != nil {
			return err
		}
		oobmigrations, err := oobmigrationStore.List(ctx)
		if err != nil {
			return err
		}

		tmpl, err := template.New("index").Parse(indexHTML)
		if err != nil {
			return err
		}

		return tmpl.Execute(w, struct {
			Frontend            migrationStatus
			CodeIntel           migrationStatus
			CodeInsights        migrationStatus
			OutOfBandMigrations []oobmigration.Migration
		}{
			Frontend:            getMigrationStatus(frontendApplied, frontendPending, frontendFailed),
			CodeIntel:           getMigrationStatus(codeintelApplied, codeintelPending, codeIntelFailed),
			CodeInsights:        getMigrationStatus(codeinsightsApplied, codeinsightsPending, codeinsightsFailed),
			OutOfBandMigrations: oobmigrations,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := handleTemplate(r.Context(), w, r); err != nil {
			obsvCtx.Logger.Error("failed to handle upgrade UI request", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}, nil
}

func getMigrationStatus(applied, pending, failed []int) migrationStatus {
	return migrationStatus{
		Percentage: getProgressPercentage(applied, pending, failed),
		Applied:    applied,
		Pending:    pending,
		Failed:     failed,
	}
}

func getProgressPercentage(applied, pending, failed []int) string {
	total := len(applied) + len(pending) + len(failed)
	if total == 0 {
		return "100%"
	}

	return fmt.Sprintf("%d%%", int(float64(len(applied))/float64(total)*100))
}
