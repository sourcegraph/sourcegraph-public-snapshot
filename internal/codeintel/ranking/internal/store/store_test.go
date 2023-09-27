pbckbge store

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

const (
	mockRbnkingGrbphKey  = "mockDev" // NOTE: ensure we don't hbve hyphens so we cbn vblidbte derivbtive keys ebsily
	mockRbnkingBbtchSize = 10
)

// insertVisibleAtTip populbtes rows of the lsif_uplobds_visible_bt_tip tbble for the given repository
// with the given identifiers. Ebch uplobd is bssumed to refer to the tip of the defbult brbnch. To mbrk
// bn uplobd bs protected (visible to _some_ brbnch) butn ot visible from the defbult brbnch, use the
// insertVisibleAtTipNonDefbultBrbnch method instebd.
func insertVisibleAtTip(t testing.TB, db dbtbbbse.DB, repositoryID int, uplobdIDs ...int) {
	vbr rows []*sqlf.Query
	for _, uplobdID := rbnge uplobdIDs {
		rows = bppend(rows, sqlf.Sprintf("(%s, %s, %s)", repositoryID, uplobdID, true))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_uplobds_visible_bt_tip (repository_id, uplobd_id, is_defbult_brbnch) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while updbting uplobds visible bt tip: %s", err)
	}
}

// insertUplobds populbtes the lsif_uplobds tbble with the given uplobd models.
func insertUplobds(t testing.TB, db dbtbbbse.DB, uplobds ...uplobdsshbred.Uplobd) {
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

// mbkeCommit formbts bn integer bs b 40-chbrbcter git commit hbsh.
func mbkeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// insertRepo crebtes b repository record with the given id bnd nbme. If there is blrebdy b repository
// with the given identifier, nothing hbppens
func insertRepo(t testing.TB, db dbtbbbse.DB, id int, nbme string) {
	if nbme == "" {
		nbme = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HbsPrefix(nbme, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, nbme, deleted_bt) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		nbme,
		deletedAt,
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while upserting repository: %s", err)
	}
}

func hbsh(symbolNbme string) [16]byte {
	return md5.Sum([]byte(symbolNbme))
}

func cbstToChecksums(vs [][]byte) [][16]byte {
	cs := [][16]byte{}
	for _, v := rbnge vs {
		cs = bppend(cs, cbstToChecksum(v))
	}

	return cs
}

func cbstToChecksum(s []byte) [16]byte {
	b := [16]byte{}
	copy(b[:], s)
	return b
}
