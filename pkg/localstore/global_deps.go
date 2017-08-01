package localstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	gogithub "github.com/sourcegraph/go-github/github"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
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
// `global_dep_private` is an identical table, except that instead of only
// storing public repository data (like `global_dep` does), it only stores
// private repository data. It includes all dependencies, public or private,
// for all private repositories.
type globalDeps struct{}

var globalDepEnabledLangs = map[string]struct{}{
	"go":         struct{}{},
	"php":        struct{}{},
	"typescript": struct{}{},
	"java":       struct{}{},
	"python":     struct{}{},
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

// RefreshIndex refreshes the global deps index for the specified repo@commit.
func (g *globalDeps) RefreshIndex(ctx context.Context, repoURI, commitID string, reposGetInventory func(context.Context, *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)) error {
	// 🚨 SECURITY: Do not remove this call. It prevents us from leaking 🚨
	// whether or not a private repo exists based on measuring the time
	// RefreshIndex takes.
	repo, err := Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return errors.Wrap(err, "Repos.GetByURI")
	}
	inv, err := reposGetInventory(ctx, &sourcegraph.RepoRevSpec{Repo: repo.ID, CommitID: commitID})
	if err != nil {
		return errors.Wrap(err, "Repos.GetInventory")
	}

	var errs []string
	for _, lang := range inv.Languages {
		langName := strings.ToLower(lang.Name)

		if _, enabled := globalDepEnabledLangs[langName]; !enabled {
			continue
		}
		if err := g.refreshIndexForLanguage(ctx, langName, repo, commitID); err != nil {
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

func (g *globalDeps) TotalRefs(ctx context.Context, repoURI string, langs []*inventory.Lang) (int, error) {
	repo, err := Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return 0, errors.Wrap(err, "Repos.GetByURI")
	}

	var sum int
	for _, lang := range langs {
		switch lang.Name {
		case inventory.LangGo:
			for _, expandedSources := range repoURIToGoPathPrefixes(repoURI) {
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
func (g *globalDeps) ListTotalRefs(ctx context.Context, repoURI string, langs []*inventory.Lang) ([]int32, error) {
	repo, err := Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return nil, errors.Wrap(err, "Repos.GetByURI")
	}

	var results []int32
	for _, lang := range langs {
		switch lang.Name {
		case inventory.LangGo:
			for _, expandedSources := range repoURIToGoPathPrefixes(repoURI) {
				refs, err := g.doListTotalRefsGo(ctx, expandedSources)
				if err != nil {
					return nil, errors.Wrap(err, "doListTotalRefsGo")
				}
				results = append(results, refs...)
			}
		case inventory.LangJava:
			refs, err := g.doListTotalRefs(ctx, repo.ID, "java")
			if err != nil {
				return nil, errors.Wrap(err, "doListTotalRefs")
			}
			results = append(results, refs...)
		}
	}
	return results, nil
}

// repoURIToGoPathPrefixes translates a repository URI like
// github.com/kubernetes/kubernetes into its _prefix_ matching Go import paths
// (e.g. k8s.io/kubernetes). In the case of the standard library,
// github.com/golang/go returns all of the Go stdlib package paths. If the
// repository URI is not special cased, []string{repoURI} is simply returned.
//
// TODO(slimsag): In the future, when the pkgs index includes Go repositories,
// use that instead of this manual mapping hack.
func repoURIToGoPathPrefixes(repoURI string) []string {
	manualMapping := map[string][]string{
		// stdlib hack: by returning an empty string (NOT no strings) we end up
		// with an SQL query like `AND dep_data->>'package' LIKE '%';` which
		// matches all Go repositories effectively. We do this for the stdlib
		// because all Go repositories will import the stdlib anyway.
		"github.com/golang/go": []string{""},

		// google.golang.org
		"github.com/grpc/grpc-go":                []string{"google.golang.org/grpc"},
		"github.com/google/google-api-go-client": []string{"google.golang.org/api"},
		"github.com/golang/appengine":            []string{"google.golang.org/appengine"},

		// go4.org
		"github.com/camlistore/go4": []string{"go4.org"},
	}
	if v, ok := manualMapping[repoURI]; ok {
		return v
	}

	switch {
	case strings.HasPrefix(repoURI, "github.com/azul3d"): // azul3d.org
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			return []string{"azul3d.org/" + split[2]}
		}

	case strings.HasPrefix(repoURI, "github.com/dskinner"): // dasa.cc
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			return []string{"dasa.cc/" + split[2]}
		}

	case strings.HasPrefix(repoURI, "github.com/kubernetes"): // k8s.io
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			return []string{"k8s.io/" + split[2]}
		}

	case strings.HasPrefix(repoURI, "github.com/uber-go"): // go.uber.org
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			// Uber also uses their non-canonical import paths for some repos.
			return []string{
				repoURI,
				"go.uber.org/" + split[2],
			}
		}

	case strings.HasPrefix(repoURI, "github.com/dominikh"): // honnef.co
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			return []string{"honnef.co/" + strings.Replace(split[2], "-", "/", -1)}
		}

	case strings.HasPrefix(repoURI, "github.com/golang") && repoURI != "github.com/golang/go": // golang.org/x
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 {
			return []string{"golang.org/x/" + split[2]}
		}

	case strings.HasPrefix(repoURI, "github.com"): // gopkg.in
		split := strings.Split(repoURI, "/")
		if len(split) >= 3 && strings.HasPrefix(split[1], "go-") {
			// Four possibilities
			return []string{
				repoURI, // github.com/go-foo/foo
				"gopkg.in/" + strings.TrimPrefix(split[1], "go-"),     // gopkg.in/foo
				"labix.org/v1/" + strings.TrimPrefix(split[1], "go-"), // labix.org/v1/foo
				"labix.org/v2/" + strings.TrimPrefix(split[1], "go-"), // labix.org/v2/foo
			}
		} else if len(split) >= 3 {
			// Two possibilities
			return []string{
				repoURI, // github.com/foo/bar
				"gopkg.in/" + split[1] + "/" + split[2], // gopkg.in/foo/bar
			}
		}
	}
	return []string{repoURI}
}

// doTotalRefs is the generic implementation of total references, using the `pkgs` table.
func (g *globalDeps) doTotalRefs(ctx context.Context, repo int32, lang string) (sum int, err error) {
	// Get packages contained in the repo
	packages, err := (&pkgs{}).ListPackages(ctx, &sourcegraph.ListPackagesOp{Lang: lang, Limit: 500, RepoID: repo})
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
	rows, err := appDBH(ctx).Db.Query(sql, args...)
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
func (g *globalDeps) doListTotalRefs(ctx context.Context, repo int32, lang string) (results []int32, err error) {
	// Get packages contained in the repo
	packages, err := (&pkgs{}).ListPackages(ctx, &sourcegraph.ListPackagesOp{Lang: lang, Limit: 500, RepoID: repo})
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
	rows, err := appDBH(ctx).Db.Query(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		results = append(results, int32(id))
	}
	return results, nil
}

// doTotalRefsGo is the Go-specific implementation of total references, since we can extract package metadata directly
// from Go repository URLs, without going through the `pkgs` table.
func (g *globalDeps) doTotalRefsGo(ctx context.Context, source string) (int, error) {
	// 🚨 SECURITY: Note that we do not speak to global_dep_private here, because 🚨
	// that could hint towards private repositories existing. We may decide to
	// relax this constraint in the future, but we should be extremely careful
	// in doing so.

	// Because global_dep only store Go package paths, not repository URIs, we
	// use a simple heuristic here by using `LIKE <repo>%`. This will work for
	// GitHub package paths (e.g. `github.com/a/b%` matches `github.com/a/b/c`)
	// but not custom import paths etc.
	rows, err := appDBH(ctx).Db.Query(`SELECT COUNT(DISTINCT repo_id)
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

// doListTotalRefsGo is the Go-specific implementation of list total
// references, since we can extract package metadata directly from Go
// repository URLs, without going through the `pkgs` table.
func (g *globalDeps) doListTotalRefsGo(ctx context.Context, source string) ([]int32, error) {
	// 🚨 SECURITY: Note that we do not speak to global_dep_private here, because 🚨
	// that could hint towards private repositories existing. We may decide to
	// relax this constraint in the future, but we should be extremely careful
	// in doing so.

	// Because global_dep only store Go package paths, not repository URIs, we
	// use a simple heuristic here by using `LIKE <repo>%`. This will work for
	// GitHub package paths (e.g. `github.com/a/b%` matches `github.com/a/b/c`)
	// but not custom import paths etc.
	rows, err := appDBH(ctx).Db.Query(`SELECT DISTINCT repo_id
		FROM global_dep
		WHERE language='go'
		AND dep_data->>'depth' = '0'
		AND dep_data->>'package' LIKE $1;
	`, source+"%")
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var results []int32
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		results = append(results, int32(id))
	}
	return results, nil
}

func (g *globalDeps) refreshIndexForLanguage(ctx context.Context, language string, repo *sourcegraph.Repo, commitID string) (err error) {
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
	rootURI := lsp.DocumentURI(vcs + "://" + repo.URI + "?" + commitID)
	var deps []lspext.DependencyReference
	err = unsafeXLangCall(ctx, language+"_bg", rootURI, "workspace/xdependencies", map[string]string{}, &deps)
	if err != nil {
		return errors.Wrap(err, "LSP Call workspace/xdependencies")
	}

	table := "global_dep"
	if repo.Private {
		table = "global_dep_private"
	}

	err = dbutil.Transaction(ctx, appDBH(ctx).Db, func(tx *sql.Tx) error {
		// Update the table.
		err = g.update(ctx, tx, table, language, deps, repo.ID)
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

	// Repo filters the returned list of DependencyReference instances
	// by repo. It should be used mutually exclusively with DepData.
	Repo int32

	// ExcludePrivate excludes private repo IDs from being included in the result set
	ExcludePrivate bool

	// Limit limits the number of returned dependency references to the
	// specified number.
	Limit int
}

var mockListUserPrivateRepoIDs func(ctx context.Context) ([]int32, error)

// listUserPrivateRepoIDs lists all of the private repository IDs that the user
// in ctx has access to.
//
// 🚨 SECURITY: This function MUST return ONLY the private repositories accessible 🚨
// by the user in ctx. Doing anything otherwise would introduce security holes.
func listUserPrivateRepoIDs(ctx context.Context) (accessible []int32, err error) {
	if mockListUserPrivateRepoIDs != nil {
		return mockListUserPrivateRepoIDs(ctx)
	}
	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "listUserPrivateRepoIDs")
	ghRepos, err := github.ListAllGitHubRepos(ctx, &gogithub.RepositoryListOptions{Visibility: "private"})
	if err != nil {
		return nil, err
	}
	for _, r := range ghRepos {
		// Because r describes a remote repository, it has no valid ID field.
		// We must fetch it from the DB.
		r, err = Repos.GetByURI(ctx, r.URI)
		if err != nil {
			if legacyerr.ErrCode(err) == legacyerr.NotFound {
				continue // ignore repos that are not yet cloned
			}
			return nil, err
		}
		accessible = append(accessible, r.ID)
	}
	return accessible, nil
}

func (g *globalDeps) Dependencies(ctx context.Context, op DependenciesOptions) (refs []*sourcegraph.DependencyReference, err error) {
	var privateRepoIDs []int32
	if !op.ExcludePrivate {
		privateRepoIDs, err = listUserPrivateRepoIDs(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Note: using global_dep_private first so those results always show up
	// first, as the user will always be more interested in their private code.
	for _, table := range []string{"global_dep_private", "global_dep"} {
		v, err := g.queryDependencies(ctx, table, op, privateRepoIDs)
		if err != nil {
			return nil, err
		}
		refs = append(refs, v...)
	}

	// 🚨 SECURITY: Verify that the user has access to the resulting dependency 🚨
	// references. In general, this should not happen, but it can occur if e.g.
	// a repository was once public but is now private. We simply remove them
	// in that situation.
	finalRefs := make([]*sourcegraph.DependencyReference, 0, len(refs))
	for _, ref := range refs {
		if _, err := Repos.Get(ctx, ref.RepoID); err != nil {
			continue
		}
		finalRefs = append(finalRefs, ref)
	}
	return finalRefs, nil
}

// queryDependencies is invoked first for `global_dep_private` (private repos)
// and then for `global_dep` (public repos). See the globalDeps type docstring
// for more concrete information.
func (g *globalDeps) queryDependencies(ctx context.Context, table string, op DependenciesOptions, privateRepoIDs []int32) (refs []*sourcegraph.DependencyReference, err error) {
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
	span.SetTag("table", table)

	var args []interface{}
	arg := func(a interface{}) string {
		args = append(args, a)
		return fmt.Sprintf("$%d", len(args))
	}

	var whereConds []string

	if op.Language != "" {
		whereConds = append(whereConds, `language=`+arg(op.Language))
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

	switch table {
	case "global_dep_private":
		// Important: without this check we would produce a query like
		// `repo_id IN ()` which is illegal / a syntax error in SQL.
		if len(privateRepoIDs) == 0 {
			return nil, nil
		}
		var privateRepoStrings []string
		for _, repoID := range privateRepoIDs {
			privateRepoStrings = append(privateRepoStrings, strconv.Itoa(int(repoID)))
		}
		privateRepos := strings.Join(privateRepoStrings, ", ")
		whereConds = append(whereConds, `repo_id IN (`+privateRepos+`)`)
	case "global_dep":
	default:
		panic(fmt.Sprintf("Defs.Dependencies: unexpected table %q", table))
	}

	selectSQL := `SELECT dep_data, repo_id, hints`
	fromSQL := `FROM ` + table
	whereSQL := ""
	if len(whereConds) > 0 {
		whereSQL = `WHERE ` + strings.Join(whereConds, " AND ")
	}
	limitSQL := ""
	if op.Limit != 0 {
		limitSQL = `LIMIT ` + arg(op.Limit)
	}
	sql := fmt.Sprintf("%s %s %s %s", selectSQL, fromSQL, whereSQL, limitSQL)

	rows, err := appDBH(ctx).Db.Query(sql, args...)
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
