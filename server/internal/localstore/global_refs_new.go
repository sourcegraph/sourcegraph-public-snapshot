package localstore

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rogpeppe/rog-go/parallel"

	"gopkg.in/gorp.v1"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
)

func init() {
	// dbDefKey DB-maps a DefKey (excluding commit-id) object. We keep
	// this in a seperate table to reduce duplication in the global_refs
	// table (postgresql does not do string interning)
	type dbDefKey struct {
		ID       int64  `db:"id"`
		Repo     string `db:"repo"`
		UnitType string `db:"unit_type"`
		Unit     string `db:"unit"`
		Path     string `db:"path"`
	}
	GraphSchema.Map.AddTableWithName(dbDefKey{}, "def_keys").SetKeys(true, "id").SetUniqueTogether("repo", "unit_type", "unit", "path")

	// dbGlobalRef DB-maps a GlobalRef object.
	type dbGlobalRef struct {
		DefKeyID  int64 `db:"def_key_id"`
		Repo      string
		File      string
		Count     int
		UpdatedAt *time.Time `db:"updated_at"`
	}
	GraphSchema.Map.AddTableWithName(dbGlobalRef{}, "global_refs_new")
	GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
		`CREATE INDEX global_refs_new_def_key_id ON global_refs_new USING btree (def_key_id);`,
		`CREATE INDEX global_refs_new_repo ON global_refs_new USING btree (repo);`,
		`CREATE MATERIALIZED VIEW global_refs_stats AS SELECT def_key_id, count(distinct repo) as repos FROM global_refs_new GROUP BY def_key_id;`,
		`CREATE UNIQUE INDEX ON global_refs_stats (def_key_id);`,
	)
	GraphSchema.DropSQL = append(GraphSchema.DropSQL,
		`DROP MATERIALIZED VIEW IF EXISTS global_refs_stats;`,
	)
}

// globalRefsNew is a DB-backed implementation of the GlobalRefs store.
type globalRefsNew struct{}

func (g *globalRefsNew) Get(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
	if op.Opt == nil {
		op.Opt = &sourcegraph.DefListRefLocationsOptions{}
	}

	// Optimization: fetch ref stats in parallel to fetching ref locations.
	var totalRepos int64
	statsDone := make(chan error)
	go func() {
		var err error
		totalRepos, err = g.getRefStats(ctx, op)
		statsDone <- err
	}()

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	innerSelectSQL := `SELECT repo, file, count FROM global_refs_new`
	innerSelectSQL += ` WHERE def_key_id=(SELECT id FROM def_keys WHERE repo=` + arg(op.Def.Repo) + ` AND unit_type=` + arg(op.Def.UnitType) + ` AND unit=` + arg(op.Def.Unit) + ` AND path=` + arg(op.Def.Path) + `)`
	innerSelectSQL += fmt.Sprintf(" LIMIT %s OFFSET %s", arg(op.Opt.PerPageOrDefault()), arg(op.Opt.Offset()))
	if len(op.Opt.Repos) > 0 {
		repoBindVars := make([]string, len(op.Opt.Repos))
		for i, r := range op.Opt.Repos {
			repoBindVars[i] = arg(r)
		}
		innerSelectSQL += " AND repo in (" + strings.Join(repoBindVars, ",") + ")"
	}

	sql := "SELECT repo, SUM(count) OVER(PARTITION BY repo) AS repo_count, file, count FROM (" + innerSelectSQL + ") res"
	orderBySQL := " ORDER BY repo_count DESC, count DESC"

	sql += orderBySQL

	// dbRefLocationsResult holds the result of the SELECT query for fetching global refs.
	type dbRefLocationsResult struct {
		Repo      string
		RepoCount int `db:"repo_count"`
		File      string
		Count     int
	}
	var dbRefResult []*dbRefLocationsResult
	if _, err := graphDBH(ctx).Select(&dbRefResult, sql, args...); err != nil {
		return nil, err
	}

	// repoRefs holds the ordered list of repos referencing this def. The list is sorted by
	// decreasing ref counts per repo, and the file list in each individual DefRepoRef is
	// also sorted by descending ref counts.
	var repoRefs []*sourcegraph.DefRepoRef
	defRepoIdx := -1
	// refsByRepo groups each referencing file by repo.
	refsByRepo := make(map[string]*sourcegraph.DefRepoRef)
	for _, r := range dbRefResult {
		if _, ok := refsByRepo[r.Repo]; !ok {
			refsByRepo[r.Repo] = &sourcegraph.DefRepoRef{
				Repo:  r.Repo,
				Count: int32(r.RepoCount),
			}
			repoRefs = append(repoRefs, refsByRepo[r.Repo])
			// Note the position of the def's own repo in the slice.
			if op.Def.Repo == r.Repo {
				defRepoIdx = len(repoRefs) - 1
			}
		}
		if r.File != "" && r.Count != 0 {
			refsByRepo[r.Repo].Files = append(refsByRepo[r.Repo].Files, &sourcegraph.DefFileRef{
				Path:  r.File,
				Count: int32(r.Count),
			})
		}
	}

	// Place the def's own repo at the head of the slice, if it exists in the
	// slice and is not at the head already.
	if defRepoIdx > 0 {
		repoRefs[0], repoRefs[defRepoIdx] = repoRefs[defRepoIdx], repoRefs[0]
	}

	// HACK: set hard limit on # of repos returned for one def, to avoid making excessive number
	// of GitHub Repos.Get calls in the accesscontrol check below.
	// TODO: remove this limit once we properly cache GitHub API responses.
	if len(repoRefs) > 100 {
		repoRefs = repoRefs[:100]
	}

	// Filter out repos that the user does not have access to.
	hasAccess := make([]bool, len(repoRefs))
	par := parallel.NewRun(30)
	var mu sync.Mutex
	for i_, r_ := range repoRefs {
		i, r := i_, r_
		par.Do(func() error {
			if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalRefs.Get", r.Repo); err == nil {
				mu.Lock()
				hasAccess[i] = true
				mu.Unlock()
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	var filteredRepoRefs []*sourcegraph.DefRepoRef
	for i, r := range repoRefs {
		if !hasAccess[i] {
			continue
		}
		filteredRepoRefs = append(filteredRepoRefs, r)
	}

	select {
	case err := <-statsDone:
		if err != nil {
			return nil, err
		}
	}

	return &sourcegraph.RefLocationsList{
		RepoRefs:   filteredRepoRefs,
		TotalRepos: int32(totalRepos),
	}, nil
}

// getRefStats fetches global ref aggregation stats pagination and display
// purposes.
func (g *globalRefsNew) getRefStats(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (int64, error) {
	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}
	where := `def_key_id=(SELECT id FROM def_keys WHERE repo=` + arg(op.Def.Repo) + ` AND unit_type=` + arg(op.Def.UnitType) + ` AND unit=` + arg(op.Def.Unit) + ` AND path=` + arg(op.Def.Path) + `)`

	// Our strategy is to defer to the potentially stale materialized view
	// if there are a large number of distinct repos. Otherwise we can
	// calculate the exact value since it should be fast to do
	count, err := graphDBH(ctx).SelectInt("SELECT repos FROM global_ref_stats WHERE "+where, args...)
	if err == nil && count > 1000 {
		return count, nil
	}

	return graphDBH(ctx).SelectInt("SELECT COUNT(DISTINCT repo) AS Repos FROM global_refs_new WHERE "+where, args...)
}

func (g *globalRefsNew) Update(ctx context.Context, op *pb.ImportOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GlobalRefs.Update", op.Repo); err != nil {
		return err
	}

	if op.Data == nil {
		return nil
	}

	// Perf optimization: Local cache of def_key_id's to avoid psql roundtrip
	defKeyIDCache := map[sourcegraph.DefSpec]int64{}

	tmpCreateSQL := `CREATE TEMPORARY TABLE global_refs_tmp (
	def_key_id bigint,
	repo TEXT,
	file TEXT,
	count integer
)
ON COMMIT DROP;`
	defKeyInsertSQL := `INSERT INTO def_keys(repo, unit_type, unit, path) VALUES($1, $2, $3, $4);`
	defKeyGetSQL := `SELECT id FROM def_keys WHERE repo=$1 AND unit_type=$2 AND unit=$3 AND path=$4`
	tmpInsertSQL := `INSERT INTO global_refs_tmp(def_key_id, repo, file, count) VALUES($1, $2, $3, 1);`
	finalDeleteSQL := `DELETE FROM global_refs_new WHERE repo=$1 AND file IN (SELECT file FROM global_refs_tmp GROUP BY file);`
	finalInsertSQL := `INSERT INTO global_refs_new(def_key_id, repo, file, count, updated_at)
	SELECT def_key_id, repo, file, sum(count) as count, now() as updated_at
	FROM global_refs_tmp
	GROUP BY def_key_id, repo, file;`

	return dbutil.Transact(graphDBH(ctx), func(tx gorp.SqlExecutor) error {
		// Create a temporary table to load all new ref data.
		if _, err := tx.Exec(tmpCreateSQL); err != nil {
			return err
		}

		// Insert refs into temporary table
		for _, r := range op.Data.Refs {
			// Ignore broken refs.
			if r.DefPath == "" {
				continue
			}
			// Ignore def refs.
			if r.Def {
				continue
			}
			if r.DefRepo == "" {
				r.DefRepo = op.Repo
			}
			if r.DefUnit == "" {
				r.DefUnit = op.Unit.Unit
			}
			if r.DefUnitType == "" {
				r.DefUnitType = op.Unit.UnitType
			}
			// Ignore ref to builtin defs of golang/go repo (string, int, bool, etc) as this
			// doesn't add significant value; yet it adds up to a lot of space in the db,
			// and queries for refs of builtin defs take long to finish.
			if r.DefUnitType == "GoPackage" && r.DefRepo == "github.com/golang/go" && r.DefUnit == "builtin" {
				continue
			}

			defKeyIDKey := sourcegraph.DefSpec{Repo: r.DefRepo, UnitType: r.DefUnitType, Unit: r.DefUnit, Path: r.DefPath}
			defKeyID, ok := defKeyIDCache[defKeyIDKey]
			if !ok {
				// Optimistically get the def key id, otherwise fallback to insertion
				var err error
				defKeyID, err = tx.SelectInt(defKeyGetSQL, r.DefRepo, r.DefUnitType, r.DefUnit, r.DefPath)
				if defKeyID == 0 || err != nil {
					if _, err = tx.Exec(defKeyInsertSQL, r.DefRepo, r.DefUnitType, r.DefUnit, r.DefPath); err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
						return err
					}
					defKeyID, err = tx.SelectInt(defKeyGetSQL, r.DefRepo, r.DefUnitType, r.DefUnit, r.DefPath)
					if err != nil {
						return err
					} else if defKeyID == 0 {
						return fmt.Errorf("Could not create or find defKeyID for (%s, %s, %s, %s)", r.DefRepo, r.DefUnitType, r.DefUnit, r.DefPath)
					}
				}
				defKeyIDCache[defKeyIDKey] = defKeyID
			}

			if _, err := tx.Exec(tmpInsertSQL, defKeyID, op.Repo, r.File); err != nil {
				return err
			}
		}

		// Purge all existing ref data for files in this source unit.
		if _, err := tx.Exec(finalDeleteSQL, op.Repo); err != nil {
			return err
		}

		// Insert refs into global refs table
		if _, err := tx.Exec(finalInsertSQL); err != nil {
			return err
		}

		return nil
	})
}

func (g *globalRefsNew) StatRefresh(ctx context.Context) error {
	_, err := graphDBH(ctx).Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY global_refs_stats;")
	return err
}
