package localstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

// globalDeps provides access to the `global_dep` table. Each row in
// the table represents a dependency relationship from a repository to
// a package-manager-level package.
//
// * The language column is the programming language in which the
//   dependency occurs (the language of the repository and the package
//   manager package)
// * The dep_data column contains JSON describing the package manager package.
//   Typically, this includes a name and version field.
// * The repo_id column identifies the repository.
// * The hints column contains JSON that contains additional hints that can
//   be used to optimized requests related to the dependency (e.g., which
//   directory in a repository contains the dependency).
//
// `global_dep_private` is an almost identical table except it is exclusively
// reserved for private repositories.
type globalDeps struct{}

var globalDepEnabledLangs = map[string]struct{}{
	"go":         struct{}{},
	"php":        struct{}{},
	"typescript": struct{}{},
	"java":       struct{}{},
}

func (g *globalDeps) CreateTable() string {
	return g.eachTable(`CREATE table $TABLE (
		language text NOT NULL,
		dep_data jsonb NOT NULL,
		repo_id integer NOT NULL,
		hints jsonb
	);
	CREATE INDEX $TABLE_idxgin ON $TABLE USING gin (dep_data jsonb_path_ops);
	CREATE INDEX $TABLE_repo_id ON $TABLE USING btree (repo_id);
	CREATE INDEX $TABLE_language ON $TABLE USING btree (language);`)
}

func (g *globalDeps) DropTable() string {
	return g.eachTable(`DROP TABLE IF EXISTS $TABLE CASCADE;`)
}

// eachTable appends the sql with "$TABLE" replaced by "global_dep" and
// "global_dep_private", and a newline separating the SQL lines. The composed
// SQL query is returned. It is obviously required that the input SQL end with
// a proper semicolon.
func (*globalDeps) eachTable(sql string) (composed string) {
	for _, table := range []string{"global_dep", "global_dep_private"} {
		composed += strings.Replace(sql, "$TABLE", table, -1) + "\n"
	}
	return
}

// UnsafeRefreshIndex refreshes the global deps index for the specified repo@commit.
//
// SECURITY: It is the caller's responsibility to ensure the repository
// described by the op parameter is accurately specified as private or not.
func (g *globalDeps) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp, langs []*inventory.Lang) error {
	var errs []string
	for _, lang := range langs {
		langName := strings.ToLower(lang.Name)

		if _, enabled := globalDepEnabledLangs[langName]; !enabled {
			continue
		}
		if err := g.refreshIndexForLanguage(ctx, langName, op); err != nil {
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

func (g *globalDeps) TotalRefs(ctx context.Context, source string) (int, error) {
	// SECURITY: Note that we do not speak to global_dep_private here, because
	// that could hint towards private repositories existing. We may decide to
	// relax this constraint in the future, but we should be extremely careful
	// in doing so.

	// Because global_dep only store Go package paths, not repository URIs, we
	// use a simple heuristic here by using `LIKE <repo>%`. This will work for
	// GitHub package paths (e.g. `github.com/a/b%` matches `github.com/a/b/c`)
	// but not custom import paths etc.
	rows, err := globalAppDBH.Db.Query(`SELECT COUNT(repo_id)
		FROM global_dep
		WHERE language='go'
		AND dep_data->>'depth' = '0'
		AND dep_data->>'package' LIKE $1;
	`, source+"%")
	if err != nil {
		return 0, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, errors.Wrap(err, "Scan")
		}
	}
	return count, nil
}

func (g *globalDeps) refreshIndexForLanguage(ctx context.Context, language string, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "refreshIndexForLanguage "+language)
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
	var deps []lspext.DependencyReference
	err = unsafeXLangCall(ctx, language+"_bg", rootPath, "workspace/xdependencies", map[string]string{}, &deps)
	if err != nil {
		return errors.Wrap(err, "LSP Call workspace/xdependencies")
	}

	table := "global_dep"
	if op.Private {
		table = "global_dep_private"
	}

	err = dbutil.Transaction(ctx, globalAppDBH.Db, func(tx *sql.Tx) error {
		// Update the table.
		err = g.update(ctx, tx, table, language, deps, op.RepoID)
		if err != nil {
			return errors.Wrap(err, "update "+table)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "executing transaction")
	}
	return nil
}

// DependenciesOptions specifies options for querying locations that reference
// a definition.
type DependenciesOptions struct {
	// Language is the type of language whose references are being queried.
	// e.g. "go" or "java".
	Language string

	// DepData is data that matches the output of xdependencies with a psql
	// jsonb containment operator. It may be a subset of data.
	DepData map[string]interface{}

	// Limit limits the number of returned dependency references to the
	// specified number.
	Limit int
}

func (g *globalDeps) Dependencies(ctx context.Context, op DependenciesOptions) (refs []*sourcegraph.DependencyReference, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "localstore.Dependencies")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Language", op.Language)
	span.SetTag("DepData", op.DepData)

	containmentQuery, err := json.Marshal(op.DepData)
	if err != nil {
		return nil, errors.New("marshaling op.DepData query")
	}

	rows, err := globalAppDBH.Db.Query(`select dep_data,repo_id,hints
		FROM global_dep
		WHERE language=$1
		AND dep_data @> $2
		LIMIT $3
	`, op.Language, string(containmentQuery), op.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			depData, hints string
			repoID         int32
		)
		if err := rows.Scan(&depData, &repoID, &hints); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		r := &sourcegraph.DependencyReference{
			RepoID: repoID,
		}
		if err := json.Unmarshal([]byte(depData), &r.DepData); err != nil {
			return nil, errors.Wrap(err, "unmarshaling xdependencies metadata from sql scan")
		}
		if err := json.Unmarshal([]byte(hints), &r.Hints); err != nil {
			return nil, errors.Wrap(err, "unmarshaling xdependencies hints from sql scan")
		}
		refs = append(refs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return refs, nil
}

// updateGlobalDep updates the global_dep table.
func (g *globalDeps) update(ctx context.Context, tx *sql.Tx, table, language string, deps []lspext.DependencyReference, indexRepo int32) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateGlobalDep "+language)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("deps", len(deps))
	span.SetTag("table", table)

	// First, create a temporary table.
	_, err = tx.Exec(`CREATE TEMPORARY TABLE new_` + table + ` (
	    language text NOT NULL,
	    dep_data jsonb NOT NULL,
	    repo_id integer NOT NULL,
	    hints jsonb
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}
	span.LogEvent("created temp table")

	// Copy the new deps into the temporary table.
	copy, err := tx.Prepare(pq.CopyIn("new_"+table,
		"language",
		"dep_data",
		"repo_id",
		"hints",
	))
	if err != nil {
		return errors.Wrap(err, "prepare copy in")
	}
	defer copy.Close()
	span.LogEvent("prepared copy in")

	for _, r := range deps {
		data, err := json.Marshal(r.Attributes)
		if err != nil {
			return errors.Wrap(err, "marshaling xdependency metadata to JSON")
		}
		hintsData, err := json.Marshal(r.Hints)
		if err != nil {
			return errors.Wrap(err, "marshaling xdependency hints to JSON")
		}

		if _, err := copy.Exec(
			language,          // language
			string(data),      // dep_data
			indexRepo,         // repo_id
			string(hintsData), // hints
		); err != nil {
			return errors.Wrap(err, "executing ref copy")
		}
	}
	span.LogEvent("executed all dep copy")
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}
	span.LogEvent("executed copy")

	if _, err := tx.Exec(`DELETE FROM `+table+` WHERE language=$1 AND repo_id=$2`, language, indexRepo); err != nil {
		return errors.Wrap(err, "executing table deletion")
	}
	span.LogEvent("executed table deletion")

	// Insert from temporary table into the real table.
	_, err = tx.Exec(`INSERT INTO ` + table + `(
		language,
		dep_data,
		repo_id,
		hints
	) SELECT d.language,
		d.dep_data,
		d.repo_id,
		d.hints
	FROM new_` + table + ` d;`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	span.LogEvent("executed final insertion from temp table")
	return nil
}
