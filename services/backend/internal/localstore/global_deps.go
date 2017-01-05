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
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

// dbGlobalDep provides access to the `global_dep` table. Each row in
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
type dbGlobalDep struct{}

func (*dbGlobalDep) CreateTable() string {
	return `CREATE table global_dep (
		language text NOT NULL,
		dep_data jsonb NOT NULL,
		repo_id integer NOT NULL,
		hints jsonb
	);
	CREATE INDEX global_dep_idxgin ON global_dep USING gin (dep_data jsonb_path_ops);
	CREATE INDEX global_dep_repo_id ON global_dep USING btree (repo_id);
	CREATE INDEX global_dep_language ON global_dep USING btree (language);`
}

func (*dbGlobalDep) DropTable() string {
	return `DROP TABLE IF EXISTS global_dep CASCADE;`
}

type globalDeps struct{}

// UnsafeRefreshIndex refreshes the global deps index for the specified repo@commit.
//
// SECURITY: It is the caller's responsibility to ensure the repository is NOT
// a private one.
func (g *globalDeps) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error {
	var errs []string
	for _, language := range []string{"go"} { // TODO(slimsag): use inventory instead
		if err := g.refreshIndexForLanguage(ctx, language, op); err != nil {
			log15.Crit("refreshing index failed", "language", language, "error", err)
			errs = append(errs, fmt.Sprintf("refreshing index failed language=%s error=%v", language, err))
		}
	}
	if len(errs) == 1 {
		return errors.New(errs[0])
	} else if len(errs) > 1 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
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
	err = xlang.UnsafeOneShotClientRequest(ctx, language+"_bg", rootPath, "workspace/xdependencies", map[string]string{}, &deps)
	if err != nil {
		return errors.Wrap(err, "LSP Call workspace/xdependencies")
	}

	err = dbutil.Transaction(ctx, globalGraphDBH.Db, func(tx *sql.Tx) error {
		// Update the global_dep table.
		err = g.updateGlobalDep(ctx, tx, language, deps, op.RepoID)
		if err != nil {
			return errors.Wrap(err, "update global_dep")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "executing transaction")
	}
	return nil
}

// RefLocationsOptions specifies options for querying locations that reference
// a definition.
type RefLocationsOptions struct {
	// Language is the type of language whose references are being queried.
	// e.g. "go" or "java".
	Language string

	// DepData is data that matches the output of xdependencies with a psql
	// jsonb containment operator. It may be a subset of data.
	DepData map[string]interface{}
}

type Dependency struct {
	DepData map[string]interface{}
	RepoID  int32
	Hints   map[string]interface{}
}

func (g *globalDeps) Dependencies(ctx context.Context, op RefLocationsOptions) (refs []*Dependency, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "localstore.RefLocations")
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

	rows, err := globalGraphDBH.Db.Query(`select distinct on (repo_id) dep_data,repo_id,hints
		FROM global_dep
		WHERE language=$1
		AND dep_data @> $2
		LIMIT 4
	`, op.Language, string(containmentQuery))
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
		r := &Dependency{
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
func (g *globalDeps) updateGlobalDep(ctx context.Context, tx *sql.Tx, language string, deps []lspext.DependencyReference, indexRepo int32) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateGlobalDep "+language)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("deps", len(deps))

	// First, create a temporary table.
	_, err = tx.Exec(`CREATE TEMPORARY TABLE new_global_dep (
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
	copy, err := tx.Prepare(pq.CopyIn("new_global_dep",
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
			indexRepo,         // ref_repo_id
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

	if _, err := tx.Exec(`DELETE FROM global_dep WHERE language=$1 AND repo_id=$2`, language, indexRepo); err != nil {
		return errors.Wrap(err, "executing global_dep deletion")
	}
	span.LogEvent("executed global_dep deletion")

	// Insert from temporary table into the real table.
	_, err = tx.Exec(`INSERT INTO global_dep(
		language,
		dep_data,
		repo_id,
		hints
	) SELECT d.language,
		d.dep_data,
		d.repo_id,
		d.hints
	FROM new_global_dep d;`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	span.LogEvent("executed final insertion from temp table")
	return nil
}
