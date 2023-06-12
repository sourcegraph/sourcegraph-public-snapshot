package cli

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

//go:embed index.html
var indexHTML string

var _ embed.FS

type migrationStatus struct {
	Percentage string
	Applied    []int
	Pending    []int
	Failed     []int
}

func makeUpgradeProgressHandler(obsvCtx *observation.Context, sqlDB *sql.DB, db database.DB) http.HandlerFunc {
	// TODO(efritz) - persist plan + progress
	// TODO(efritz) - query plan and progress for display

	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		panic(err.Error()) // TODO
	}
	codeintelDB, err := connections.RawNewCodeIntelDB(obsvCtx, dsns["codeintel"], appName)
	if err != nil {
		panic(err.Error()) // TODO
	}
	codeinsightsDB, err := connections.RawNewCodeInsightsDB(obsvCtx, dsns["codeinsights"], appName)
	if err != nil {
		panic(err.Error()) // TODO
	}

	store := migrationstore.NewWithDB(obsvCtx, sqlDB, schemas.Frontend.MigrationsTableName)
	codeintelStore := migrationstore.NewWithDB(obsvCtx, codeintelDB, schemas.CodeIntel.MigrationsTableName)
	codeinsightsStore := migrationstore.NewWithDB(obsvCtx, codeinsightsDB, schemas.CodeInsights.MigrationsTableName)

	ctx := context.Background()
	s := oobmigration.NewStoreWithDB(db)
	if err := s.SynchronizeMetadata(ctx); err != nil {
		panic(err.Error()) // TODO
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var value int
		if err := func() (err error) {
			value, _, err = basestore.ScanFirstInt(db.QueryContext(ctx, `SELECT 4`))
			return err
		}(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// TODO (numbers88s): don't know what this is, leaving it for @efritz.
		_ = value

		ms, err := s.List(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}
		// TODO (numbers88s): don't know what this is, leaving it for @efritz.
		_ = ms

		// FRONTEND
		applied, pending, failed, err := store.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}
		frontEndData := getMigrationStatus(applied, pending, failed)

		// CODEINTEL
		applied, pending, failed, err = codeintelStore.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}
		codeIntelData := getMigrationStatus(applied, pending, failed)

		// CODEINSIGHTS
		applied, pending, failed, err = codeinsightsStore.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}
		codeInsightsData := getMigrationStatus(applied, pending, failed)

		data := struct {
			Frontend     migrationStatus
			CodeIntel    migrationStatus
			CodeInsights migrationStatus
		}{
			Frontend:     frontEndData,
			CodeIntel:    codeIntelData,
			CodeInsights: codeInsightsData,
		}

		tmpl, err := template.New("index").Parse(indexHTML)
		if err != nil {
			panic(err.Error()) // TODO
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err.Error()) // TODO
		}
	}
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

	val := int(float64(len(applied)) / float64(total) * 100)

	return fmt.Sprintf("%d%%", val)
}
