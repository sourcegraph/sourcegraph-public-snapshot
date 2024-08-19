package cliutil

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

//go:embed tenant_up.sql tenant_down.sql
var tenantTemplates embed.FS

// EnforceTenant is a temporary command while we still have tenant_id not
// enforced.
func EnforceTenant(commandName string, factory RunnerFactory, outFactory OutputFactory) *cli.Command {
	schemaNamesFlag := &cli.StringSliceFlag{
		Name:    "schema",
		Usage:   "The target `schema(s)` to modify. Comma-separated values are accepted. Possible values are 'frontend', 'codeintel', 'codeinsights' and 'all'.",
		Value:   cli.NewStringSlice("all"),
		Aliases: []string{"db"},
	}

	disableFlag := &cli.BoolFlag{
		Name:  "disable",
		Usage: "Disables RLS instead of enabling.",
		Value: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schemaNames := sanitizeSchemaNames(schemaNamesFlag.Get(cmd), out)
		if len(schemaNames) == 0 {
			return flagHelp(out, "supply a schema via -db")
		}
		enforce := !disableFlag.Get(cmd)

		templates, err := template.ParseFS(tenantTemplates, "*.sql")
		if err != nil {
			return err
		}
		var queryTemplate *template.Template
		if enforce {
			queryTemplate = templates.Lookup("tenant_up.sql")
		} else {
			queryTemplate = templates.Lookup("tenant_down.sql")
		}

		count := 0
		err = forEachTable(ctx, factory, schemaNames, func(ctx context.Context, tx *sql.Tx, table schemas.TableDescription) error {
			i := slices.IndexFunc(table.Columns, func(column schemas.ColumnDescription) bool {
				return column.Name == "tenant_id"
			})
			if i < 0 {
				return nil
			}

			count++

			var query strings.Builder
			err := queryTemplate.Execute(&query, map[string]string{"Table": table.Name})
			if err != nil {
				return err
			}

			_, err = tx.ExecContext(ctx, query.String())
			return err
		})
		if err != nil {
			return err
		}

		if enforce {
			out.WriteLine(output.Emoji(output.EmojiSuccess, fmt.Sprintf("tenant_id enforced for %d tables!", count)))
		} else {
			out.WriteLine(output.Emoji(output.EmojiSuccess, fmt.Sprintf("tenant_id enforcement disabled for %d tables!", count)))
		}

		return nil
	})

	return &cli.Command{
		Name:        "enforce-tenant-id",
		ArgsUsage:   "",
		Usage:       "Rewrite schemas definitions to enable RLS on tenant_id columns",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNamesFlag,
			disableFlag,
		},
	}
}

func forEachTable(ctx context.Context, factory RunnerFactory, schemaNames []string, f func(context.Context, *sql.Tx, schemas.TableDescription) error) error {
	r, err := setupRunner(factory, schemaNames...)
	if err != nil {
		return err
	}

	for _, schemaName := range schemaNames {
		store, err := r.Store(ctx, schemaName)
		if err != nil {
			return err
		}

		descriptions, err := store.Describe(ctx)
		if err != nil {
			return err
		}

		description := descriptions["public"]

		db, ok := basestore.Raw(store)
		if !ok {
			return errors.New("store does not support direct database handle access")
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		for _, table := range description.Tables {
			err := f(ctx, tx, table)
			if err != nil {
				_ = tx.Rollback()
				return errors.Wrapf(err, "on table %s.%s", schemaName, table.Name)
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
