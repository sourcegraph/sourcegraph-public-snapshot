package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/lib/pq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
)

// pkgs provides access to the `pkgs` table.
//
// For a detailed overview of the schema, see schema.txt.
type pkgs struct{}

func (p *pkgs) UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, pks []lspext.PackageInformation) (err error) {
	err = db.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		// Update the pkgs table.
		err = p.update(ctx, tx, repo, language, pks)
		if err != nil {
			return errors.Wrap(err, "pkgs.update")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "executing transaction")
	}
	return nil
}

type dbQueryer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (p *pkgs) update(ctx context.Context, tx *sql.Tx, indexRepo api.RepoID, language string, pks []lspext.PackageInformation) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pkgs.update "+language)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("pkgs", len(pks))

	// First, create a temporary table.
	_, err = tx.ExecContext(ctx, `CREATE TEMPORARY TABLE new_pkgs (
		pkg jsonb NOT NULL,
		language text NOT NULL,
		repo_id integer NOT NULL
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}
	span.LogFields(otlog.String("event", "created temp table"))

	// Copy the new deps into the temporary table.
	copy, err := tx.Prepare(pq.CopyIn("new_pkgs",
		"repo_id",
		"language",
		"pkg",
	))
	if err != nil {
		return errors.Wrap(err, "prepare copy in")
	}
	defer copy.Close()
	span.LogFields(otlog.String("event", "prepared copy in"))

	for _, r := range pks {
		pkgData, err := json.Marshal(r.Package)
		if err != nil {
			return errors.Wrap(err, "marshaling package metadata to JSON")
		}

		if _, err := copy.Exec(
			indexRepo,       // repo_id
			language,        // language
			string(pkgData), // pkg
		); err != nil {
			return errors.Wrap(err, "executing pkg copy")
		}
	}
	span.LogFields(otlog.String("event", "executed all pkg copy"))
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}
	span.LogFields(otlog.String("event", "executed copy"))

	if _, err := tx.ExecContext(ctx, `DELETE FROM pkgs WHERE language=$1 AND repo_id=$2`, language, indexRepo); err != nil {
		return errors.Wrap(err, "executing pkgs deletion")
	}
	span.LogFields(otlog.String("event", "executed pkgs deletion"))

	// Insert from temporary table into the real table.
	_, err = tx.ExecContext(ctx, `INSERT INTO pkgs(
		repo_id,
		language,
		pkg
	)
	SELECT p.repo_id,
		p.language,
		p.pkg
	FROM new_pkgs p;
	`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	span.LogFields(otlog.String("event", "executed final insertion from temp table"))
	return nil
}

func (p *pkgs) ListPackages(ctx context.Context, op *api.ListPackagesOp) (pks []*api.PackageInfo, err error) {
	if db.Mocks.Pkgs.ListPackages != nil {
		return db.Mocks.Pkgs.ListPackages(ctx, op)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "pkgs.ListPackages")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Lang", op.Lang)
	span.SetTag("PkgQuery", op.PkgQuery)

	var args []interface{}
	arg := func(a interface{}) string {
		args = append(args, a)
		return fmt.Sprintf("$%d", len(args))
	}

	var whereClauses []string
	if op.PkgQuery != nil {
		containmentQuery, err := json.Marshal(op.PkgQuery)
		if err != nil {
			return nil, errors.New("marshaling op.PkgQuery")
		}
		whereClauses = append(whereClauses, `pkgs.pkg @> `+arg(string(containmentQuery)))
	}
	if op.RepoID != 0 {
		whereClauses = append(whereClauses, `repo_id=`+arg(op.RepoID))
	}
	if op.Lang != "" {
		whereClauses = append(whereClauses, `pkgs.language=`+arg(op.Lang))
	}
	if len(whereClauses) == 0 {
		return nil, fmt.Errorf("no filtering options specified, must specify at least one")
	}
	whereSQL := "(" + strings.Join(whereClauses, ") AND (") + ")"
	sql := `
		SELECT pkgs.*
		FROM pkgs INNER JOIN repo ON pkgs.repo_id=repo.id
		WHERE ` + whereSQL + `
		ORDER BY repo.created_at ASC NULLS LAST, pkgs.repo_id ASC
		LIMIT ` + arg(op.Limit)
	rows, err := dbconn.Global.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var rawPkgs []*api.PackageInfo
	for rows.Next() {
		var (
			pkg, lang string
			repo      api.RepoID
		)
		if err := rows.Scan(&repo, &lang, &pkg); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		r := api.PackageInfo{
			RepoID: repo,
			Lang:   lang,
			// NOTE: Dependency info (in api.PackageInfo's Dependencies field) is not set
			// here because it is stored separately in the global_dep table in a way that
			// is slow and difficult to get in this code path. Currently callers that use
			// DB-persisted package info do not need the dependency info, so this is
			// acceptable.
		}
		if err := json.Unmarshal([]byte(pkg), &r.Pkg); err != nil {
			return nil, errors.Wrap(err, "unmarshaling xdependencies metadata from sql scan")
		}
		rawPkgs = append(rawPkgs, &r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}

	return rawPkgs, nil
}

func (p *pkgs) Delete(ctx context.Context, repo api.RepoID) error {
	if db.Mocks.Pkgs.Delete != nil {
		return db.Mocks.Pkgs.Delete(ctx, repo)
	}

	_, err := dbconn.Global.ExecContext(ctx, `DELETE FROM pkgs WHERE repo_id=$1`, repo)
	return err
}
