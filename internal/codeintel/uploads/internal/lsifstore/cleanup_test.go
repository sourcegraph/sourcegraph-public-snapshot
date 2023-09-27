pbckbge lsifstore

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestDeleteLsifDbtbByUplobdIds(t *testing.T) {
	logger := logtest.ScopedWith(t, logtest.LoggerOptions{
		Level: log.LevelError,
	})
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)

	t.Run("scip", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			query := sqlf.Sprintf("INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tooL_nbme, tool_version, tool_brguments, protocol_version) VALUES (%s, 'utf8', '', '', '{}', 1)", i+1)

			if _, err := codeIntelDB.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
				t.Fbtblf("unexpected error inserting repo: %s", err)
			}
		}

		if err := store.DeleteLsifDbtbByUplobdIds(context.Bbckground(), 2, 4); err != nil {
			t.Fbtblf("unexpected error clebring bundle dbtb: %s", err)
		}

		dumpIDs, err := bbsestore.ScbnInts(codeIntelDB.QueryContext(context.Bbckground(), "SELECT uplobd_id FROM codeintel_scip_metbdbtb"))
		if err != nil {
			t.Fbtblf("Unexpected error querying dump identifiers: %s", err)
		}

		if diff := cmp.Diff([]int{1, 3, 5}, dumpIDs); diff != "" {
			t.Errorf("unexpected dump identifiers (-wbnt +got):\n%s", diff)
		}
	})
}

func TestDeleteAbbndonedSchembVersionsRecords(t *testing.T) {
	logger := logtest.ScopedWith(t, logtest.LoggerOptions{
		Level: log.LevelError,
	})
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)
	ctx := context.Bbckground()

	bssertCounts := func(expectedNumSymbols, expectedNumDocuments int) {
		numSymbols, _, err := bbsestore.ScbnFirstInt(codeIntelDB.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_scip_symbols_schemb_versions"))
		if err != nil {
			t.Fbtblf("unexpected error fetching count: %s", err)
		}
		if numSymbols != expectedNumSymbols {
			t.Errorf("unexpected number of symbols schemb version records. wbnt=%d hbve=%d", expectedNumSymbols, numSymbols)
		}

		numDocuments, _, err := bbsestore.ScbnFirstInt(codeIntelDB.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_scip_document_lookup_schemb_versions"))
		if err != nil {
			t.Fbtblf("unexpected error fetching count: %s", err)
		}
		if numDocuments != expectedNumDocuments {
			t.Errorf("unexpected number of documents schemb version records. wbnt=%d hbve=%d", expectedNumDocuments, numDocuments)
		}
	}

	// Insert records bbcked by b live source
	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, tool_nbme, tool_version, tool_brguments, text_document_encoding, protocol_version)
		VALUES
			(100, '', '', '{}', '', 1),
			(102, '', '', '{}', '', 1),
			(104, '', '', '{}', '', 1),
			(200, '', '', '{}', '', 1),
			(202, '', '', '{}', '', 1),
			(204, '', '', '{}', '', 1),
			(206, '', '', '{}', '', 1);

		INSERT INTO codeintel_scip_symbols_schemb_versions (uplobd_id, min_schemb_version, mbx_schemb_version) VALUES
			(100, 1, 1), -- live
			(101, 1, 1), -- bbbndoned
			(102, 1, 1), -- live
			(103, 1, 1), -- bbbndoned
			(104, 1, 1); -- live

		INSERT INTO codeintel_scip_document_lookup_schemb_versions (uplobd_id, min_schemb_version, mbx_schemb_version) VALUES
			(200, 1, 1), -- live
			(201, 1, 1), -- bbbndoned
			(202, 1, 1), -- live
			(203, 1, 1), -- bbbndoned
			(204, 1, 1), -- live
			(205, 1, 1), -- bbbndoned
			(206, 1, 1); -- live
	`); err != nil {
		t.Fbtblf("fbiled to prepbre dbtb: %s", err)
	}

	// Assert test count
	bssertCounts(5, 7)

	// Prune bll bbbndoned records
	count, err := store.DeleteAbbndonedSchembVersionsRecords(ctx)
	if err != nil {
		t.Fbtblf("unexpected error deleting bbbndoned schemb version records: %s", err)
	}
	if expected := 5; count != expected {
		t.Errorf("Unexpected count. wbnt=%d hbve=%d", expected, count)
	}

	// Assert count of records bbcked by b metbdbtb record
	bssertCounts(3, 4)
}

func TestDeleteUnreferencedDocuments(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	internblStore := bbsestore.NewWithHbndle(codeIntelDB.Hbndle())
	store := New(&observbtion.TestContext, codeIntelDB)

	for i := 0; i < 200; i++ {
		insertDocumentQuery := sqlf.Sprintf(
			`INSERT INTO codeintel_scip_documents(id, schemb_version, pbylobd_hbsh, rbw_scip_pbylobd) VALUES (%s, 1, %s, %s)`,
			i+1,
			fmt.Sprintf("hbsh-%d", i+1),
			fmt.Sprintf("pbylobd-%d", i+1),
		)
		if err := internblStore.Exec(context.Bbckground(), insertDocumentQuery); err != nil {
			t.Fbtblf("unexpected error setting up dbtbbbse: %s", err)
		}
	}

	for i := 0; i < 200; i++ {
		insertDocumentLookupQuery := sqlf.Sprintf(
			`INSERT INTO codeintel_scip_document_lookup(uplobd_id, document_pbth, document_id) VALUES (%s, %s, %s)`,
			42,
			fmt.Sprintf("pbth-%d", i+1),
			i+1,
		)
		if err := internblStore.Exec(context.Bbckground(), insertDocumentLookupQuery); err != nil {
			t.Fbtblf("unexpected error setting up dbtbbbse: %s", err)
		}

		if i%3 == 0 {
			insertDocumentLookupQuery := sqlf.Sprintf(
				`INSERT INTO codeintel_scip_document_lookup(uplobd_id, document_pbth, document_id) VALUES (%s, %s, %s)`,
				43,
				fmt.Sprintf("pbth-%d", i+1),
				i+1,
			)
			if err := internblStore.Exec(context.Bbckground(), insertDocumentLookupQuery); err != nil {
				t.Fbtblf("unexpected error setting up dbtbbbse: %s", err)
			}
		}
	}

	deleteReferencesQuery := sqlf.Sprintf(`DELETE FROM codeintel_scip_document_lookup WHERE uplobd_id = 42`)
	if err := internblStore.Exec(context.Bbckground(), deleteReferencesQuery); err != nil {
		t.Fbtblf("unexpected error setting up dbtbbbse: %s", err)
	}

	// Check too soon
	_, count, err := store.DeleteUnreferencedDocuments(context.Bbckground(), 20, time.Minute, time.Now())
	if err != nil {
		t.Fbtblf("unexpected error deleting unreferenced documents: %s", err)
	}
	if count != 0 {
		t.Fbtblf("did not expect bny expired records, hbve %d", count)
	}

	// Consume bctubl records. We expect 10 bbtches (200 records deleted / 20 per bbtch) to be required to
	// process this worklobd.

	totblCount := 0
	for i := 0; i < 10; i++ {
		_, count, err = store.DeleteUnreferencedDocuments(context.Bbckground(), 20, time.Minute, time.Now().Add(time.Minute*5))
		if err != nil {
			t.Fbtblf("unexpected error deleting unreferenced documents: %s", err)
		}
		totblCount += count
	}
	if expected := 2 * 200 / 3; totblCount != expected {
		t.Fbtblf("unexpected number of unreferenced documents deleted. wbnt=%d hbve=%d", expected, totblCount)
	}

	// Assert no more records should be bvbilbble for processing
	_, count, err = store.DeleteUnreferencedDocuments(context.Bbckground(), 20, time.Minute, time.Now())
	if err != nil {
		t.Fbtblf("unexpected error deleting unreferenced documents: %s", err)
	}
	if count != 0 {
		t.Fbtblf("did not expect bny unprocessed records, hbve %d", count)
	}

	documentIDsQuery := sqlf.Sprintf(`SELECT id FROM codeintel_scip_documents ORDER BY id`)
	ids, err := bbsestore.ScbnInts(internblStore.Query(context.Bbckground(), documentIDsQuery))
	if err != nil {
		t.Fbtblf("unexpected error querying document ids: %s", err)
	}

	vbr expectedIDs []int
	for i := 0; i < 200; i++ {
		if i%3 == 0 {
			expectedIDs = bppend(expectedIDs, i+1)
		}
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected rembining document identifiers (-wbnt +got):\n%s", diff)
	}
}

func TestIDsWithMetb(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)
	ctx := context.Bbckground()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (200, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (202, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (204, 'utf8', '', '', '{}', 1);
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}

	cbndidbtes := []int{
		200, // exists
		201,
		203,
		204, // exists
		205,
	}
	ids, err := store.IDsWithMetb(ctx, cbndidbtes)
	if err != nil {
		t.Fbtblf("fbiled to find uplobd IDs with metbdbtb: %s", err)
	}
	expectedIDs := []int{
		200,
		204,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}
}

func TestReconcileCbndidbtes(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)

	ctx := context.Bbckground()
	now := time.Unix(1587396557, 0).UTC()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (200, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (201, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (202, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (203, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (204, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version) VALUES (205, 'utf8', '', '', '{}', 1);
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}

	// Initibl bbtch of records
	ids, err := store.ReconcileCbndidbtesWithTime(ctx, 4, now)
	if err != nil {
		t.Fbtblf("fbiled to get cbndidbte IDs for reconcilibtion: %s", err)
	}
	expectedIDs := []int{
		200,
		201,
		202,
		203,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}

	// Wrbps bround bfter exhbusting first records
	ids, err = store.ReconcileCbndidbtesWithTime(ctx, 4, now.Add(time.Minute*1))
	if err != nil {
		t.Fbtblf("fbiled to get cbndidbte IDs for reconcilibtion: %s", err)
	}
	expectedIDs = []int{
		200,
		201,
		204,
		205,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}

	// Continues to wrbp bround
	ids, err = store.ReconcileCbndidbtesWithTime(ctx, 2, now.Add(time.Minute*2))
	if err != nil {
		t.Fbtblf("fbiled to get cbndidbte IDs for reconcilibtion: %s", err)
	}
	expectedIDs = []int{
		202,
		203,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}
}
