package cli

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	migrationstore "github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
)

//go:embed templates/upgrade.html
var rawTemplate string

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

	funcs := template.FuncMap{
		"FormatPercentage": func(v float64) string { return fmt.Sprintf("%.2f%%", v*100) },
	}

	handleTemplate := func(ctx context.Context, w http.ResponseWriter) error {
		scan := basestore.NewFirstScanner(func(s dbutil.Scanner) (u upgradeStatus, _ error) {
			var rawPlan []byte
			if err := s.Scan(
				&u.FromVersion,
				&u.ToVersion,
				&rawPlan,
				&u.StartedAt,
				&dbutil.NullTime{Time: &u.FinishedAt},
				&dbutil.NullBool{B: &u.Success},
			); err != nil {
				return upgradeStatus{}, err
			}

			if err := json.Unmarshal(rawPlan, &u.Plan); err != nil {
				return upgradeStatus{}, err
			}

			return u, nil
		})
		upgrade, _, err := scan(db.QueryContext(ctx, `
			SELECT
				from_version,
				to_version,
				plan,
				started_at,
				finished_at,
				success
			FROM upgrade_logs
			ORDER BY id DESC
			LIMIT 1
		`))
		if err != nil {
			return err
		}

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
		oobmigrations, err := oobmigrationStore.GetByIDs(ctx, upgrade.Plan.OutOfBandMigrationIDs)
		if err != nil {
			return err
		}

		unfinishedOutOfBandMigrations := oobmigrations[:0]
		for _, migration := range oobmigrations {
			if migration.Progress != 1 {
				filteredErrs := migration.Errors[:0]
				for _, err := range migration.Errors {
					if err.Created.After(upgrade.StartedAt) {
						filteredErrs = append(filteredErrs, err)
					}
				}
				migration.Errors = filteredErrs

				unfinishedOutOfBandMigrations = append(unfinishedOutOfBandMigrations, migration)
			}
		}

		tmpl, err := template.New("index").Funcs(funcs).Parse(rawTemplate)
		if err != nil {
			return err
		}

		return tmpl.Execute(w, struct {
			Upgrade                          upgradeStatus
			Frontend                         migrationStatus
			CodeIntel                        migrationStatus
			CodeInsights                     migrationStatus
			NumUnfinishedOutOfBandMigrations int
			OutOfBandMigrations              []oobmigration.Migration
		}{
			Upgrade:                          upgrade,
			Frontend:                         getMigrationStatus(upgrade.Plan.MigrationNames["frontend"], upgrade.Plan.Migrations["frontend"], frontendApplied, frontendPending, frontendFailed),
			CodeIntel:                        getMigrationStatus(upgrade.Plan.MigrationNames["codeintel"], upgrade.Plan.Migrations["codeintel"], codeintelApplied, codeintelPending, codeIntelFailed),
			CodeInsights:                     getMigrationStatus(upgrade.Plan.MigrationNames["codeinsights"], upgrade.Plan.Migrations["codeinsights"], codeinsightsApplied, codeinsightsPending, codeinsightsFailed),
			NumUnfinishedOutOfBandMigrations: len(unfinishedOutOfBandMigrations),
			OutOfBandMigrations:              unfinishedOutOfBandMigrations,
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := handleTemplate(r.Context(), w); err != nil {
			obsvCtx.Logger.Error("failed to handle upgrade UI request", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		}
	}, nil
}

type upgradeStatus struct {
	FromVersion string
	ToVersion   string
	Plan        upgradestore.UpgradePlan
	StartedAt   time.Time
	FinishedAt  time.Time
	Success     bool
}

type migrationState struct {
	ID    int
	Name  string
	State string
}

type migrationStatus struct {
	NumMigrationsRequired int
	HasFailure            bool
	Migrations            []migrationState
}

// adjust this to show some leading applied migrations
const numLeadingAppliedMigrations = 0

func getMigrationStatus(migrationNames map[int]string, expected, applied, pending, failed []int) migrationStatus {
	expectedMap := map[int]struct{}{}
	for _, id := range expected {
		expectedMap[id] = struct{}{}
	}
	appliedMap := map[int]struct{}{}
	for _, id := range applied {
		appliedMap[id] = struct{}{}
	}
	pendingMap := map[int]struct{}{}
	for _, id := range pending {
		pendingMap[id] = struct{}{}
	}
	failedMap := map[int]struct{}{}
	for _, id := range failed {
		failedMap[id] = struct{}{}
	}

	hasFailure := false
	numMigrationsRequired := 0
	migrations := make([]migrationState, 0, len(expected))

	for _, id := range expected {
		state := ""
		if _, ok := appliedMap[id]; ok {
			state = "applied"
		} else if _, ok := pendingMap[id]; ok {
			state = "in-progress"
		} else if _, ok := failedMap[id]; ok {
			state = "failed"
			hasFailure = true
			numMigrationsRequired++
		} else {
			state = "queued"
			numMigrationsRequired++
		}

		migrations = append(migrations, migrationState{ID: id, Name: migrationNames[id], State: state})
	}

	for id := range pendingMap {
		if _, ok := expectedMap[id]; !ok {
			migrations = append(migrations, migrationState{ID: id, Name: migrationNames[id], State: "pending"})
		}
	}
	for id := range failedMap {
		if _, ok := expectedMap[id]; !ok {
			migrations = append(migrations, migrationState{ID: id, Name: migrationNames[id], State: "failed"})
		}
	}

	numLeadingApplied := 0
	for _, migration := range migrations {
		if migration.State != "applied" {
			break
		}
		numLeadingApplied++
	}
	strip := numLeadingApplied - numLeadingAppliedMigrations
	if strip < 0 {
		strip = 0
	}

	return migrationStatus{
		NumMigrationsRequired: numMigrationsRequired,
		HasFailure:            hasFailure,
		Migrations:            migrations[strip:],
	}
}
