package localstore

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory/filelang"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// updateDefs is a helper function to update definitions in the global_defs and defs tables.
func updateDefs(ctx context.Context, local bool, repo string, commitID string, unitType string, unitName string) error {
	table := "global_defs"
	if local {
		table = "defs"
	}

	trackedRepo := repotrackutil.GetTrackedRepo(repo)
	observe := func(part string, start time.Time) {
		since := time.Since(start)
		defsUpdateDuration.WithLabelValues(table, trackedRepo, part).Observe(since.Seconds())
	}
	defer observe("total", time.Now())

	start := time.Now()
	defs_, err := store.GraphFromContext(ctx).Defs(
		sstore.ByRepoCommitIDs(sstore.Version{Repo: repo, CommitID: commitID}),
		sstore.ByUnits(unit.ID2{Type: unitType, Name: unitName}),
	)
	observe("graphstore", start)
	if err != nil {
		return err
	}

	start = time.Now()
	langWarnCount := 0
	type upsert struct {
		query string
		args  []interface{}
	}
	var upsertSQLs []upsert
	for _, d := range defs_ {
		if !shouldIndex(d) {
			continue
		}

		if d.Repo == "" {
			d.Repo = repo
		}

		var docstring string
		if len(d.Docs) == 1 {
			docstring = d.Docs[0].Data
		} else {
			for _, candidate := range d.Docs {
				if candidate.Format == "" || strings.ToLower(candidate.Format) == "text/plain" {
					docstring = candidate.Data
				}
			}
		}

		data, err := d.Data.Marshal()
		if err != nil {
			data = []byte{}
		}
		bow := strings.Join(search.BagOfWordsToTokens(search.BagOfWords(d)), " ")

		languageID, err := toDBLang(strings.ToLower(graph.PrintFormatter(d).Language()))
		if err != nil {
			langWarnCount++
		}

		var args []interface{}
		arg := func(v interface{}) string {
			args = append(args, v)
			return gorp.PostgresDialect{}.BindVar(len(args) - 1)
		}

		upsertSQL := `
WITH upsert AS (
UPDATE ` + table + ` SET name=` + arg(d.Name) +
			`, kind=` + arg(d.Kind) +
			`, file=` + arg(d.File) +
			`, language=` + arg(languageID) +
			`, commit_id=` + arg(d.CommitID) +
			`, updated_at=now()` +
			`, data=` + arg(data) +
			`, bow=` + arg(bow) +
			`, doc=` + arg(docstring) +
			` WHERE repo=` + arg(d.Repo) +
			` AND unit_type=` + arg(d.UnitType) +
			` AND unit=` + arg(d.Unit) +
			` AND path=` + arg(d.Path) +
			` RETURNING *
)
INSERT INTO ` + table + ` (repo, commit_id, unit_type, unit, path, name, kind, file, language, updated_at, data, bow, doc) SELECT ` +
			arg(d.Repo) + `, ` +
			arg(d.CommitID) + `, ` +
			arg(d.UnitType) + `, ` +
			arg(d.Unit) + `, ` +
			arg(d.Path) + `, ` +
			arg(d.Name) + `, ` +
			arg(d.Kind) + `, ` +
			arg(d.File) + `, ` +
			arg(languageID) + `, ` +
			`now(), ` +
			arg(data) + `, ` +
			arg(bow) + `, ` +
			arg(docstring) + `
WHERE NOT EXISTS (SELECT * FROM upsert);`
		upsertSQLs = append(upsertSQLs, upsert{query: upsertSQL, args: args})
	}
	if langWarnCount > 0 {
		log15.Warn("could not determine language for all defs", "noLang", langWarnCount, "allDefs", len(defs_))
	}
	observe("genupserts", start)

	defer observe("transaction", time.Now())
	if err := dbutil.Transact(graphDBH(ctx), func(tx gorp.SqlExecutor) error {
		start = time.Now()
		for _, upsertSQL := range upsertSQLs {
			if _, err := tx.Exec(upsertSQL.query, upsertSQL.args...); err != nil {
				return err
			}
		}
		observe("upsert", start)

		// Delete old entries
		start = time.Now()
		if _, err := tx.Exec(`DELETE FROM `+table+` WHERE repo=$1 AND unit_type=$2 AND unit=$3 AND commit_id!=$4`,
			repo, unitType, unitName, commitID); err != nil {
			return err
		}
		observe("delete", start)
		return nil

	}); err != nil { // end transaction
		return err
	}
	return nil
}

func shouldIndex(d *graph.Def) bool {
	// Ignore broken defs
	if d.Path == "" {
		return false
	}
	// Ignore local defs (KLUDGE)
	if d.Local || strings.Contains(d.Path, "$") {
		return false
	}
	// Ignore vendored defs
	if filelang.IsVendored(d.File, false) {
		return false
	}
	// Ignore defs in Go test files
	if strings.HasSuffix(d.File, "_test.go") {
		return false
	}
	return true
}

var defsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "update_duration_seconds",
	Help:      "Duration for updating a def",
	MaxAge:    time.Hour,
}, []string{"table", "repo", "part"})

func init() {
	prometheus.MustRegister(defsUpdateDuration)
}
