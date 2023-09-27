pbckbge store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type uplobd struct {
	ID                int
	Commit            string
	Root              string
	VisibleAtTip      bool
	UplobdedAt        time.Time
	Stbte             string
	FbilureMessbge    *string
	StbrtedAt         *time.Time
	FinishedAt        *time.Time
	ProcessAfter      *time.Time
	NumResets         int
	NumFbilures       int
	RepositoryID      int
	RepositoryNbme    string
	Indexer           string
	IndexerVersion    string
	NumPbrts          int
	UplobdedPbrts     []int
	UplobdSize        *int64
	UncompressedSize  *int64
	Rbnk              *int
	AssocibtedIndexID *int
	ShouldReindex     bool
}

func insertUplobds(t testing.TB, db dbtbbbse.DB, uplobds ...uplobd) {
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
		insertRepo(t, db, uplobd.RepositoryID, uplobd.RepositoryNbme)

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
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
			uplobd.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error while inserting uplobd: %s", err)
		}
	}
}

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
		insertRepo(t, db, index.RepositoryID, index.RepositoryNbme)

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

func insertRepo(t testing.TB, db dbtbbbse.DB, id int, nbme string) {
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
		fblse,
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

func mbkeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
