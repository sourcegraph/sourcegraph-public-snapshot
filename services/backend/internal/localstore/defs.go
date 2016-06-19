package localstore

import (
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sqs/pbtypes"
)

type dbRepoRev struct {
	ID     int64  `db:"id"`
	Repo   string `db:"repo"`
	Commit string `db:"commit"`
}

type dbDef struct {
	// Rev is the foreign key to the repo_rev table
	Rev int64 `db:"rid"`

	// DefKey is the foreign key to the def_keys table
	DefKey int64 `db:"defid"`

	// These fields mirror the fields in sourcegraph.Def
	Name string `db:"name"`
	Kind string `db:"kind"`
	File string `db:"file"`
	Data []byte `db:"data"`

	// Language is the deteced language of the definition. This is used for language-filtered queries.
	Language dbLang `db:"language"`

	// RefCount is the computed number of references (internal + external) that refer to this definition.
	RefCount int `db:"ref_ct"`

	// UpdatedAt is the last time at which this row as updated in the DB.
	UpdatedAt *time.Time `db:"updated_at"`

	// BoW is a "bag of words" representation of the definition that is used for text-based searches.
	BoW string `db:"bow"`

	// Doc is the docstring attached to the definition.
	Doc string `db:"doc"`
}

type dbDefSearchResult struct {
	dbDef
	Score float64 `db:"score"`
}

func init() {
	GraphSchema.Map.AddTableWithName(dbRepoRev{}, "repo_revs").SetKeys(true, "id").SetUniqueTogether("repo", "commit")
	GraphSchema.Map.AddTableWithName(dbDef{}, "defs").SetKeys(false, "rid", "defid")
	for _, table := range []string{"defs"} {
		GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
			`ALTER TABLE `+table+` ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
			`ALTER TABLE `+table+` ALTER COLUMN ref_ct SET DEFAULT 0;`,
			`ALTER TABLE `+table+` ALTER COLUMN language TYPE smallint`,
			`CREATE INDEX `+table+`_bow_idx ON `+table+` USING gin(to_tsvector('english', bow));`,
			`CREATE INDEX `+table+`_doc_idx ON `+table+` USING gin(to_tsvector('english', doc));`,
			`CREATE INDEX `+table+`_name ON `+table+` USING btree (lower(name));`,
			`CREATE INDEX `+table+`_rid_idx ON `+table+` using btree (rid);`,
			`CREATE INDEX `+table+`_defid_idx ON `+table+` using btree (defid);`,
		)
	}
}

type defs struct{}

var _ store.Defs = (*defs)(nil)

func (s *defs) Search(ctx context.Context, op store.DefSearchOp) (*sourcegraph.SearchResultsList, error) {
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return gorp.PostgresDialect{}.BindVar(len(args) - 1)
	}

	if op.Opt == nil {
		op.Opt = &sourcegraph.SearchOptions{}
	}

	if len(op.Opt.NotRepos) > 0 {
		return nil, fmt.Errorf("NotRepos option currently unsupported")
	}

	if len(op.TokQuery) == 0 && len(op.Opt.Repos) == 0 {
		return &sourcegraph.SearchResultsList{}, nil
	}

	bowQuery := search.UserQueryToksToTSQuery(op.TokQuery)
	lastTok := ""
	if len(op.TokQuery) > 0 {
		lastTok = op.TokQuery[len(op.TokQuery)-1]
	}

	if len(op.Opt.Repos) != 1 || op.Opt.CommitID == "" {
		return nil, fmt.Errorf("must specify exactly one repository and commit ID")
	}
	rp, err := store.ReposFromContext(ctx).Get(ctx, op.Opt.Repos[0])
	if err != nil {
		return nil, err
	}
	rid, err := getRID(ctx, graphDBH(ctx), rp.URI, op.Opt.CommitID)
	if err != nil {
		return nil, fmt.Errorf("could not get repo rev ID: %s", err)
	}

	var scoreSQL string
	if bowQuery != "" {
		// The ranking critieron is the weighted sum of xref count,
		// text similarity score, and whether the last term matches
		// the name.
		scoreSQL = `5.0*log(10 + ref_ct) + 100.0*ts_rank(to_tsvector('english', bow), to_tsquery('english', ` + arg(bowQuery) + `)) + 100.0*((LOWER(name)=LOWER(` + arg(lastTok) + `))::int) score`
	} else {
		scoreSQL = `ref_ct score`
	}
	selectSQL := `SELECT rid, defid, name, kind, file, data, doc, ref_ct, ` + scoreSQL + ` FROM defs`
	var whereSQL, prefixSQL string
	{
		var wheres []string

		// Repository/commit filtering.
		wheres = append(wheres, `rid=`+arg(rid))

		// Language filtering.
		if len(op.Opt.Languages) > 0 {
			var l []string
			for _, language := range op.Opt.Languages {
				id, err := toDBLang(language)
				if err != nil {
					continue
				}
				l = append(l, arg(id))
			}
			wheres = append(wheres, `language IN (`+strings.Join(l, ", ")+`)`)
		}
		if len(op.Opt.NotLanguages) > 0 {
			var l []string
			for _, language := range op.Opt.NotLanguages {
				id, err := toDBLang(language)
				if err != nil {
					continue
				}
				l = append(l, arg(id))
			}
			wheres = append(wheres, `language NOT IN (`+strings.Join(l, ", ")+`)`)
		}

		if len(op.Opt.Kinds) > 0 {
			var kindList []string
			for _, kind := range op.Opt.Kinds {
				kindList = append(kindList, arg(kind))
			}
			wheres = append(wheres, `kind IN (`+strings.Join(kindList, ", ")+`)`)
		}
		if len(op.Opt.NotKinds) > 0 {
			var notKindList []string
			for _, kind := range op.Opt.NotKinds {
				notKindList = append(notKindList, arg(kind))
			}
			wheres = append(wheres, `kind NOT IN (`+strings.Join(notKindList, ", ")+`)`)
		}

		if len(op.TokQuery) == 1 { // special-case single token queries for performance
			wheres = append(wheres, `lower(name)=lower(`+arg(op.TokQuery[0])+`)`)

			// Skip matching for too less characters.
			if op.Opt.PrefixMatch && len(op.TokQuery[0]) > 2 {
				prefixSQL = ` OR to_tsquery('english', ` + arg(op.TokQuery[0]+":*") + `) @@ to_tsvector('english', bow)`
			}
		} else if bowQuery != "" {
			wheres = append(wheres, "bow != ''")
			if op.Opt.PrefixMatch {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery+":*")+`) @@ to_tsvector('english', bow)`)
			} else {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery)+`) @@ to_tsvector('english', bow)`)
			}
		}

		whereSQL = fmt.Sprint(`WHERE (`+strings.Join(wheres, ") AND (")+`)`) + prefixSQL
	}
	orderSQL := `ORDER BY score DESC`
	limitSQL := `LIMIT ` + arg(op.Opt.PerPageOrDefault())

	sql := strings.Join([]string{selectSQL, whereSQL, orderSQL, limitSQL}, "\n")

	var dbSearchResults []*dbDefSearchResult
	if _, err := graphDBH(ctx).Select(&dbSearchResults, sql, args...); err != nil {
		return nil, err
	}

	var results []*sourcegraph.DefSearchResult
	for _, d := range dbSearchResults {
		// convert dbDef to Def
		var def sourcegraph.Def
		{
			// TODO(beyang): a possible optimization is to do this as a JOIN in the DB.
			dk, err := getDefKey(ctx, graphDBH(ctx), d.DefKey)
			if err != nil {
				return nil, err
			}

			rv, err := getRepoRev(ctx, graphDBH(ctx), d.Rev)
			if err != nil {
				return nil, err
			}

			var data pbtypes.RawMessage
			data.Unmarshal(d.Data)
			def = sourcegraph.Def{
				Def: graph.Def{
					DefKey: graph.DefKey{
						Repo:     dk.Repo,
						CommitID: rv.Commit,
						UnitType: dk.UnitType,
						Unit:     dk.Unit,
						Path:     dk.Path,
					},

					Name: d.Name,
					Kind: d.Kind,
					File: d.File,

					Data: data,
				},
			}
			if d.Doc != "" {
				def.Docs = []*graph.DefDoc{{Format: "text/plain", Data: d.Doc}}
			}
		}

		// Critical permissions check. DO NOT REMOVE.
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalDefs.Search", def.Repo); err != nil {
			continue
		}

		results = append(results, &sourcegraph.DefSearchResult{
			Def:      def,
			RefCount: int32(d.RefCount),
			Score:    float32(d.Score),
		})
	}
	return &sourcegraph.SearchResultsList{DefResults: results}, nil
}

// UpdateFromSrclibStore is a stop-gap method. Eventually, defs will replace
// srclib store as the canonical storage for defs. Until then, the canonical
// storage is srclib store. UpdateFromSrclibStore will sync the data in defs to
// reflect what is in srclib store for a given (repo, commit).
func (s *defs) UpdateFromSrclibStore(ctx context.Context, op store.DefUpdateOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Defs.UpdateFromSrclibStore", op.Repo); err != nil {
		return err
	}

	repo, err := store.ReposFromContext(ctx).Get(ctx, op.Repo)
	if err != nil {
		return err
	}

	if len(op.CommitID) != 40 {
		return fmt.Errorf("commit ID must be 40 characters long, was: %q", op.CommitID)
	}

	defs_, err := store.GraphFromContext(ctx).Defs(
		sstore.ByRepoCommitIDs(sstore.Version{Repo: repo.URI, CommitID: op.CommitID}),
	)
	if err != nil {
		return err
	}

	op.Defs = defs_
	return s.Update(ctx, op)
}

func (s *defs) Update(ctx context.Context, op store.DefUpdateOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Defs.Update", op.Repo); err != nil {
		return err
	}

	repo, err := store.ReposFromContext(ctx).Get(ctx, op.Repo)
	if err != nil {
		return err
	}

	// Validate input
	if op.Repo == 0 || op.CommitID == "" {
		return fmt.Errorf("both op.Repo and op.CommitID must be non-empty")
	}
	if len(op.CommitID) != 40 {
		return fmt.Errorf("commit must be 40 characters long, was: %q", op.CommitID)
	}

	for _, def := range op.Defs {
		if def.Repo != "" && def.Repo != repo.URI {
			return fmt.Errorf("cannot update def with non-matching repo (%s != %s)", def.Repo, repo.URI)
		}
		if def.CommitID != "" && def.CommitID != op.CommitID {
			return fmt.Errorf("cannot update def with non-matching revision (%s != %s)", def.CommitID, op.CommitID)
		}
	}

	// KLUDGE to improve search quality. This info ideally would be emitted by srclib toolchains.
	var chosenDefs []*graph.Def
	for _, d := range op.Defs {
		if shouldIndex(d) {
			chosenDefs = append(chosenDefs, d)
		}
	}

	dbh := graphDBH(ctx)

	// Update def_keys
	if err := updateDefKeys(ctx, dbh, repo.URI, chosenDefs); err != nil {
		return err
	}

	defKeys := make(map[graph.DefKey]struct{})
	for _, def := range chosenDefs {
		var dk graph.DefKey = def.DefKey
		dk.Repo, dk.CommitID = repo.URI, ""
		defKeys[dk] = struct{}{}
	}
	dbDefKeys, err := getDefKeys(ctx, dbh, defKeys)
	if err != nil {
		return err
	}
	defKeyIDs := make(map[graph.DefKey]int64)
	for _, dk := range dbDefKeys {
		defKeyIDs[graph.DefKey{Repo: dk.Repo, UnitType: dk.UnitType, Unit: dk.Unit, Path: dk.Path}] = dk.ID
	}

	// Update repo_revs
	repoRevsInsertSQL := `INSERT INTO repo_revs(repo, commit) (SELECT $1 AS repo, $2 AS commit WHERE NOT EXISTS (SELECT 1 FROM repo_revs WHERE repo=$1 AND commit=$2))`
	if _, err := dbh.Exec(repoRevsInsertSQL, repo.URI, op.CommitID); err != nil {
		return fmt.Errorf("repo_rev update failed: %s", err)
	}
	repoRevID, err := dbh.SelectInt(`SELECT id FROM repo_revs WHERE repo=$1 AND commit=$2`, repo.URI, op.CommitID)
	if err != nil {
		return fmt.Errorf("repo_rev id fetch failed: %s", err)
	}

	// Compute defs to insert
	dbDefs := make([]*dbDef, len(chosenDefs))
	now := time.Now()
	for i, def := range chosenDefs {
		dk := def.DefKey
		dk.Repo, dk.CommitID = repo.URI, ""

		data, err := def.Data.Marshal()
		if err != nil {
			data = []byte{}
		}
		bow := strings.Join(search.BagOfWordsToTokens(search.BagOfWords(def)), " ")

		var docstring string
		if len(def.Docs) == 1 {
			docstring = def.Docs[0].Data
		} else {
			for _, candidate := range def.Docs {
				if candidate.Format == "" || strings.ToLower(candidate.Format) == "text/plain" {
					docstring = candidate.Data
				}
			}
		}

		// TODO(beyang): kludge. Should not rely on def formatter for this information.
		languageID, err := toDBLang(strings.ToLower(graph.PrintFormatter(def).Language()))
		if err != nil {
			log15.Warn("could not determine language for def", "def", def.Path, "repo", def.Repo)
		}

		dbDefs[i] = &dbDef{
			Rev:       repoRevID,
			DefKey:    defKeyIDs[dk],
			Name:      def.Name,
			Kind:      def.Kind,
			File:      def.File,
			Data:      data,
			Language:  languageID,
			RefCount:  0,
			UpdatedAt: &now,
			BoW:       bow,
			Doc:       docstring,
		}
	}

	// Update defs
	createTmpSQL := `CREATE TEMPORARY TABLE defnew (
rid bigint,
defid bigint,
name TEXT,
kind TEXT,
file TEXT,
data bytea,
language smallint,
ref_ct integer,
bow TEXT,
doc TEXT)
	ON COMMIT DROP;`

	deleteSQL := `DELETE FROM defs WHERE rid=$1`
	insertSQL := `INSERT INTO defs(rid, defid, name, kind, file, data, language, ref_ct, bow, doc, updated_at)
SELECT defnew.*, now()
FROM defnew;
`

	return dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if _, err := tx.Exec(createTmpSQL); err != nil {
			return fmt.Errorf("defnew create failed: %s", err)
		}

		copy, err := dbutil.Prepare(tx, pq.CopyIn("defnew", "rid", "defid", "name", "kind", "file", "data", "language", "ref_ct", "bow", "doc"))
		if err != nil {
			return fmt.Errorf("defnew copy prepare failed: %s", err)
		}
		for _, df := range dbDefs {
			if _, err := copy.Exec(df.Rev, df.DefKey, df.Name, df.Kind, df.File, df.Data, df.Language, df.RefCount, df.BoW, df.Doc); err != nil {
				return fmt.Errorf("defnew copy failed: %s", err)
			}
		}
		if _, err := copy.Exec(); err != nil {
			return fmt.Errorf("defnew final copy failed: %s", err)
		}
		if _, err := tx.Exec(deleteSQL, repoRevID); err != nil {
			return fmt.Errorf("defs delete failed: %s", err)
		}
		if _, err := tx.Exec(insertSQL); err != nil {
			return fmt.Errorf("defs insert failed: %s", err)
		}
		return nil
	})
}

func getDefKey(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbDefKey, error) {
	var d dbDefKey
	if err := dbh.SelectOne(&d, `SELECT * FROM def_keys WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

func getRID(ctx context.Context, dbh gorp.SqlExecutor, repo string, commit string) (int64, error) {
	return dbh.SelectInt(`SELECT id FROM repo_revs WHERE repo=$1 AND commit=$2`, repo, commit)
}

func getRepoRev(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbRepoRev, error) {
	var d dbRepoRev
	if err := dbh.SelectOne(&d, `SELECT * FROM repo_revs WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

func getDefKeys(ctx context.Context, dbh gorp.SqlExecutor, defKeys map[graph.DefKey]struct{}) ([]*dbDefKey, error) {
	var dbDefKeys []*dbDefKey
	createTmpSQL := `CREATE TEMPORARY TABLE def_keys_tmp (
repo TEXT,
unit_type TEXT,
unit TEXT,
path TEXT)
ON COMMIT DROP;`
	selectSQL := `SELECT * FROM def_keys WHERE (repo, unit_type, unit, path) IN (SELECT * from def_keys_tmp)`
	err := dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if _, err := tx.Exec(createTmpSQL); err != nil {
			return err
		}
		copy, err := dbutil.Prepare(tx, pq.CopyIn("def_keys_tmp", "repo", "unit_type", "unit", "path"))
		if err != nil {
			return fmt.Errorf("def_keys_tmp copy prepare failed: %s", err)
		}
		for defKey, _ := range defKeys {
			if _, err := copy.Exec(defKey.Repo, defKey.UnitType, defKey.Unit, defKey.Path); err != nil {
				return fmt.Errorf("def_keys_tmp copy failed: %s", err)
			}
		}
		if _, err := copy.Exec(); err != nil {
			return fmt.Errorf("def_keys_tmp final copy failed: %s", err)
		}
		if _, err := tx.Select(&dbDefKeys, selectSQL); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dbDefKeys, nil
}

// updateDefKeys, given a list of defs, inserts all missing defs into def_keys.
// It never deletes anything.
func updateDefKeys(ctx context.Context, dbh gorp.SqlExecutor, repo string, defs []*graph.Def) error {
	createTmpSQL := `CREATE TEMPORARY TABLE def_keys_tmp (
repo TEXT,
unit_type TEXT,
unit TEXT,
path TEXT)
ON COMMIT DROP;`
	insertSQL := `INSERT INTO def_keys(repo, unit_type, unit, path)
SELECT repo, unit_type, unit, path
FROM def_keys_tmp dkt
WHERE NOT EXISTS (
	SELECT id FROM def_keys dk where dk.repo=dkt.repo AND dk.unit_type=dkt.unit_type AND dk.unit=dkt.unit AND dk.path=dkt.path
)`

	return dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if _, err := tx.Exec(createTmpSQL); err != nil {
			return fmt.Errorf("def_keys_tmp create failed: %s", err)
		}
		copy, err := dbutil.Prepare(tx, pq.CopyIn("def_keys_tmp", "repo", "unit_type", "unit", "path"))
		if err != nil {
			return fmt.Errorf("def_keys_tmp prepare failed: %s", err)
		}
		for _, def := range defs {
			rp := def.Repo
			if rp == "" {
				rp = repo
			}
			if _, err := copy.Exec(rp, def.UnitType, def.Unit, def.Path); err != nil {
				return fmt.Errorf("def_keys_tmp copy failed: %s", err)
			}
		}
		if _, err := copy.Exec(); err != nil {
			return fmt.Errorf("def_keys_tmp final copy failed: %s", err)
		}
		if _, err := tx.Exec(insertSQL); err != nil {
			return fmt.Errorf("def_keys insert failed: %s", err)
		}
		return nil
	})
}
