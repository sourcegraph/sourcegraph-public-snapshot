package cli

import (
	"context"
	"database/sql"
	"fmt"
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

func makeUpgradeProgressHandler(obsvCtx *observation.Context, sqlDB *sql.DB, db database.DB) http.HandlerFunc {
	// TODO(efritz) - persist plan + progress
	// TODO(efritz) - query plan and progress for display

	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		panic(err.Error()) // TODO
	}
	codeintelDB, err := connections.RawNewCodeIntelDB(obsvCtx, dsns["codeintel"], "frontend")
	if err != nil {
		panic(err.Error()) // TODO
	}
	codeinsightsDB, err := connections.RawNewCodeInsightsDB(obsvCtx, dsns["codeinsights"], "frontend")
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

		ms, err := s.List(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, upgradeProgressHandlerTemplate, value)

		for _, m := range ms {
			fmt.Fprintf(w, "<p>Migration #%d is at %.2f%%</p>", m.ID, m.Progress*100)
		}

		fmt.Fprintf(w, `<h1>FRONTEND</h1>`)
		applied, pending, failed, err := store.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}

		for _, a := range applied {
			fmt.Fprintf(w, "<p>applied migration %d</p>", a)
		}
		for _, p := range pending {
			fmt.Fprintf(w, "<p>pending migration %d</p>", p)
		}
		for _, f := range failed {
			fmt.Fprintf(w, "<p>failed migration %d</p>", f)
		}

		fmt.Fprintf(w, `<h1>CODEINTEL</h1>`)
		applied, pending, failed, err = codeintelStore.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}

		for _, a := range applied {
			fmt.Fprintf(w, "<p>applied migration %d</p>", a)
		}
		for _, p := range pending {
			fmt.Fprintf(w, "<p>pending migration %d</p>", p)
		}
		for _, f := range failed {
			fmt.Fprintf(w, "<p>failed migration %d</p>", f)
		}

		fmt.Fprintf(w, `<h1>CODEINSIGHTS</h1>`)
		applied, pending, failed, err = codeinsightsStore.Versions(ctx)
		if err != nil {
			panic(err.Error()) // TODO
		}

		for _, a := range applied {
			fmt.Fprintf(w, "<p>applied migration %d</p>", a)
		}
		for _, p := range pending {
			fmt.Fprintf(w, "<p>pending migration %d</p>", p)
		}
		for _, f := range failed {
			fmt.Fprintf(w, "<p>failed migration %d</p>", f)
		}
	}
}

const upgradeProgressHandlerTemplate = `
<body>
	<h1>FANCY MIGRATION IN PROGRESS: %d</h1>
</body>
`
