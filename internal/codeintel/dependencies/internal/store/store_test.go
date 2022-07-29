package store

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestPreciseDependenciesAndDependents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	// Note: repo identifiers match the name due to insertion order
	for _, repo := range []string{"repo-1", "repo-2", "repo-3", "repo-4", "repo-5"} {
		if err := store.db.Exec(ctx, sqlf.Sprintf(`INSERT INTO repo (name) VALUES (%s)`, repo)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	for _, uploadSpec := range []struct {
		RepositoryID int
		Commit       string
	}{
		{1, "0000000000000000000000000000000000000001"}, // uploadID = 1
		{2, "0000000000000000000000000000000000000002"}, // uploadID = 2
		{3, "0000000000000000000000000000000000000003"}, // uploadID = 3
		{4, "0000000000000000000000000000000000000004"}, // uploadID = 4
		{5, "0000000000000000000000000000000000000005"}, // uploadID = 5
	} {
		if err := store.db.Exec(ctx, sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				repository_id, commit, state, indexer, num_parts, uploaded_parts
			) VALUES (
				%s, %s, 'COMPLETED', '', 1, '{}'
			)
		`,
			uploadSpec.RepositoryID,
			uploadSpec.Commit,
		)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	for _, dependencySpec := range []struct {
		pkgID int
		refID int
	}{
		{1, 2}, // 2 depends on 1
		{1, 3}, // 3 depends on 1
		{1, 4}, // 4 depends on 1
		{2, 4}, // 4 depends on 2
		{4, 3}, // 3 depends on 4
		{3, 5}, // 5 depends on 3
	} {
		if err := store.db.Exec(ctx, sqlf.Sprintf(`
			INSERT INTO lsif_packages (dump_id, scheme, name, version) VALUES (%s, 'A' || %s, 'B' || %s, 'C' || %s)
		`,
			dependencySpec.pkgID,
			strconv.Itoa(dependencySpec.pkgID),
			strconv.Itoa(dependencySpec.pkgID),
			strconv.Itoa(dependencySpec.pkgID),
		)); err != nil {
			t.Fatalf(err.Error())
		}

		if err := store.db.Exec(ctx, sqlf.Sprintf(`
			INSERT INTO lsif_references (dump_id, scheme, name, version) VALUES (%s, 'A' || %s, 'B' || %s, 'C' || %s)
		`,
			dependencySpec.refID,
			strconv.Itoa(dependencySpec.pkgID),
			strconv.Itoa(dependencySpec.pkgID),
			strconv.Itoa(dependencySpec.pkgID),
		)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	t.Run("dependencies", func(t *testing.T) {
		for _, expectedDependency := range []struct {
			repoName     string
			commit       string
			dependencies map[api.RepoName]types.RevSpecSet
		}{
			{"repo-1", "0000000000000000000000000000000000000001", map[api.RepoName]types.RevSpecSet{
				// empty
			}},
			{"repo-2", "0000000000000000000000000000000000000002", map[api.RepoName]types.RevSpecSet{
				"repo-1": {"0000000000000000000000000000000000000001": struct{}{}},
			}},
			{"repo-3", "0000000000000000000000000000000000000003", map[api.RepoName]types.RevSpecSet{
				"repo-1": {"0000000000000000000000000000000000000001": struct{}{}},
				"repo-4": {"0000000000000000000000000000000000000004": struct{}{}},
			}},
			{"repo-4", "0000000000000000000000000000000000000004", map[api.RepoName]types.RevSpecSet{
				"repo-1": {"0000000000000000000000000000000000000001": struct{}{}},
				"repo-2": {"0000000000000000000000000000000000000002": struct{}{}},
			}},
			{"repo-5", "0000000000000000000000000000000000000005", map[api.RepoName]types.RevSpecSet{
				"repo-3": {"0000000000000000000000000000000000000003": struct{}{}},
			}},
		} {
			dependencies, err := store.PreciseDependencies(ctx, expectedDependency.repoName, expectedDependency.commit)
			if err != nil {
				t.Fatalf(err.Error())
			}

			if diff := cmp.Diff(expectedDependency.dependencies, dependencies); diff != "" {
				t.Fatalf("unexpected dependencies (-have, +want): %s", diff)
			}
		}
	})

	t.Run("dependents", func(t *testing.T) {
		for _, expectedDependency := range []struct {
			repoName   string
			commit     string
			dependents map[api.RepoName]types.RevSpecSet
		}{
			{"repo-1", "0000000000000000000000000000000000000001", map[api.RepoName]types.RevSpecSet{
				"repo-2": {"0000000000000000000000000000000000000002": struct{}{}},
				"repo-3": {"0000000000000000000000000000000000000003": struct{}{}},
				"repo-4": {"0000000000000000000000000000000000000004": struct{}{}},
			}},
			{"repo-2", "0000000000000000000000000000000000000002", map[api.RepoName]types.RevSpecSet{
				"repo-4": {"0000000000000000000000000000000000000004": struct{}{}},
			}},
			{"repo-3", "0000000000000000000000000000000000000003", map[api.RepoName]types.RevSpecSet{
				"repo-5": {"0000000000000000000000000000000000000005": struct{}{}},
			}},
			{"repo-4", "0000000000000000000000000000000000000004", map[api.RepoName]types.RevSpecSet{
				"repo-3": {"0000000000000000000000000000000000000003": struct{}{}},
			}},
			{"repo-5", "0000000000000000000000000000000000000004", map[api.RepoName]types.RevSpecSet{
				// empty
			}},
		} {
			dependents, err := store.PreciseDependents(ctx, expectedDependency.repoName, expectedDependency.commit)
			if err != nil {
				t.Fatalf(err.Error())
			}

			if diff := cmp.Diff(expectedDependency.dependents, dependents); diff != "" {
				t.Fatalf("unexpected dependents (-have, +want): %s", diff)
			}
		}
	})
}

func TestUpsertLockfileGraph(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral("A", "1", "2", "pkg-A", "4")
	packageB := shared.TestPackageDependencyLiteral("B", "2", "3", "pkg-B", "5")
	packageC := shared.TestPackageDependencyLiteral("C", "3", "4", "pkg-C", "6")
	packageD := shared.TestPackageDependencyLiteral("D", "4", "5", "pkg-D", "7")
	packageE := shared.TestPackageDependencyLiteral("E", "5", "6", "pkg-E", "8")
	packageF := shared.TestPackageDependencyLiteral("F", "6", "7", "pkg-F", "9")

	t.Run("timestamps", func(t *testing.T) {
		now := timeutil.Now()

		deps := []shared.PackageDependency{packageA}
		graph := shared.TestDependencyGraphLiteral([]shared.PackageDependency{packageA}, false, [][]shared.PackageDependency{})
		commit := "cafebabe"

		if err := store.upsertLockfileGraphAt(ctx, "foo", commit, "lock.file", deps, graph, now); err != nil {
			t.Fatalf("error: %s", err)
		}

		index, err := store.GetLockfileIndex(ctx, GetLockfileIndexOpts{RepoName: "foo", Commit: commit, Lockfile: "lock.file"})
		if err != nil {
			t.Fatal(err)
		}

		if !index.CreatedAt.Equal(now) {
			t.Fatalf("createdAt not equal to timestamp passed in. want=%s, have=%s", now, index.CreatedAt)
		}

		if !index.UpdatedAt.Equal(now) {
			t.Fatalf("updatedAt not equal to timestamp passed in. want=%s, have=%s", now, index.UpdatedAt)
		}

		// upsert again, with different timestamp, to see that `updated_at` is updated
		now2 := now.Add(5 * time.Hour)

		if err := store.upsertLockfileGraphAt(ctx, "foo", commit, "lock.file", deps, graph, now2); err != nil {
			t.Fatalf("error: %s", err)
		}

		index, err = store.GetLockfileIndex(ctx, GetLockfileIndexOpts{RepoName: "foo", Commit: commit, Lockfile: "lock.file"})
		if err != nil {
			t.Fatal(err)
		}

		if !index.CreatedAt.Equal(now) {
			t.Fatalf("createdAt not equal to timestamp passed in. want=%s, have=%s", now, index.CreatedAt)
		}

		if !index.UpdatedAt.Equal(now2) {
			t.Fatalf("updatedAt not equal to timestamp passed in. want=%s, have=%s", now2, index.UpdatedAt)
		}
	})

	t.Run("with graph", func(t *testing.T) {
		deps := []shared.PackageDependency{packageA, packageB, packageC, packageD, packageE, packageF}
		//    -> b -> E
		//   /
		// a --> c
		//   \
		//    -> d -> F
		graph := shared.TestDependencyGraphLiteral(
			[]shared.PackageDependency{packageA},
			false,
			[][]shared.PackageDependency{
				// A
				{packageA, packageB},
				{packageA, packageC},
				{packageA, packageD},
				// B
				{packageB, packageE},
				// D
				{packageD, packageF},
			},
		)
		commit := "cafebabe"

		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, graph); err != nil {
			t.Fatalf("error: %s", err)
		}

		// Now check whether the direct dependency was inserted
		names, err := queryDirectDeps(t, ctx, store, "foo", commit, "lock.file")
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}

		wantNames := []string{string(packageA.PackageSyntax())}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}

		// Check that all packages have been inserted
		names, err = queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		wantNames = []string{}
		for _, pkg := range deps {
			wantNames = append(wantNames, string(pkg.PackageSyntax()))
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}

		// Upsert again to check idempotency
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, graph); err != nil {
			t.Fatalf("error: %s", err)
		}
		names, err = basestore.ScanStrings(store.db.Query(ctx, sqlf.Sprintf(`SELECT package_name FROM codeintel_lockfile_references ORDER BY package_name`)))
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}
	})

	t.Run("with graph and undeterminable roots", func(t *testing.T) {
		deps := []shared.PackageDependency{packageA, packageB, packageC, packageD, packageE, packageF}
		//    -> b -> E
		//   /
		// a --> c
		//   \
		//    -> d -> F
		graph := shared.TestDependencyGraphLiteral(
			[]shared.PackageDependency{packageA},
			true,
			[][]shared.PackageDependency{
				// A
				{packageA, packageB},
				{packageA, packageC},
				{packageA, packageD},
				// B
				{packageB, packageE},
				// D
				{packageD, packageF},
			},
		)
		commit := "cafebabe"

		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, graph); err != nil {
			t.Fatalf("error: %s", err)
		}

		// Now check whether the direct dependency was inserted
		names, err := queryDirectDeps(t, ctx, store, "foo", commit, "lock.file")
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}

		wantNames := []string{string(packageA.PackageSyntax())}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}

		// Check that all packages have been inserted
		names, err = queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		wantNames = []string{}
		for _, pkg := range deps {
			wantNames = append(wantNames, string(pkg.PackageSyntax()))
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}

		// Upsert again to check idempotency
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, graph); err != nil {
			t.Fatalf("error: %s", err)
		}
		names, err = basestore.ScanStrings(store.db.Query(ctx, sqlf.Sprintf(`SELECT package_name FROM codeintel_lockfile_references ORDER BY package_name`)))
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}
	})

	t.Run("without graph", func(t *testing.T) {
		deps := []shared.PackageDependency{packageA, packageB, packageC, packageD, packageE, packageF}
		nilGraph := shared.SerializeDependencyGraph(nil)
		commit := "d34df00d"

		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, nilGraph); err != nil {
			t.Fatalf("error: %s", err)
		}

		// Check that all dependencies have been inserted as direct dependencies
		names, err := queryDirectDeps(t, ctx, store, "foo", commit, "lock.file")
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}

		wantNames := make([]string, 0, len(deps))
		for _, d := range deps {
			wantNames = append(wantNames, string(d.PackageSyntax()))
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected direct dependencies (-want +got):\n%s", diff)
		}

		// Upsert again to check idempotency
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, nilGraph); err != nil {
			t.Fatalf("error: %s", err)
		}
		names, err = queryDirectDeps(t, ctx, store, "foo", commit, "lock.file")
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}
	})

	t.Run("multiple lockfiles", func(t *testing.T) {
		results := []struct {
			lockfile string
			deps     []shared.PackageDependency
			graph    shared.DependencyGraph
		}{
			{
				lockfile: "lock1.file",
				deps:     []shared.PackageDependency{packageA, packageB, packageC},
				graph: shared.TestDependencyGraphLiteral(
					[]shared.PackageDependency{packageA},
					false,
					// a -> b -> c
					[][]shared.PackageDependency{{packageA, packageB}, {packageB, packageC}},
				),
			},
			{
				lockfile: "lock2.file",
				deps:     []shared.PackageDependency{packageD, packageE, packageF},
				graph: shared.TestDependencyGraphLiteral(
					[]shared.PackageDependency{packageD},
					false,
					// d -> e -> f
					[][]shared.PackageDependency{{packageD, packageE}, {packageE, packageF}},
				),
			},
		}

		commit := "d34dd00d"

		for _, res := range results {
			// Upsert twice to test idempotency
			if err := store.UpsertLockfileGraph(ctx, "foo", commit, res.lockfile, res.deps, res.graph); err != nil {
				t.Fatalf("error: %s", err)
			}
			if err := store.UpsertLockfileGraph(ctx, "foo", commit, res.lockfile, res.deps, res.graph); err != nil {
				t.Fatalf("error: %s", err)
			}
		}

		// Check that all packages have been inserted
		names, err := queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		wantNames := []string{}
		for _, res := range results {
			for _, pkg := range res.deps {
				wantNames = append(wantNames, string(pkg.PackageSyntax()))
			}
		}
		if diff := cmp.Diff(wantNames, names); diff != "" {
			t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
		}

		// Check per lockfile now
		for _, res := range results {
			// Now check whether the direct dependency was inserted
			names, err := queryDirectDeps(t, ctx, store, "foo", commit, res.lockfile)
			if err != nil {
				t.Fatalf("database query error: %s", err)
			}

			wantNames := []string{}
			roots, _ := res.graph.Roots()
			for _, r := range roots {
				wantNames = append(wantNames, string(r.PackageSyntax()))
			}
			if diff := cmp.Diff(wantNames, names); diff != "" {
				t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
			}

			// Check that packages have been inserted with right lockfile
			names, err = basestore.ScanStrings(store.db.Query(ctx, sqlf.Sprintf(`SELECT package_name FROM codeintel_lockfile_references WHERE resolution_lockfile = %s ORDER BY package_name`, res.lockfile)))
			if err != nil {
				t.Fatalf("database query error: %s", err)
			}
			wantNames = []string{}
			for _, pkg := range res.deps {
				wantNames = append(wantNames, string(pkg.PackageSyntax()))
			}
			if diff := cmp.Diff(wantNames, names); diff != "" {
				t.Errorf("unexpected lockfile packages (-want +got):\n%s", diff)
			}
		}
	})
}

func queryDirectDeps(t *testing.T, ctx context.Context, store *store, repoName, commit, lockfile string) ([]string, error) {
	t.Helper()

	q := sqlf.Sprintf(`
	SELECT package_name
	FROM codeintel_lockfile_references
	WHERE (
		SELECT codeintel_lockfile_reference_ids
		FROM codeintel_lockfiles lf
		JOIN repo r ON r.id = lf.repository_id
		WHERE
		r.name = %s AND
		lf.commit_bytea = %s AND
		lf.lockfile = %s
	) @> ARRAY[id]
	ORDER BY package_name;
    `,
		repoName,
		dbutil.CommitBytea(commit),
		lockfile,
	)

	return basestore.ScanStrings(store.db.Query(ctx, q))
}

func queryLockfileReferences(t *testing.T, ctx context.Context, store *store, repoName, commit string) ([]string, error) {
	t.Helper()

	q := sqlf.Sprintf(`
	SELECT package_name
	FROM codeintel_lockfile_references
	WHERE
	resolution_repository_id = (SELECT id FROM repo WHERE name = %s) AND
	resolution_commit_bytea = %s
	ORDER BY package_name
`,
		"foo",
		dbutil.CommitBytea(commit),
	)

	return basestore.ScanStrings(store.db.Query(ctx, q))
}

func TestLockfileDependencies_SingleLockfile(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral("A", "1", "2", "pkg-A", "4")
	packageB := shared.TestPackageDependencyLiteral("B", "2", "3", "pkg-B", "5")
	packageC := shared.TestPackageDependencyLiteral("C", "3", "4", "pkg-C", "6")
	packageD := shared.TestPackageDependencyLiteral("D", "4", "5", "pkg-D", "7")
	packageE := shared.TestPackageDependencyLiteral("E", "5", "6", "pkg-E", "8")
	packageF := shared.TestPackageDependencyLiteral("F", "6", "7", "pkg-F", "9")

	lockfile := "lock.file"

	depsAtCommit := map[string]struct {
		list  []shared.PackageDependency
		graph shared.DependencyGraph
	}{
		"cafebabe": {
			list: []shared.PackageDependency{packageA, packageB, packageC},
			graph:
			// a -> b -> c
			shared.DependencyGraphLiteral{
				RootPkgs: []shared.PackageDependency{packageA},
				Edges:    [][]shared.PackageDependency{{packageA, packageB}, {packageB, packageC}},
			},
		},
		"deadbeef": {
			list: []shared.PackageDependency{packageA, packageB, packageD, packageE},
			//  / b
			// a
			//  \ d -> e
			graph: shared.DependencyGraphLiteral{
				RootPkgs: []shared.PackageDependency{packageA},
				Edges: [][]shared.PackageDependency{
					{packageA, packageB},
					{packageA, packageD},
					{packageD, packageE},
				},
			},
		},
		"deadc0de": {
			list: []shared.PackageDependency{packageB, packageF},
			graph: shared.DependencyGraphLiteral{
				// both roots:
				// b
				// f
				RootPkgs: []shared.PackageDependency{packageB, packageF},
				Edges:    [][]shared.PackageDependency{},
			},
		},
		// no list, no graph
		"deadd00d": {list: nil, graph: nil},
		// list, but no graph
		"deadd002": {list: []shared.PackageDependency{packageA, packageB}, graph: nil},
	}

	for commit, deps := range depsAtCommit {
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, lockfile, deps.list, deps.graph); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	// Update twice to show idempotency
	for commit, deps := range depsAtCommit {
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, lockfile, deps.list, deps.graph); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	t.Run("IncludeTransitive:false", func(t *testing.T) {
		// Query direct dependencies
		for commit, expectedDeps := range depsAtCommit {
			directDeps, found, err := store.LockfileDependencies(ctx, LockfileDependenciesOpts{
				RepoName:          "foo",
				Commit:            commit,
				IncludeTransitive: false,
			})
			if err != nil {
				t.Fatalf("unexpected error querying lockfile dependencies of %s: %s", commit, err)
			}
			if !found {
				t.Fatalf("expected dependencies to be cached for %s", commit)
			}

			var wantDirectDeps []shared.PackageDependency
			if expectedDeps.graph == nil {
				// If we don't have a graph we expect all deps to be direct deps
				wantDirectDeps = expectedDeps.list
			} else {
				graph := expectedDeps.graph.(shared.DependencyGraphLiteral)
				wantDirectDeps = graph.RootPkgs
			}

			if a, b := len(wantDirectDeps), len(directDeps); a != b {
				t.Fatalf("unexpected len of dependencies for commit %s: want=%d, have=%d", commit, a, b)
			}

			if diff := cmp.Diff(wantDirectDeps, directDeps); diff != "" {
				t.Fatalf("unexpected dependencies for commit %s (-have, +want): %s", commit, diff)
			}
		}
	})

	t.Run("IncludeTransitive:true", func(t *testing.T) {
		// Query direct + transitive dependencies
		for commit, expectedDeps := range depsAtCommit {
			allDeps, found, err := store.LockfileDependencies(ctx, LockfileDependenciesOpts{
				RepoName:          "foo",
				Commit:            commit,
				IncludeTransitive: true,
			})
			if err != nil {
				t.Fatalf("unexpected error querying lockfile dependencies of %s: %s", commit, err)
			}
			if !found {
				t.Fatalf("expected dependencies to be cached for %s", commit)
			}

			// With IncludeTransitive
			if a, b := len(expectedDeps.list), len(allDeps); a != b {
				t.Fatalf("unexpected len of dependencies for commit %s: want=%d, have=%d", commit, a, b)
			}
			if diff := cmp.Diff(expectedDeps.list, allDeps); diff != "" {
				t.Fatalf("unexpected dependencies for commit %s (-have, +want): %s", commit, diff)
			}
		}
	})

	missingCommit := "d00dd00d"
	_, found, err := store.LockfileDependencies(ctx, LockfileDependenciesOpts{
		RepoName: "foo",
		Commit:   missingCommit,
	})
	if err != nil {
		t.Fatalf("unexpected error querying lockfile dependencies: %s", err)
	} else if found {
		t.Fatalf("expected no dependencies to be cached for %s", missingCommit)
	}
}

func TestLockfileDependencies_MultipleLockfiles(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral("A", "1", "2", "pkg-A", "4")
	packageB := shared.TestPackageDependencyLiteral("B", "2", "3", "pkg-B", "5")
	packageC := shared.TestPackageDependencyLiteral("C", "3", "4", "pkg-C", "6")
	packageD := shared.TestPackageDependencyLiteral("D", "4", "5", "pkg-D", "7")
	packageE := shared.TestPackageDependencyLiteral("E", "5", "6", "pkg-E", "8")
	packageF := shared.TestPackageDependencyLiteral("F", "6", "7", "pkg-F", "9")

	depsInLockfile := map[string]struct {
		deps  []shared.PackageDependency
		graph shared.DependencyGraph
	}{
		"lockfile.1": {
			deps: []shared.PackageDependency{packageA, packageB, packageC},
			graph:
			// a -> b -> c
			shared.DependencyGraphLiteral{
				RootPkgs: []shared.PackageDependency{packageA},
				Edges:    [][]shared.PackageDependency{{packageA, packageB}, {packageB, packageC}},
			},
		},
		"lockfile.2": {
			deps: []shared.PackageDependency{packageD, packageE, packageF},
			// d -> e -> f
			graph: shared.DependencyGraphLiteral{
				RootPkgs: []shared.PackageDependency{packageD},
				Edges: [][]shared.PackageDependency{
					{packageD, packageE},
					{packageE, packageF},
				},
			},
		},
	}

	commit := "d34db33f"

	for lockfile, result := range depsInLockfile {
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, lockfile, result.deps, result.graph); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	// Update twice to show idempotency
	for lockfile, result := range depsInLockfile {
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, lockfile, result.deps, result.graph); err != nil {
			t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
		}
	}

	// Query per lockfile
	for lockfile, expectedDeps := range depsInLockfile {
		directDeps, found, err := store.LockfileDependencies(ctx, LockfileDependenciesOpts{
			RepoName:          "foo",
			Commit:            commit,
			IncludeTransitive: false,
			Lockfile:          lockfile,
		})
		if err != nil {
			t.Fatalf("unexpected error querying lockfile dependencies of %s: %s", commit, err)
		}
		if !found {
			t.Fatalf("expected dependencies to be cached for %s", commit)
		}
		sort.Slice(directDeps, func(i, j int) bool { return directDeps[i].RepoName() < directDeps[j].RepoName() })

		graph := expectedDeps.graph.(shared.DependencyGraphLiteral)
		wantDirectDeps := graph.RootPkgs

		if a, b := len(wantDirectDeps), len(directDeps); a != b {
			t.Fatalf("unexpected len of dependencies for commit %s: want=%d, have=%d", commit, a, b)
		}

		if diff := cmp.Diff(wantDirectDeps, directDeps); diff != "" {
			t.Fatalf("unexpected dependencies for commit %s (-have, +want): %s", commit, diff)
		}
	}

	// Query without specifying lockfile
	deps, found, err := store.LockfileDependencies(ctx, LockfileDependenciesOpts{
		RepoName:          "foo",
		Commit:            commit,
		IncludeTransitive: false,
	})
	if err != nil {
		t.Fatalf("unexpected error querying lockfile dependencies of %s: %s", commit, err)
	}
	if !found {
		t.Fatalf("expected dependencies to be cached for %s", commit)
	}

	wantDirectDeps := []shared.PackageDependency{packageA, packageD}
	if a, b := len(wantDirectDeps), len(deps); a != b {
		t.Fatalf("unexpected len of dependencies for commit %s: want=%d, have=%d", commit, a, b)
	}

	if diff := cmp.Diff(wantDirectDeps, deps); diff != "" {
		t.Fatalf("unexpected dependencies for commit %s (-have, +want): %s", commit, diff)
	}
}

func TestUpsertDependencyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	batches := [][]shared.Repo{
		{
			// Test same-set flushes
			shared.Repo{Scheme: "npm", Name: "bar", Version: "2.0.0"}, // id=1
			shared.Repo{Scheme: "npm", Name: "bar", Version: "2.0.0"}, // id=2, duplicate
		},
		{
			shared.Repo{Scheme: "npm", Name: "bar", Version: "3.0.0"}, // id=3
			shared.Repo{Scheme: "npm", Name: "foo", Version: "1.0.0"}, // id=4
		},
		{
			// Test different-set flushes
			shared.Repo{Scheme: "npm", Name: "foo", Version: "1.0.0"}, // id=5, duplicate
			shared.Repo{Scheme: "npm", Name: "foo", Version: "2.0.0"}, // id=6
		},
	}

	var allNewDeps []shared.Repo
	for _, batch := range batches {
		newDeps, err := store.UpsertDependencyRepos(ctx, batch)
		if err != nil {
			t.Fatal(err)
		}

		allNewDeps = append(allNewDeps, newDeps...)
	}

	want := []shared.Repo{
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 3, Scheme: "npm", Name: "bar", Version: "3.0.0"},
		{ID: 4, Scheme: "npm", Name: "foo", Version: "1.0.0"},
		{ID: 6, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}
	if diff := cmp.Diff(allNewDeps, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}

	have, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}
}

func TestListDependencyRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	batches := []shared.Repo{
		{Scheme: "npm", Name: "bar", Version: "2.0.0"},    // id=1
		{Scheme: "npm", Name: "foo", Version: "1.0.0"},    // id=2
		{Scheme: "npm", Name: "bar", Version: "2.0.1"},    // id=3
		{Scheme: "npm", Name: "foo", Version: "1.0.0"},    // id=4
		{Scheme: "npm", Name: "bar", Version: "3.0.0"},    // id=5
		{Scheme: "npm", Name: "banana", Version: "2.0.0"}, // id=6
		{Scheme: "npm", Name: "turtle", Version: "4.2.0"}, // id=7
	}

	if _, err := store.UpsertDependencyRepos(ctx, batches); err != nil {
		t.Fatal(err)
	}

	var lastName reposource.PackageName
	lastName = ""
	for _, test := range [][]shared.Repo{
		{{Scheme: "npm", Name: "banana"}, {Scheme: "npm", Name: "bar"}, {Scheme: "npm", Name: "foo"}},
		{{Scheme: "npm", Name: "turtle"}},
	} {
		depRepos, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
			Scheme:          "npm",
			After:           lastName,
			Limit:           3,
			ExcludeVersions: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := range depRepos {
			depRepos[i].ID = 0
		}

		lastName = depRepos[len(depRepos)-1].Name

		if diff := cmp.Diff(depRepos, test); diff != "" {
			t.Fatalf("mismatch (-have, +want): %s", diff)
		}
	}
}

func TestDeleteDependencyReposByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	repos := []shared.Repo{
		// Test same-set flushes
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 2, Scheme: "npm", Name: "bar", Version: "3.0.0"}, // deleted
		{ID: 3, Scheme: "npm", Name: "foo", Version: "1.0.0"}, // deleted
		{ID: 4, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}

	if _, err := store.UpsertDependencyRepos(ctx, repos); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteDependencyReposByID(ctx, 2, 3); err != nil {
		t.Fatalf(err.Error())
	}

	have, err := store.ListDependencyRepos(ctx, ListDependencyReposOpts{
		Scheme: shared.NpmPackagesScheme,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []shared.Repo{
		{ID: 1, Scheme: "npm", Name: "bar", Version: "2.0.0"},
		{ID: 4, Scheme: "npm", Name: "foo", Version: "2.0.0"},
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatalf("mismatch (-have, +want): %s", diff)
	}
}

func TestSelectRepoRevisionsToResolve(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('repo-1')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral("A", "v1", "2", "3", "4")
	packageB := shared.TestPackageDependencyLiteral("B", "v2", "3", "4", "5")
	packageC := shared.TestPackageDependencyLiteral("C", "v3", "4", "5", "6")
	packageD := shared.TestPackageDependencyLiteral("D", "v4", "5", "6", "7")
	packageE := shared.TestPackageDependencyLiteral("E", "v5", "6", "7", "8")

	commit := "d34df00d"
	repoName := "repo-1"
	packages := []shared.PackageDependency{packageA, packageB, packageC, packageD, packageE}

	if err := store.UpsertLockfileGraph(ctx, repoName, commit, "lock.file", packages, nil); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}

	now := timeutil.Now()

	selected, err := store.selectRepoRevisionsToResolve(ctx, 3, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions := map[string][]string{
		"A": {"v1"},
		"B": {"v2"},
		"C": {"v3"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected repo revisions (-want +got):\n%s", diff)
	}

	selected, err = store.selectRepoRevisionsToResolve(ctx, 3, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{
		"D": {"v4"},
		"E": {"v5"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected repo revisions (-want +got):\n%s", diff)
	}

	// Run it again, but all should be resolved in timeframe
	selected, err = store.selectRepoRevisionsToResolve(ctx, 6, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected repo revisions (-want +got):\n%s", diff)
	}

	// Run it again, but in the future, all should be resolved in timeframe
	now = now.Add(24 * time.Hour)

	selected, err = store.selectRepoRevisionsToResolve(ctx, 6, 24*time.Hour, now)
	if err != nil {
		t.Fatalf("unexpected error selecting repo revisions to resolve: %s", err)
	}

	expectedRepoRevisions = map[string][]string{
		"A": {"v1"},
		"B": {"v2"},
		"C": {"v3"},
		"D": {"v4"},
		"E": {"v5"},
	}
	if diff := cmp.Diff(selected, expectedRepoRevisions); diff != "" {
		t.Errorf("unexpected sourced commits (-want +got):\n%s", diff)
	}
}

func TestUpdateResolvedRevisions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for _, repo := range []string{"repo-1", "repo-2", "repo-3", "pkg-1", "pkg-2", "pkg-3", "pkg-4"} {
		if err := store.db.Exec(ctx, sqlf.Sprintf(`INSERT INTO repo (name) VALUES (%s)`, repo)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	var (
		packageA = shared.TestPackageDependencyLiteral("pkg-1", "v1", "2", "3", "4")
		packageB = shared.TestPackageDependencyLiteral("pkg-2", "v2", "3", "4", "5")
		packageC = shared.TestPackageDependencyLiteral("pkg-3", "v3", "4", "5", "6")
		packageD = shared.TestPackageDependencyLiteral("pkg-4", "v4", "5", "6", "7")
	)

	if err := store.UpsertLockfileGraph(ctx, "repo-1", "cafebabe", "lock.file", []shared.PackageDependency{packageA, packageB, packageC, packageD}, nil); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}

	resolvedRevisions := map[string]map[string]string{
		"pkg-2": {"v2": "deadbeef"},
		"pkg-3": {"v3": "deadd00d"},
	}
	if err := store.UpdateResolvedRevisions(ctx, resolvedRevisions); err != nil {
		t.Fatalf("unexpected error updating resolved revisions: %s", err)
	}

	for repoName, resolvedRevs := range resolvedRevisions {
		for revspec, commit := range resolvedRevs {
			q := sqlf.Sprintf(`
				SELECT 1
				FROM codeintel_lockfile_references lr
				JOIN repo r ON r.id = lr.repository_id
				WHERE r.name = %s AND lr.revspec = %s AND lr.commit_bytea = %s
			`,
				repoName,
				revspec,
				dbutil.CommitBytea(commit),
			)

			_, ok, err := basestore.ScanFirstInt(store.db.Query(ctx, q))
			if err != nil {
				t.Fatalf("failed to query resolved revisions for repo %s: %s", repoName, err)
			}
			if !ok {
				t.Fatalf("revspec %q for repo %q was not updated to commit %s", revspec, repoName, commit)
			}
		}
	}
}

func TestLockfileDependents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for _, repo := range []string{"repo-1", "repo-2", "repo-3", "pkg-1", "pkg-2", "pkg-3", "pkg-4"} {
		if err := store.db.Exec(ctx, sqlf.Sprintf(`INSERT INTO repo (name) VALUES (%s)`, repo)); err != nil {
			t.Fatalf(err.Error())
		}
	}

	var (
		packageA = shared.TestPackageDependencyLiteral("pkg-1", "v1", "2", "3", "4")
		packageB = shared.TestPackageDependencyLiteral("pkg-2", "v2", "3", "4", "5")
		packageC = shared.TestPackageDependencyLiteral("pkg-3", "v3", "4", "5", "6")
		packageD = shared.TestPackageDependencyLiteral("pkg-4", "v4", "5", "6", "7")
	)

	if err := store.UpsertLockfileGraph(ctx, "repo-1", "cafebabe", "lock.file", []shared.PackageDependency{packageA, packageB, packageC, packageD}, nil); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}
	if err := store.UpsertLockfileGraph(ctx, "repo-2", "cafebeef", "lock.file", []shared.PackageDependency{packageB}, nil); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}
	if err := store.UpsertLockfileGraph(ctx, "repo-3", "d00dd00d", "lock.file", []shared.PackageDependency{packageC}, nil); err != nil {
		t.Fatalf("unexpected error upserting lockfile dependencies: %s", err)
	}

	resolvedRevisions := map[string]map[string]string{
		"pkg-2": {"v2": "deadbeef"},
		"pkg-3": {"v3": "deadd00d"},
	}
	if err := store.UpdateResolvedRevisions(ctx, resolvedRevisions); err != nil {
		t.Fatalf("unexpected error updating resolved revisions: %s", err)
	}

	// Should be empty; nothing resolved here
	deps, err := store.LockfileDependents(ctx, "pkg-1", "cafecafe")
	if err != nil {
		t.Fatalf("unexpected error listing lockfile dependents: %s", err)
	}
	if diff := cmp.Diff([]api.RepoCommit(nil), deps); diff != "" {
		t.Errorf("unexpected lockfile dependents (-want +got):\n%s", diff)
	}

	// Should include repo-1 and repo-2
	deps, err = store.LockfileDependents(ctx, "pkg-2", "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error listing lockfile dependents: %s", err)
	}
	expectedDeps := []api.RepoCommit{
		{Repo: api.RepoName("repo-1"), CommitID: api.CommitID("cafebabe")},
		{Repo: api.RepoName("repo-2"), CommitID: api.CommitID("cafebeef")},
	}
	if diff := cmp.Diff(expectedDeps, deps); diff != "" {
		t.Errorf("unexpected lockfile dependents (-want +got):\n%s", diff)
	}

	// Should include repo-1 and repo-3
	deps, err = store.LockfileDependents(ctx, "pkg-3", "deadd00d")
	if err != nil {
		t.Fatalf("unexpected error listing lockfile dependents: %s", err)
	}
	expectedDeps = []api.RepoCommit{
		{Repo: "repo-1", CommitID: "cafebabe"},
		{Repo: "repo-3", CommitID: "d00dd00d"},
	}
	if diff := cmp.Diff(expectedDeps, deps); diff != "" {
		t.Errorf("unexpected lockfile dependents (-want +got):\n%s", diff)
	}
}

func TestListAndGetLockfileIndexes(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	now := timeutil.Now()

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('bar')`); err != nil {
		t.Fatalf(err.Error())
	}

	packageA := shared.TestPackageDependencyLiteral(api.RepoName("A"), "1", "2", "pkg-A", "4")
	packageB := shared.TestPackageDependencyLiteral(api.RepoName("B"), "2", "3", "pkg-B", "5")

	// Insert data
	for _, tt := range []struct {
		repoName string
		deps     []shared.PackageDependency
		graph    shared.DependencyGraph
		commit   string
		lockfile string
	}{
		{
			repoName: "foo",
			deps:     []shared.PackageDependency{packageA, packageB},
			graph:    shared.TestDependencyGraphLiteral([]shared.PackageDependency{packageA}, false, [][]shared.PackageDependency{{packageA, packageB}}),
			commit:   "cafebabe",
			lockfile: "lock.file",
		},
		{
			repoName: "foo",
			deps:     []shared.PackageDependency{packageA, packageB},
			graph:    shared.TestDependencyGraphLiteral([]shared.PackageDependency{packageA}, false, [][]shared.PackageDependency{{packageA, packageB}}),
			commit:   "d34db33f",
			lockfile: "lock.file",
		},
		{
			repoName: "bar",
			deps:     []shared.PackageDependency{packageA},
			graph:    nil,
			commit:   "d34db33f",
			lockfile: "lock2.file",
		},
	} {
		if err := store.upsertLockfileGraphAt(ctx, tt.repoName, tt.commit, tt.lockfile, tt.deps, tt.graph, now); err != nil {
			t.Fatalf("error: %s", err)
		}
	}

	// Query
	lockfileIndexes := []shared.LockfileIndex{
		{ID: 1, RepositoryID: 1, Commit: "cafebabe", LockfileReferenceIDs: []int{1}, Lockfile: "lock.file", Fidelity: "graph", CreatedAt: now, UpdatedAt: now},
		{ID: 2, RepositoryID: 1, Commit: "d34db33f", LockfileReferenceIDs: []int{3}, Lockfile: "lock.file", Fidelity: "graph", CreatedAt: now, UpdatedAt: now},
		{ID: 3, RepositoryID: 2, Commit: "d34db33f", LockfileReferenceIDs: []int{5}, Lockfile: "lock2.file", Fidelity: "flat", CreatedAt: now, UpdatedAt: now},
	}

	for i, tt := range []struct {
		opts          ListLockfileIndexesOpts
		expected      []shared.LockfileIndex
		expectedCount int
	}{
		{
			opts:          ListLockfileIndexesOpts{},
			expected:      []shared.LockfileIndex{lockfileIndexes[0], lockfileIndexes[1], lockfileIndexes[2]},
			expectedCount: 3,
		},
		{
			opts:          ListLockfileIndexesOpts{Limit: 2},
			expected:      []shared.LockfileIndex{lockfileIndexes[0], lockfileIndexes[1]},
			expectedCount: 3,
		},
		{
			opts:          ListLockfileIndexesOpts{After: 1, Limit: 2},
			expected:      []shared.LockfileIndex{lockfileIndexes[1], lockfileIndexes[2]},
			expectedCount: 3,
		},
		{
			opts:          ListLockfileIndexesOpts{After: 2, Limit: 2},
			expected:      []shared.LockfileIndex{lockfileIndexes[2]},
			expectedCount: 3,
		},
		{
			opts:          ListLockfileIndexesOpts{RepoName: "foo"},
			expected:      []shared.LockfileIndex{lockfileIndexes[0], lockfileIndexes[1]},
			expectedCount: 2,
		},
		{
			opts:          ListLockfileIndexesOpts{RepoName: "bar"},
			expected:      []shared.LockfileIndex{lockfileIndexes[2]},
			expectedCount: 1,
		},
		{
			opts:          ListLockfileIndexesOpts{Commit: "cafebabe"},
			expected:      []shared.LockfileIndex{lockfileIndexes[0]},
			expectedCount: 1,
		},
		{
			opts:          ListLockfileIndexesOpts{Commit: "d34db33f", RepoName: "bar"},
			expected:      []shared.LockfileIndex{lockfileIndexes[2]},
			expectedCount: 1,
		},
		{
			opts:          ListLockfileIndexesOpts{Lockfile: "lock.file"},
			expected:      []shared.LockfileIndex{lockfileIndexes[0], lockfileIndexes[1]},
			expectedCount: 2,
		},
		{
			opts:          ListLockfileIndexesOpts{Lockfile: "lock2.file"},
			expected:      []shared.LockfileIndex{lockfileIndexes[2]},
			expectedCount: 1,
		},
	} {
		lockfiles, count, err := store.ListLockfileIndexes(ctx, tt.opts)
		if err != nil {
			t.Fatalf("error: %s", err)
		}

		if diff := cmp.Diff(tt.expected, lockfiles); diff != "" {
			t.Errorf("[%d] unexpected lockfiles (-want +got):\n%s", i, diff)
		}

		if diff := cmp.Diff(tt.expectedCount, count); diff != "" {
			t.Errorf("[%d] unexpected lockfiles count (-want +got):\n%s", i, diff)
		}
	}

	for i, tt := range []struct {
		opts     GetLockfileIndexOpts
		expected shared.LockfileIndex
	}{
		{
			opts:     GetLockfileIndexOpts{ID: lockfileIndexes[0].ID},
			expected: lockfileIndexes[0],
		},
		{
			opts:     GetLockfileIndexOpts{ID: lockfileIndexes[1].ID},
			expected: lockfileIndexes[1],
		},
		{
			// two indexes for this repo, but first one is returned
			opts:     GetLockfileIndexOpts{RepoName: "foo"},
			expected: lockfileIndexes[0],
		},
		{
			opts:     GetLockfileIndexOpts{RepoName: "foo", Commit: "d34db33f"},
			expected: lockfileIndexes[1],
		},
		{
			opts:     GetLockfileIndexOpts{Lockfile: "lock2.file"},
			expected: lockfileIndexes[2],
		},
	} {
		lockfile, err := store.GetLockfileIndex(ctx, tt.opts)
		if err != nil {
			t.Fatalf("error: %s", err)
		}

		if diff := cmp.Diff(tt.expected, lockfile); diff != "" {
			t.Errorf("[%d] unexpected lockfiles (-want +got):\n%s", i, diff)
		}
	}

	_, err := store.GetLockfileIndex(ctx, GetLockfileIndexOpts{ID: 1999})
	if err != ErrLockfileIndexNotFound {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestDeleteLockfileIndexByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo')`); err != nil {
		t.Fatalf(err.Error())
	}

	t.Run("with dependencies", func(t *testing.T) {
		packageA := shared.TestPackageDependencyLiteral("A", "1", "2", "pkg-A", "4")
		packageB := shared.TestPackageDependencyLiteral("B", "2", "3", "pkg-B", "5")
		packageC := shared.TestPackageDependencyLiteral("C", "3", "4", "pkg-C", "6")

		deps := []shared.PackageDependency{packageA, packageB, packageC}
		graph := shared.TestDependencyGraphLiteral(
			[]shared.PackageDependency{packageA},
			false,
			[][]shared.PackageDependency{{packageA, packageB}, {packageA, packageC}},
		)
		commit := "cafebabe"

		// Insert and check everything's been inserted
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, graph); err != nil {
			t.Fatalf("error: %s", err)
		}
		index, err := store.GetLockfileIndex(ctx, GetLockfileIndexOpts{RepoName: "foo", Commit: commit, Lockfile: "lock.file"})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		names, err := queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if len(names) != 3 {
			t.Fatalf("references not inserted")
		}

		// Delete
		err = store.DeleteLockfileIndexByID(ctx, index.ID)
		if err != nil {
			t.Fatalf("failed to delete: %s", err)
		}

		// Query again to make sure it's deleted
		_, err = store.GetLockfileIndex(ctx, GetLockfileIndexOpts{ID: index.ID})
		if err != ErrLockfileIndexNotFound {
			t.Fatalf("unexpected error: %s", err)
		}

		// Query references again to make sure they're deleted
		names, err = queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if len(names) != 0 {
			t.Fatalf("references not inserted")
		}

		// Delete again
		err = store.DeleteLockfileIndexByID(ctx, index.ID)
		if !errors.Is(err, ErrLockfileIndexNotFound) {
			t.Fatalf("wrong error: %s (%T)", err, err)
		}
	})

	t.Run("without any dependencies", func(t *testing.T) {
		deps := []shared.PackageDependency{}
		commit := "d34db30f"

		// Insert and check everything's been inserted
		if err := store.UpsertLockfileGraph(ctx, "foo", commit, "lock.file", deps, nil); err != nil {
			t.Fatalf("error: %s", err)
		}
		index, err := store.GetLockfileIndex(ctx, GetLockfileIndexOpts{RepoName: "foo", Commit: commit, Lockfile: "lock.file"})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		names, err := queryLockfileReferences(t, ctx, store, "foo", commit)
		if err != nil {
			t.Fatalf("database query error: %s", err)
		}
		if len(names) != 0 {
			t.Fatalf("references not inserted")
		}

		// Delete
		err = store.DeleteLockfileIndexByID(ctx, index.ID)
		if err != nil {
			t.Fatalf("failed to delete: %s", err)
		}

		// Query again to make sure it's deleted
		_, err = store.GetLockfileIndex(ctx, GetLockfileIndexOpts{ID: index.ID})
		if err != ErrLockfileIndexNotFound {
			t.Fatalf("unexpected error: %s", err)
		}

		// Delete again
		err = store.DeleteLockfileIndexByID(ctx, index.ID)
		if !errors.Is(err, ErrLockfileIndexNotFound) {
			t.Fatalf("wrong error: %s", err)
		}
	})
}
