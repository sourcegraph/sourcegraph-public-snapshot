package localstore

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

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

type dbPkgs struct{}

func (*dbPkgs) CreateTable() string {
	return `CREATE table pkgs (
		repo_id integer NOT NULL,
		lang text NOT NULL,
		pkg jsonb NOT NULL
	);
	CREATE INDEX pkg_pkg_idx ON pkgs USING gin (pkg jsonb_path_ops);
	CREATE INDEX pkg_lang_idx on pkgs USING btree (lang);
	CREATE INDEX pkg_repo_idx ON pkgs USING btree (repo_id);
`
}

func (*dbPkgs) DropTable() string {
	return `DROP TABLE IF EXISTS pkgs CASCADE;`
}

func (p *dbPkgs) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp, langs []*inventory.Lang) error {
	var errs []string
	for _, lang := range langs {
		langName := strings.ToLower(lang.Name)

		if _, enabled := globalDepEnabledLangs[langName]; !enabled {
			continue
		}
		if err := p.refreshIndexForLanguage(ctx, langName, op); err != nil {
			log15.Crit("refreshing index failed", "language", langName, "error", err)
			errs = append(errs, fmt.Sprintf("refreshing index failed language=%s error=%v", langName, err))
		}
	}
	if len(errs) == 1 {
		return errors.New(errs[0])
	} else if len(errs) > 1 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (p *dbPkgs) refreshIndexForLanguage(ctx context.Context, language string, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "dbPkgs.refreshIndexForLanguage "+language)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vcs := "git" // TODO: store VCS type in *sourcegraph.Repo object.

	// Query all external dependencies for the repository. We do this using the
	// "<language>_bg" mode which runs this request on a separate language
	// server explicitly for background tasks such as workspace/xdependencies.
	// This makes it such that indexing repositories does not interfere in
	// terms of resource usage with real user requests.
	rootPath := vcs + "://" + op.RepoURI + "?" + op.CommitID
	var pkgs []lspext.PackageInformation
	err = xlangCall(ctx, language+"_bg", rootPath, "workspace/packages", map[string]string{}, &pkgs)
	if err != nil {
		return errors.Wrap(err, "LSP Call workspace/packages")
	}

	err = dbutil.Transaction(ctx, globalGraphDBH.Db, func(tx *sql.Tx) error {
		// Update the pkgs table.
		err = p.update(ctx, tx, op.RepoID, language, pkgs)
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
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func (p *dbPkgs) get(ctx context.Context, db dbQueryer, whereSQL string, args ...interface{}) (packages []sourcegraph.PackageInfo, err error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM pkgs %s", whereSQL), args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			repoID   int32
			language string
			pkg      string
		)
		if err := rows.Scan(&repoID, &language, &pkg); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		p := sourcegraph.PackageInfo{
			RepoID: repoID,
			Lang:   language,
		}
		if err := json.Unmarshal([]byte(pkg), &p.Pkg); err != nil {
			return nil, errors.Wrap(err, "unmarshaling package metadata from SQL scan")
		}
		packages = append(packages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return packages, nil
}

func (p *dbPkgs) update(ctx context.Context, tx *sql.Tx, indexRepo int32, language string, pkgs []lspext.PackageInformation) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "dbPkgs.update "+language)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("pkgs", len(pkgs))

	// First, create a temporary table.
	_, err = tx.Exec(`CREATE TEMPORARY TABLE new_pkgs (
		pkg jsonb NOT NULL,
		lang text NOT NULL,
		repo_id integer NOT NULL
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}
	span.LogEvent("created temp table")

	// Copy the new deps into the temporary table.
	copy, err := tx.Prepare(pq.CopyIn("new_pkgs",
		"repo_id",
		"lang",
		"pkg",
	))
	if err != nil {
		return errors.Wrap(err, "prepare copy in")
	}
	defer copy.Close()
	span.LogEvent("prepared copy in")

	for _, r := range pkgs {
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
	span.LogEvent("executed all pkg copy")
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}
	span.LogEvent("executed copy")

	if _, err := tx.Exec(`DELETE FROM pkgs WHERE lang=$1 AND repo_id=$2`, language, indexRepo); err != nil {
		return errors.Wrap(err, "executing pkgs deletion")
	}
	span.LogEvent("executed pkgs deletion")

	// Insert from temporary table into the real table.
	_, err = tx.Exec(`INSERT INTO pkgs(
		repo_id,
		lang,
		pkg
	)
	SELECT p.repo_id, p.lang, p.pkg
	FROM new_pkgs p;`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	span.LogEvent("executed final insertion from temp table")
	return nil
}

func (p *dbPkgs) ListPackages(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "dbPkgs.ListPackages")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Lang", op.Lang)
	span.SetTag("PkgQuery", op.PkgQuery)

	containmentQuery, err := json.Marshal(op.PkgQuery)
	if err != nil {
		return nil, errors.New("marshaling op.PkgQuery")
	}

	rows, err := globalGraphDBH.Db.Query(`
		SELECT *
		FROM pkgs
		WHERE lang=$1
		AND pkg @> $2
		LIMIT $3`, op.Lang, string(containmentQuery), op.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	var rawPkgs []sourcegraph.PackageInfo
	for rows.Next() {
		var (
			pkg, lang string
			repoID    int32
		)
		if err := rows.Scan(&repoID, &lang, &pkg); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		r := sourcegraph.PackageInfo{
			RepoID: repoID,
			Lang:   lang,
		}
		if err := json.Unmarshal([]byte(pkg), &r.Pkg); err != nil {
			return nil, errors.Wrap(err, "unmarshaling xdependencies metadata from sql scan")
		}
		rawPkgs = append(rawPkgs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}

	if len(rawPkgs) > 0 {
		pkgs = make([]sourcegraph.PackageInfo, 0, len(rawPkgs))
	}
	for _, pkg := range rawPkgs {
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "dpkgs.ListPackages", pkg.RepoID); err == nil {
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}
