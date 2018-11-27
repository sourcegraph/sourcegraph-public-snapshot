package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/inventory"
	"github.com/sourcegraph/sourcegraph/xlang"
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
// For a detailed overview of the schema, see schema.txt.
type globalDeps struct{}

func (g *globalDeps) TotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) (int, error) {
	var sum int
	for _, lang := range langs {
		switch lang.Name {
		case inventory.LangGo:
			for _, expandedSources := range repoNameToGoPathPrefixes(repo.Name) {
				refs, err := g.doTotalRefsGo(ctx, expandedSources)
				if err != nil {
					return 0, errors.Wrap(err, "doTotalRefsGo")
				}
				sum += refs
			}
		case inventory.LangJava:
			refs, err := g.doTotalRefs(ctx, repo.ID, "java")
			if err != nil {
				return 0, errors.Wrap(err, "doTotalRefs")
			}
			sum += refs
		}
	}
	return sum, nil
}

// ListTotalRefs is like TotalRefs, except it returns a list of repo IDs
// instead of just the length of that list. Obviously, this is less efficient
// if you just need the count, however.
func (g *globalDeps) ListTotalRefs(ctx context.Context, repo *types.Repo, langs []*inventory.Lang) ([]api.RepoID, error) {
	var repos []api.RepoID
	for _, lang := range langs {
		switch lang.Name {
		case inventory.LangGo:
			for _, expandedSources := range repoNameToGoPathPrefixes(repo.Name) {
				refs, err := g.doListTotalRefsGo(ctx, expandedSources)
				if err != nil {
					return nil, errors.Wrap(err, "doListTotalRefsGo")
				}
				repos = append(repos, refs...)
			}
		case inventory.LangJava:
			refs, err := g.doListTotalRefs(ctx, repo.ID, "java")
			if err != nil {
				return nil, errors.Wrap(err, "doListTotalRefs")
			}
			repos = append(repos, refs...)
		}
	}
	return repos, nil
}

// repoNameToGoPathPrefixes translates a repository name like
// github.com/kubernetes/kubernetes into its _prefix_ matching Go import paths
// (e.g. k8s.io/kubernetes). In the case of the standard library,
// github.com/golang/go returns all of the Go stdlib package paths. If the
// repository name is not special cased, []string{repoName} is simply returned.
//
// TODO(slimsag): In the future, when the pkgs index includes Go repositories,
// use that instead of this manual mapping hack.
func repoNameToGoPathPrefixes(repoName api.RepoName) []string {
	manualMapping := map[api.RepoName][]string{
		// stdlib hack: by returning an empty string (NOT no strings) we end up
		// with an SQL query like `AND dep_data->>'package' LIKE '%';` which
		// matches all Go repositories effectively. We do this for the stdlib
		// because all Go repositories will import the stdlib anyway.
		"github.com/golang/go": {""},

		// google.golang.org
		"github.com/grpc/grpc-go":                {"google.golang.org/grpc"},
		"github.com/google/google-api-go-client": {"google.golang.org/api"},
		"github.com/golang/appengine":            {"google.golang.org/appengine"},

		// go4.org
		"github.com/camlistore/go4": {"go4.org"},

		// At special request of a user, since we don't support custom import
		// paths generically here yet. See https://github.com/sourcegraph/sourcegraph/issues/12488
		"github.com/goadesign/goa": {"github.com/goadesign/goa", "goa.design/goa"},
	}
	if v, ok := manualMapping[repoName]; ok {
		return v
	}

	switch {
	case strings.HasPrefix(string(repoName), "github.com/azul3d"): // azul3d.org
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			return []string{"azul3d.org/" + split[2]}
		}

	case strings.HasPrefix(string(repoName), "github.com/dskinner"): // dasa.cc
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			return []string{"dasa.cc/" + split[2]}
		}

	case strings.HasPrefix(string(repoName), "github.com/kubernetes"): // k8s.io
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			return []string{"k8s.io/" + split[2]}
		}

	case strings.HasPrefix(string(repoName), "github.com/uber-go"): // go.uber.org
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			// Some repos use non-canonical import paths.
			return []string{
				string(repoName),
				"go.uber.org/" + split[2],
			}
		}

	case strings.HasPrefix(string(repoName), "github.com/dominikh"): // honnef.co
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			return []string{"honnef.co/" + strings.Replace(split[2], "-", "/", -1)}
		}

	case strings.HasPrefix(string(repoName), "github.com/golang") && repoName != "github.com/golang/go": // golang.org/x
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 {
			return []string{"golang.org/x/" + split[2]}
		}

	case strings.HasPrefix(string(repoName), "github.com"): // gopkg.in
		split := strings.Split(string(repoName), "/")
		if len(split) >= 3 && strings.HasPrefix(split[1], "go-") {
			// Four possibilities
			return []string{
				string(repoName), // github.com/go-foo/foo
				"gopkg.in/" + strings.TrimPrefix(split[1], "go-"),     // gopkg.in/foo
				"labix.org/v1/" + strings.TrimPrefix(split[1], "go-"), // labix.org/v1/foo
				"labix.org/v2/" + strings.TrimPrefix(split[1], "go-"), // labix.org/v2/foo
			}
		} else if len(split) >= 3 {
			// Two possibilities
			return []string{
				string(repoName),                        // github.com/foo/bar
				"gopkg.in/" + split[1] + "/" + split[2], // gopkg.in/foo/bar
			}
		}
	}
	return []string{string(repoName)}
}

// doTotalRefs is the generic implementation of total references, using the `pkgs` table.
func (g *globalDeps) doTotalRefs(ctx context.Context, repo api.RepoID, lang string) (sum int, err error) {
	// Get packages contained in the repo
	packages, err := (&pkgs{}).ListPackages(ctx, &api.ListPackagesOp{Lang: lang, Limit: 500, RepoID: repo})
	if err != nil {
		return 0, errors.Wrap(err, "ListPackages")
	}
	if len(packages) == 0 {
		return 0, nil
	}

	// Find number of repos that depend on that set of packages
	var args []interface{}
	arg := func(a interface{}) string {
		args = append(args, a)
		return fmt.Sprintf("$%d", len(args))
	}
	var pkgClauses []string
	for _, pkg := range packages {
		pkgID, ok := xlang.PackageIdentifier(pkg.Pkg, lang)
		if !ok {
			return 0, errors.Wrap(err, "PackageIdentifier")
		}
		containmentQuery, err := json.Marshal(pkgID)
		if err != nil {
			return 0, errors.Wrap(err, "Marshal")
		}
		pkgClauses = append(pkgClauses, `dep_data @> `+arg(string(containmentQuery)))
	}
	whereSQL := `(language=` + arg(lang) + `) AND ((` + strings.Join(pkgClauses, ") OR (") + `))`
	sql := `SELECT count(distinct(repo_id))
			FROM global_dep
			WHERE ` + whereSQL
	rows, err := dbconn.Global.QueryContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, errors.Wrap(err, "Scan")
		}
	}
	return count, nil
}

// doListTotalRefs is the generic implementation of list total references,
// using the `pkgs` table.
func (g *globalDeps) doListTotalRefs(ctx context.Context, repo api.RepoID, lang string) ([]api.RepoID, error) {
	// Get packages contained in the repo
	packages, err := (&pkgs{}).ListPackages(ctx, &api.ListPackagesOp{Lang: lang, Limit: 500, RepoID: repo})
	if err != nil {
		return nil, errors.Wrap(err, "ListPackages")
	}
	if len(packages) == 0 {
		return nil, nil
	}

	// Find all repos that depend on that set of packages
	var args []interface{}
	arg := func(a interface{}) string {
		args = append(args, a)
		return fmt.Sprintf("$%d", len(args))
	}
	var pkgClauses []string
	for _, pkg := range packages {
		pkgID, ok := xlang.PackageIdentifier(pkg.Pkg, lang)
		if !ok {
			return nil, errors.Wrap(err, "PackageIdentifier")
		}
		containmentQuery, err := json.Marshal(pkgID)
		if err != nil {
			return nil, errors.Wrap(err, "Marshal")
		}
		pkgClauses = append(pkgClauses, `dep_data @> `+arg(string(containmentQuery)))
	}
	whereSQL := `(language=` + arg(lang) + `) AND ((` + strings.Join(pkgClauses, ") OR (") + `))`
	sql := `SELECT distinct(repo_id)
			FROM global_dep
			WHERE ` + whereSQL
	rows, err := dbconn.Global.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var repos []api.RepoID
	for rows.Next() {
		var repo api.RepoID
		if err := rows.Scan(&repo); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// doTotalRefsGo is the Go-specific implementation of total references, since we can extract package metadata directly
// from Go repository URLs, without going through the `pkgs` table.
func (g *globalDeps) doTotalRefsGo(ctx context.Context, source string) (int, error) {
	// Because global_dep only stores Go package paths, not repository names, we
	// use a simple heuristic here by using `LIKE <repo>%`. This will work for
	// GitHub package paths (e.g. `github.com/a/b%` matches `github.com/a/b/c`)
	// but not custom import paths etc.
	rows, err := dbconn.Global.QueryContext(ctx, `SELECT COUNT(DISTINCT repo_id)
		FROM global_dep
		WHERE language='go'
		AND dep_data->>'depth' = '0'
		AND ( -- in C locale, this is equivalent to matching "$1/*", but matches much faster
			(dep_data->>'package' COLLATE "C" < $1 || '0' COLLATE "C" AND dep_data->>'package' COLLATE "C" > $1 || '/' COLLATE "C")
			OR (dep_data->>'package' COLLATE "C" = $1)
		);
	`, source)
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

// doListTotalRefsGo is the Go-specific implementation of list total
// references, since we can extract package metadata directly from Go
// repository URLs, without going through the `pkgs` table.
func (g *globalDeps) doListTotalRefsGo(ctx context.Context, source string) ([]api.RepoID, error) {
	// Because global_dep only stores Go package paths, not repository names, we
	// use a simple heuristic here by using `LIKE <repo>%`. This will work for
	// GitHub package paths (e.g. `github.com/a/b%` matches `github.com/a/b/c`)
	// but not custom import paths etc.
	rows, err := dbconn.Global.QueryContext(ctx, `SELECT DISTINCT repo_id
		FROM global_dep
		WHERE language='go'
		AND dep_data->>'depth' = '0'
		AND dep_data->>'package' LIKE $1;
	`, source+"%")
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var repos []api.RepoID
	for rows.Next() {
		var repo api.RepoID
		err := rows.Scan(&repo)
		if err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (g *globalDeps) UpdateIndexForLanguage(ctx context.Context, language string, repo api.RepoID, deps []lspext.DependencyReference) (err error) {
	err = db.Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		// Update the table.
		err = g.update(ctx, tx, language, deps, repo)
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

func (g *globalDeps) Dependencies(ctx context.Context, op db.DependenciesOptions) (refs []*api.DependencyReference, err error) {
	if db.Mocks.GlobalDeps.Dependencies != nil {
		return db.Mocks.GlobalDeps.Dependencies(ctx, op)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "db.Dependencies")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Language", op.Language)
	span.SetTag("DepData", op.DepData)

	var args []interface{}
	arg := func(a interface{}) string {
		args = append(args, a)
		return fmt.Sprintf("$%d", len(args))
	}

	var whereConds []string

	if op.Language != "" {
		whereConds = append(whereConds, `gd.language=`+arg(op.Language))
	}

	if op.DepData != nil {
		containmentQuery, err := json.Marshal(op.DepData)
		if err != nil {
			return nil, errors.New("marshaling op.DepData query")
		}
		whereConds = append(whereConds, `dep_data @> `+arg(string(containmentQuery)))
	}
	if op.Repo != 0 {
		whereConds = append(whereConds, `repo_id = `+arg(op.Repo))
	}

	selectSQL := `SELECT gd.language, dep_data, repo_id, hints`
	fromSQL := `FROM global_dep AS gd INNER JOIN repo AS r ON gd.repo_id=r.id`
	whereSQL := ""
	if len(whereConds) > 0 {
		whereSQL = `WHERE ` + strings.Join(whereConds, " AND ")
	}
	limitSQL := ""
	if op.Limit != 0 {
		limitSQL = `LIMIT ` + arg(op.Limit)
	}
	sql := fmt.Sprintf("%s %s %s %s", selectSQL, fromSQL, whereSQL, limitSQL)

	rows, err := dbconn.Global.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			language, depData, hints string
			repo                     api.RepoID
		)
		if err := rows.Scan(&language, &depData, &repo, &hints); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		r := &api.DependencyReference{
			RepoID:   repo,
			Language: language,
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
func (g *globalDeps) update(ctx context.Context, tx *sql.Tx, language string, deps []lspext.DependencyReference, indexRepo api.RepoID) (err error) {
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
	_, err = tx.ExecContext(ctx, `CREATE TEMPORARY TABLE new_global_dep (
	    language text NOT NULL,
	    dep_data jsonb NOT NULL,
	    repo_id integer NOT NULL,
	    hints jsonb
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}
	span.LogFields(otlog.String("event", "created temp table"))

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
	span.LogFields(otlog.String("event", "prepared copy in"))

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
	span.LogFields(otlog.String("event", "executed all dep copy"))
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}
	span.LogFields(otlog.String("event", "executed copy"))

	if _, err := tx.ExecContext(ctx, `DELETE FROM global_dep WHERE language=$1 AND repo_id=$2`, language, indexRepo); err != nil {
		return errors.Wrap(err, "executing table deletion")
	}
	span.LogFields(otlog.String("event", "executed table deletion"))

	// Insert from temporary table into the real table.
	_, err = tx.ExecContext(ctx, `INSERT INTO global_dep(
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
	span.LogFields(otlog.String("event", "executed final insertion from temp table"))
	return nil
}

func (g *globalDeps) Delete(ctx context.Context, repo api.RepoID) error {
	_, err := dbconn.Global.ExecContext(ctx, `DELETE FROM global_dep WHERE repo_id=$1`, repo)
	return err
}
