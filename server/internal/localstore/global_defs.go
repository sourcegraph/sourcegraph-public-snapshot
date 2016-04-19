package localstore

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/gorp.v1"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/search"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	GraphSchema.Map.AddTableWithName(dbGlobalDef{}, "global_defs").SetKeys(false, "Repo", "CommitID", "UnitType", "Unit", "Path")
	GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
		`ALTER TABLE global_defs ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
		`ALTER TABLE global_defs ALTER COLUMN ref_ct SET DEFAULT 0;`,
		`CREATE INDEX bow_idx ON global_defs USING gin(to_tsvector('english', bow));`,
		`CREATE INDEX bow_idx ON global_defs USING gin(to_tsvector('english', doc));`,
	)
}

// dbGlobalDef DB-maps a GlobalDef object.
type dbGlobalDef struct {
	Repo     string `db:"repo"`
	CommitID string `db:"commit_id"`
	UnitType string `db:"unit_type"`
	Unit     string `db:"unit"`
	Path     string `db:"path"`

	Name string `db:"name"`
	Kind string `db:"kind"`
	File string `db:"file"`

	RefCount  int        `db:"ref_ct"`
	UpdatedAt *time.Time `db:"updated_at"`

	Data []byte `db:"data"`

	BoW string `db:"bow"`
	Doc string `db:"doc"`
}

func fromDBDef(d *dbGlobalDef) *sourcegraph.Def {
	if d == nil {
		return nil
	}
	var data pbtypes.RawMessage
	if err := data.Unmarshal(d.Data); err != nil {

	}
	return &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{
				Repo:     d.Repo,
				UnitType: d.UnitType,
				Unit:     d.Unit,
				Path:     d.Path,
			},

			Name: d.Name,
			Kind: d.Kind,
			File: d.File,

			Data: data,
		},
	}
}

func toDBDef(d *sourcegraph.Def) *dbGlobalDef {
	if d == nil {
		return nil
	}
	data, err := d.Data.Marshal()
	if err != nil {
		data = []byte{}
	}
	return &dbGlobalDef{
		Repo:     d.Repo,
		UnitType: d.UnitType,
		Unit:     d.Unit,
		Path:     d.Path,

		Name: d.Name,
		Kind: d.Kind,
		File: d.File,

		Data: data,
	}
}

// dbGlobalSearchResult holds the result of the SELECT query for global def search.
type dbGlobalSearchResult struct {
	dbGlobalDef

	Score float64 `db:"score"`
}

// globalDefs is a DB-backed implementation of the GlobalDefs store.
type globalDefs struct{}

func (g *globalDefs) Search(ctx context.Context, op *store.GlobalDefSearchOp) (*sourcegraph.SearchResultsList, error) {
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if op.Opt == nil {
		op.Opt = &sourcegraph.SearchOptions{}
	}

	var scoreSQL string
	if op.BoWQuery != "" {
		scoreSQL = `0.5*log(10 + ref_ct) + 100*ts_rank(to_tsvector('english', min(bow)), to_tsquery('english', ` + arg(op.BoWQuery) + `)) score`
	} else {
		scoreSQL = `ref_ct score`
	}
	selectSQL := `SELECT repo, unit_type, unit, path, name, kind, file, data, ref_ct, ` + scoreSQL + ` FROM global_defs`
	var whereSQL string
	{
		var wheres []string
		if op.RepoQuery != "" {
			wheres = append(wheres, `repo=`+arg(op.RepoQuery))
		}
		if op.UnitQuery != "" {
			wheres = append(wheres, `unit=`+arg(op.UnitQuery))
		}
		if op.UnitTypeQuery != "" {
			wheres = append(wheres, `lower(unit_type)=lower(`+arg(op.UnitTypeQuery)+`)`)
		}
		// TODO(beyang): make use of op.CaseSensitive?
		if op.BoWQuery != "" {
			wheres = append(wheres, "bow != ''")
			wheres = append(wheres, `to_tsquery('english', `+arg(op.BoWQuery)+`) @@ to_tsvector('english', bow)`)
		}
		wheres = append(wheres, `commit_id=''`) // HACK

		whereSQL = fmt.Sprint(`WHERE (` + strings.Join(wheres, ") AND (") + `)`)
	}
	groupSQL := `GROUP BY repo, unit_type, unit, path, name, kind, file, data, ref_ct`
	orderSQL := `ORDER BY score DESC`
	limitSQL := `LIMIT ` + arg(op.Opt.PerPageOrDefault())

	sql := strings.Join([]string{selectSQL, whereSQL, groupSQL, orderSQL, limitSQL}, "\n")

	var dbSearchResults []*dbGlobalSearchResult
	if _, err := graphDBH(ctx).Select(&dbSearchResults, sql, args...); err != nil {
		return nil, err
	}

	var results []*sourcegraph.SearchResult
	for _, d := range dbSearchResults {
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalDefs.Search", d.Repo); err != nil {
			continue
		}
		def := fromDBDef(&d.dbGlobalDef)
		results = append(results, &sourcegraph.SearchResult{
			Def:      *def,
			RefCount: int32(d.RefCount),
			Score:    float32(d.Score),
		})
	}
	return &sourcegraph.SearchResultsList{Results: results}, nil
}

func (g *globalDefs) Update(ctx context.Context, op *pb.ImportOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GlobalDefs.Update", op.Repo); err != nil {
		return err
	}

	if op.Data == nil {
		return nil
	}

	defDocs := make(map[graph.DefKey]string)
	for _, doc := range op.Data.Docs {
		if doc.Format == "" || doc.Format == "text" {
			defDocs[doc.DefKey] = doc.Data
		}
	}

	for _, d := range op.Data.Defs {
		// Ignore broken defs
		if d.Path == "" {
			continue
		}
		// Ignore local defs
		if d.Local || strings.Contains(d.Path, "$") {
			continue
		}
		if d.Repo == "" {
			d.Repo = op.Repo
		}
		d.CommitID = op.CommitID
		if d.Unit == "" {
			d.Unit = op.Unit.Unit
		}
		if d.UnitType == "" {
			d.UnitType = op.Unit.UnitType
		}

		data, err := d.Data.Marshal()
		if err != nil {
			data = []byte{}
		}
		bow := strings.Join(search.BagOfWordsToTokens(search.BagOfWords(d)), " ")
		docstring := defDocs[d.DefKey]

		var args []interface{}
		arg := func(v interface{}) string {
			args = append(args, v)
			return fmt.Sprintf("$%d", len(args))
		}

		upsertSQL := `
WITH upsert AS (
UPDATE global_defs SET name=` + arg(d.Name) +
			`, kind=` + arg(d.Kind) +
			`, file=` + arg(d.File) +
			`, updated_at=now(), data=` + arg(data) +
			`, bow=` + arg(bow) +
			`, doc=` + arg(docstring) +
			` WHERE repo=` + arg(d.Repo) +
			` AND commit_id=` + arg(d.CommitID) +
			` AND unit_type=` + arg(d.UnitType) +
			` AND unit=` + d.Unit +
			` AND path=` + arg(d.Path) +
			` RETURNING *
)
INSERT INTO global_defs (repo, commit_id, unit_type, unit, path, name, kind, file, updated_at, data, bow, doc) SELECT (` +
			arg(d.Repo) + `, ` +
			arg(d.CommitID) + `, ` +
			arg(d.UnitType) + `, ` +
			arg(d.Unit) + `, ` +
			arg(d.Path) + `, ` +
			arg(d.Name) + `, ` +
			arg(d.Kind) + `, ` +
			arg(d.File) + `, ` +
			`now(), ` +
			arg(data) + `, ` +
			arg(bow) + `, ` +
			arg(docstring) + `
) WHERE NOT EXISTS (SELECT * FROM upsert);`

		if _, err := graphDBH(ctx).Exec(upsertSQL, args...); err != nil {
			return err
		}
	}

	return nil
}

func (g *globalDefs) RefreshRefCounts(ctx context.Context, repos []string) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "GlobalDefs.RefreshRefCounts"); err != nil {
		return err
	}
	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	repoBindVars := make([]string, len(repos))
	for i, repo := range repos {
		repoBindVars[i] = arg(repo)
	}

	updateSQL := `UPDATE global_defs d
SET ref_ct = refs.ref_ct
FROM (SELECT def_repo, def_unit_type, def_unit, def_path, sum(count) ref_ct
      FROM global_refs
      WHERE def_repo in (` + strings.Join(repoBindVars, ",") + `)
      GROUP BY def_repo, def_unit_type, def_unit, def_path) refs
WHERE repo = def_repo AND unit_type = refs.def_unit_type AND unit = refs.def_unit AND path = refs.def_path;`

	_, err := graphDBH(ctx).Exec(updateSQL, args...)
	return err
}
