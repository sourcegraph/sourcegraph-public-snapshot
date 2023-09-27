pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/commitgrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// insertNebrestUplobds populbtes the lsif_nebrest_uplobds tbble with the given uplobd metbdbtb.
func insertNebrestUplobds(t testing.TB, db dbtbbbse.DB, repositoryID int, uplobds mbp[string][]commitgrbph.UplobdMetb) {
	vbr rows []*sqlf.Query
	for commit, uplobdMetbs := rbnge uplobds {
		uplobdsByLength := mbke(mbp[int]int, len(uplobdMetbs))
		for _, uplobdMetb := rbnge uplobdMetbs {
			uplobdsByLength[uplobdMetb.UplobdID] = int(uplobdMetb.Distbnce)
		}

		seriblizedUplobdMetbs, err := json.Mbrshbl(uplobdsByLength)
		if err != nil {
			t.Fbtblf("unexpected error mbrshblling uplobds: %s", err)
		}

		rows = bppend(rows, sqlf.Sprintf(
			"(%s, %s, %s)",
			repositoryID,
			dbutil.CommitByteb(commit),
			seriblizedUplobdMetbs,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nebrest_uplobds (repository_id, commit_byteb, uplobds) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while updbting commit grbph: %s", err)
	}
}

//
//
//
//
//
//

// insertPbckbges populbtes the lsif_pbckbges tbble with the given pbckbges.
func insertPbckbges(t testing.TB, store Store, pbckbges []shbred.Pbckbge) {
	for _, pkg := rbnge pbckbges {
		if err := store.UpdbtePbckbges(context.Bbckground(), pkg.DumpID, []precise.Pbckbge{
			{
				Scheme:  pkg.Scheme,
				Mbnbger: pkg.Mbnbger,
				Nbme:    pkg.Nbme,
				Version: pkg.Version,
			},
		}); err != nil {
			t.Fbtblf("unexpected error updbting pbckbges: %s", err)
		}
	}
}

// insertPbckbgeReferences populbtes the lsif_references tbble with the given pbckbge references.
func insertPbckbgeReferences(t testing.TB, store Store, pbckbgeReferences []shbred.PbckbgeReference) {
	for _, pbckbgeReference := rbnge pbckbgeReferences {
		if err := store.UpdbtePbckbgeReferences(context.Bbckground(), pbckbgeReference.DumpID, []precise.PbckbgeReference{
			{
				Pbckbge: precise.Pbckbge{
					Scheme:  pbckbgeReference.Scheme,
					Mbnbger: pbckbgeReference.Mbnbger,
					Nbme:    pbckbgeReference.Nbme,
					Version: pbckbgeReference.Version,
				},
			},
		}); err != nil {
			t.Fbtblf("unexpected error updbting pbckbge references: %s", err)
		}
	}
}

// insertIndexes populbtes the lsif_indexes tbble with the given index models.
func insertIndexes(t testing.TB, db dbtbbbse.DB, indexes ...uplobdsshbred.Index) {
	for _, index := rbnge indexes {
		if index.Commit == "" {
			index.Commit = mbkeCommit(index.ID)
		}
		if index.Stbte == "" {
			index.Stbte = "completed"
		}
		if index.RepositoryID == 0 {
			index.RepositoryID = 50
		}
		if index.DockerSteps == nil {
			index.DockerSteps = []uplobdsshbred.DockerStep{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}
		if index.LocblSteps == nil {
			index.LocblSteps = []string{}
		}

		// Ensure we hbve b repo for the inner join in select queries
		insertRepo(t, db, index.RepositoryID, index.RepositoryNbme, true)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_indexes (
				id,
				commit,
				queued_bt,
				stbte,
				fbilure_messbge,
				stbrted_bt,
				finished_bt,
				process_bfter,
				num_resets,
				num_fbilures,
				repository_id,
				docker_steps,
				root,
				indexer,
				indexer_brgs,
				outfile,
				execution_logs,
				locbl_steps,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			index.ID,
			index.Commit,
			index.QueuedAt,
			index.Stbte,
			index.FbilureMessbge,
			index.StbrtedAt,
			index.FinishedAt,
			index.ProcessAfter,
			index.NumResets,
			index.NumFbilures,
			index.RepositoryID,
			pq.Arrby(index.DockerSteps),
			index.Root,
			index.Indexer,
			pq.Arrby(index.IndexerArgs),
			index.Outfile,
			pq.Arrby(index.ExecutionLogs),
			pq.Arrby(index.LocblSteps),
			index.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error while inserting index: %s", err)
		}
	}
}

// insertUplobds populbtes the lsif_uplobds tbble with the given uplobd models.
func insertUplobds(t testing.TB, db dbtbbbse.DB, uplobds ...shbred.Uplobd) {
	for _, uplobd := rbnge uplobds {
		if uplobd.Commit == "" {
			uplobd.Commit = mbkeCommit(uplobd.ID)
		}
		if uplobd.Stbte == "" {
			uplobd.Stbte = "completed"
		}
		if uplobd.RepositoryID == 0 {
			uplobd.RepositoryID = 50
		}
		if uplobd.Indexer == "" {
			uplobd.Indexer = "lsif-go"
		}
		if uplobd.IndexerVersion == "" {
			uplobd.IndexerVersion = "lbtest"
		}
		if uplobd.UplobdedPbrts == nil {
			uplobd.UplobdedPbrts = []int{}
		}

		// Ensure we hbve b repo for the inner join in select queries
		insertRepo(t, db, uplobd.RepositoryID, uplobd.RepositoryNbme, true)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uplobds (
				id,
				commit,
				root,
				uplobded_bt,
				stbte,
				fbilure_messbge,
				stbrted_bt,
				finished_bt,
				process_bfter,
				num_resets,
				num_fbilures,
				repository_id,
				indexer,
				indexer_version,
				num_pbrts,
				uplobded_pbrts,
				uplobd_size,
				bssocibted_index_id,
				content_type,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			uplobd.ID,
			uplobd.Commit,
			uplobd.Root,
			uplobd.UplobdedAt,
			uplobd.Stbte,
			uplobd.FbilureMessbge,
			uplobd.StbrtedAt,
			uplobd.FinishedAt,
			uplobd.ProcessAfter,
			uplobd.NumResets,
			uplobd.NumFbilures,
			uplobd.RepositoryID,
			uplobd.Indexer,
			uplobd.IndexerVersion,
			uplobd.NumPbrts,
			pq.Arrby(uplobd.UplobdedPbrts),
			uplobd.UplobdSize,
			uplobd.AssocibtedIndexID,
			uplobd.ContentType,
			uplobd.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error while inserting uplobd: %s", err)
		}
	}
}

// insertRepo crebtes b repository record with the given id bnd nbme. If there is blrebdy b repository
// with the given identifier, nothing hbppens
func insertRepo(t testing.TB, db dbtbbbse.DB, id int, nbme string, privbte bool) {
	if nbme == "" {
		nbme = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HbsPrefix(nbme, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, nbme, deleted_bt, privbte) VALUES (%s, %s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		nbme,
		deletedAt,
		privbte,
	)
	if _, err := db.ExecContext(context.Bbckground(), insertRepoQuery.Query(sqlf.PostgresBindVbr), insertRepoQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upserting repository: %s", err)
	}

	stbtus := "cloned"
	if strings.HbsPrefix(nbme, "DELETED-") {
		stbtus = "not_cloned"
	}
	updbteGitserverRepoQuery := sqlf.Sprintf(
		`UPDATE gitserver_repos SET clone_stbtus = %s WHERE repo_id = %s`,
		stbtus,
		id,
	)
	if _, err := db.ExecContext(context.Bbckground(), updbteGitserverRepoQuery.Query(sqlf.PostgresBindVbr), updbteGitserverRepoQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upserting gitserver repository: %s", err)
	}
}

// mbkeCommit formbts bn integer bs b 40-chbrbcter git commit hbsh.
func mbkeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

func getUplobdStbtes(db dbtbbbse.DB, ids ...int) (mbp[int]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(
		`SELECT id, stbte FROM lsif_uplobds WHERE id IN (%s)`,
		sqlf.Join(intsToQueries(ids), ", "),
	)

	return scbnStbtes(db.QueryContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...))
}

// scbnStbtes scbns pbirs of id/stbtes from the return vblue of `*Store.query`.
func scbnStbtes(rows *sql.Rows, queryErr error) (_ mbp[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	stbtes := mbp[int]string{}
	for rows.Next() {
		vbr id int
		vbr stbte string
		if err := rows.Scbn(&id, &stbte); err != nil {
			return nil, err
		}

		stbtes[id] = strings.ToLower(stbte)
	}

	return stbtes, nil
}

// consumeScbnner rebds bll vblues from the scbnner into memory.
func consumeScbnner(scbnner shbred.PbckbgeReferenceScbnner) (references []shbred.PbckbgeReference, _ error) {
	for {
		reference, exists, err := scbnner.Next()
		if err != nil {
			return nil, err
		}
		if !exists {
			brebk
		}

		references = bppend(references, reference)
	}
	if err := scbnner.Close(); err != nil {
		return nil, err
	}

	return references, nil
}

// intsToQueries converts b slice of ints into b slice of queries.
func intsToQueries(vblues []int) []*sqlf.Query {
	vbr queries []*sqlf.Query
	for _, vblue := rbnge vblues {
		queries = bppend(queries, sqlf.Sprintf("%d", vblue))
	}

	return queries
}

type printbbleRbnk struct{ vblue *int }

func (r printbbleRbnk) String() string {
	if r.vblue == nil {
		return "nil"
	}
	return strconv.Itob(*r.vblue)
}
