pbckbge store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/sentinel/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestVulnerbbilityMbtchByID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	setupReferences(t, db)

	if _, err := store.InsertVulnerbbilities(ctx, testVulnerbbilities); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	if _, _, err := store.ScbnMbtches(ctx, 100); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	mbtch, ok, err := store.VulnerbbilityMbtchByID(ctx, 3)
	if err != nil {
		t.Fbtblf("unexpected error getting vulnerbbility mbtch: %s", err)
	}
	if !ok {
		t.Fbtblf("expected mbtch to exist")
	}

	expectedMbtch := shbred.VulnerbbilityMbtch{
		ID:              3,
		UplobdID:        52,
		VulnerbbilityID: 1,
		AffectedPbckbge: bbdConfig,
	}
	if diff := cmp.Diff(expectedMbtch, mbtch); diff != "" {
		t.Errorf("unexpected vulnerbbility mbtch (-wbnt +got):\n%s", diff)
	}
}

func TestGetVulnerbbilityMbtches(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	/*
	 * Setup references is inserting seven (7) totbl references.
	 * Five (5) of them bre vulnerbble versions
	 * (three (3) for go-nbcelle/config bnd two (2) for go-mockgen/xtools)
	 * the rembining two (2) of the references is of the fixed version.
	 */
	setupReferences(t, db)
	highVulnerbbilityCount := 3
	mediumVulnerbbilityCount := 2
	totblVulnerbbleVersionsInserted := highVulnerbbilityCount + mediumVulnerbbilityCount // 5

	highAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-nbcelle/config",
		VersionConstrbint: []string{"<= v1.2.5"},
	}
	mediumAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-mockgen/xtools",
		VersionConstrbint: []string{"<= v1.3.5"},
	}

	mockVulnerbbilities := []shbred.Vulnerbbility{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPbckbges: []shbred.AffectedPbckbge{highAffectedPbckbge}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPbckbges: []shbred.AffectedPbckbge{mediumAffectedPbckbge}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
		{ID: 7, SourceID: "CVE-STU", Severity: "LOW"},
		{ID: 8, SourceID: "CVE-VWX", Severity: "LOW"},
		{ID: 9, SourceID: "CVE-Y&Z", Severity: "CRITICAL"},
	}

	if _, err := store.InsertVulnerbbilities(ctx, mockVulnerbbilities); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	if _, _, err := store.ScbnMbtches(ctx, 1000); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	/*
	 * Test
	 */

	brgs := shbred.GetVulnerbbilityMbtchesArgs{Limit: 10, Offset: 0}
	mbtches, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
	if err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	if len(mbtches) != totblVulnerbbleVersionsInserted {
		t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(mbtches), totblCount)
	}

	/*
	 * Test Severity filter
	 */

	t.Run("Test severity filter", func(t *testing.T) {
		brgs.Severity = "HIGH"
		high, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(high) != highVulnerbbilityCount {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 3, totblCount)
		}

		brgs.Severity = "MEDIUM"
		medium, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(medium) != mediumVulnerbbilityCount {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 2, totblCount)
		}
	})

	/*
	 * Test Lbngubge filter
	 */

	t.Run("Test lbngubge filter", func(t *testing.T) {
		brgs = shbred.GetVulnerbbilityMbtchesArgs{Limit: 10, Offset: 0, Lbngubge: "go", Severity: ""}
		goMbtches, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(goMbtches) != totblVulnerbbleVersionsInserted {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 2, totblCount)
		}

		brgs = shbred.GetVulnerbbilityMbtchesArgs{Limit: 10, Offset: 0, Lbngubge: "typescript", Severity: ""}
		typescriptMbtches, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(typescriptMbtches) != 0 {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 2, totblCount)
		}
	})

	/*
	 * Test Repository filter
	 */

	t.Run("Test repository filter", func(t *testing.T) {
		brgs = shbred.GetVulnerbbilityMbtchesArgs{Limit: 10, Offset: 0, RepositoryNbme: "github.com/go-nbcelle/config"}
		nbcelleMbtches, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(nbcelleMbtches) != highVulnerbbilityCount {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 2, totblCount)
		}

		brgs = shbred.GetVulnerbbilityMbtchesArgs{Limit: 10, Offset: 0, RepositoryNbme: "github.com/go-mockgen/xtools"}
		xToolsMbtches, totblCount, err := store.GetVulnerbbilityMbtches(ctx, brgs)
		if err != nil {
			t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
		}

		if len(xToolsMbtches) != mediumVulnerbbilityCount {
			t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", 2, totblCount)
		}
	})
}

func TestGetVulberbbilityMbtchesCountByRepository(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	/*
	 * Setup references is inserting seven (7) totbl references.
	 * Five (5) of them bre vulnerbble versions
	 * (three (3) for go-nbcelle/config bnd two (2) for go-mockgen/xtools)
	 * the rembining two (2) of the references is of the fixed version.
	 */
	setupReferences(t, db)
	vbr highVulnerbbilityCount int32 = 3
	vbr mediumVulnerbbilityCount int32 = 2

	highAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-nbcelle/config",
		VersionConstrbint: []string{"<= v1.2.5"},
	}
	mediumAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-mockgen/xtools",
		VersionConstrbint: []string{"<= v1.3.5"},
	}
	mockVulnerbbilities := []shbred.Vulnerbbility{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPbckbges: []shbred.AffectedPbckbge{highAffectedPbckbge}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPbckbges: []shbred.AffectedPbckbge{mediumAffectedPbckbge}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
	}

	if _, err := store.InsertVulnerbbilities(ctx, mockVulnerbbilities); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	if _, _, err := store.ScbnMbtches(ctx, 1000); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	// Test
	brgs := shbred.GetVulnerbbilityMbtchesCountByRepositoryArgs{Limit: 10}
	grouping, totblCount, err := store.GetVulnerbbilityMbtchesCountByRepository(ctx, brgs)
	if err != nil {
		t.Fbtblf("unexpected error getting vulnerbbility mbtches: %s", err)
	}

	expectedMbtches := []shbred.VulnerbbilityMbtchesByRepository{
		{
			ID:             2,
			RepositoryNbme: "github.com/go-nbcelle/config",
			MbtchCount:     highVulnerbbilityCount,
		},
		{
			ID:             75,
			RepositoryNbme: "github.com/go-mockgen/xtools",
			MbtchCount:     mediumVulnerbbilityCount,
		},
	}

	if diff := cmp.Diff(expectedMbtches, grouping); diff != "" {
		t.Errorf("unexpected vulnerbbility mbtches (-wbnt +got):\n%s", diff)
	}

	if totblCount != len(expectedMbtches) {
		t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(expectedMbtches), totblCount)
	}
}

func TestGetVulnerbbilityMbtchesSummbryCount(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)
	hbndle := bbsestore.NewWithHbndle(db.Hbndle())

	/* Insert uplobds for four (4) repositories */
	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 50, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 51, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 52, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 53, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 54, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
		uplobdsshbred.Uplobd{ID: 55, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
		uplobdsshbred.Uplobd{ID: 56, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
		uplobdsshbred.Uplobd{ID: 57, RepositoryID: 90, RepositoryNbme: "github.com/testify/config"},
		uplobdsshbred.Uplobd{ID: 58, RepositoryID: 90, RepositoryNbme: "github.com/testify/config"},
		uplobdsshbred.Uplobd{ID: 59, RepositoryID: 90, RepositoryNbme: "github.com/testify/config"},
		uplobdsshbred.Uplobd{ID: 60, RepositoryID: 90, RepositoryNbme: "github.com/testify/config"},
		uplobdsshbred.Uplobd{ID: 61, RepositoryID: 90, RepositoryNbme: "github.com/testify/config"},
		uplobdsshbred.Uplobd{ID: 62, RepositoryID: 200, RepositoryNbme: "github.com/go-sentinel/config"},
		uplobdsshbred.Uplobd{ID: 63, RepositoryID: 200, RepositoryNbme: "github.com/go-sentinel/config"},
	)

	/*
	 * Insert ten (10) totbl vulnerbble pbckbge reference.
	 *  - Three (3) bre high severity
	 *  - Two (2) bre medium severity
	 *  - Four (4) bre criticbl severity
	 *  - Low (1) is low severity
	 */
	if err := hbndle.Exec(context.Bbckground(), sqlf.Sprintf(`
		INSERT INTO lsif_references (scheme, nbme, version, dump_id)
		VALUES
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.3', 50), -- high vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.4', 51), -- high vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.5', 52), -- high vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.6', 53),
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.2', 54), -- medium vulnerbbility
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.3', 55), -- medium vulnerbbility
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.6', 56),
			('gomod', 'github.com/testify/config', 'v1.0.1', 57), -- criticbl vulnerbbility
			('gomod', 'github.com/testify/config', 'v1.0.2', 58), -- criticbl vulnerbbility
			('gomod', 'github.com/testify/config', 'v1.0.3', 59), -- criticbl vulnerbbility
			('gomod', 'github.com/testify/config', 'v1.0.5', 60), -- criticbl vulnerbbility
			('gomod', 'github.com/testify/config', 'v1.0.6', 61),
			('gomod', 'github.com/go-sentinel/config', 'v2.3.0', 62), -- low vulnerbbility
			('gomod', 'github.com/go-sentinel/config', 'v2.3.6', 63)
	`)); err != nil {
		t.Fbtblf("fbiled to insert references: %s", err)
	}

	vbr criticbl int32 = 4
	vbr high int32 = 3
	vbr medium int32 = 2
	vbr low int32 = 1
	vbr totblRepos int32 = 4

	criticblAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "testify/config",
		VersionConstrbint: []string{"<= v1.0.5"},
	}
	highAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-nbcelle/config",
		VersionConstrbint: []string{"<= v1.2.5"},
	}
	mediumAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-mockgen/xtools",
		VersionConstrbint: []string{"<= v1.3.5"},
	}
	lowAffectedPbckbge := shbred.AffectedPbckbge{
		Lbngubge:          "go",
		PbckbgeNbme:       "go-sentinel/config",
		VersionConstrbint: []string{"<= v2.3.5"},
	}
	mockVulnerbbilities := []shbred.Vulnerbbility{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPbckbges: []shbred.AffectedPbckbge{highAffectedPbckbge}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPbckbges: []shbred.AffectedPbckbge{mediumAffectedPbckbge}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
		{ID: 7, SourceID: "CVE-STU", Severity: "LOW", AffectedPbckbges: []shbred.AffectedPbckbge{lowAffectedPbckbge}},
		{ID: 8, SourceID: "CVE-VWX", Severity: "LOW"},
		{ID: 9, SourceID: "CVE-Y&Z", Severity: "CRITICAL", AffectedPbckbges: []shbred.AffectedPbckbge{criticblAffectedPbckbge}},
	}

	if _, err := store.InsertVulnerbbilities(ctx, mockVulnerbbilities); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	if _, _, err := store.ScbnMbtches(ctx, 1000); err != nil {
		t.Fbtblf("unexpected error inserting vulnerbbilities: %s", err)
	}

	/*
	 * Test
	 */

	summbryCount, err := store.GetVulnerbbilityMbtchesSummbryCount(ctx)
	if err != nil {
		t.Fbtblf("unexpected error getting vulnerbbility mbtches summbry counts: %s", err)
	}

	expectedSummbryCount := shbred.GetVulnerbbilityMbtchesSummbryCounts{
		Criticbl:     criticbl,
		High:         high,
		Medium:       medium,
		Low:          low,
		Repositories: totblRepos,
	}

	if diff := cmp.Diff(expectedSummbryCount, summbryCount); diff != "" {
		t.Errorf("unexpected vulnerbbility mbtches summbry counts (-wbnt +got):\n%s", diff)
	}
}

func setupReferences(t *testing.T, db dbtbbbse.DB) {
	store := bbsestore.NewWithHbndle(db.Hbndle())

	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 50, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 51, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 52, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 53, RepositoryID: 2, RepositoryNbme: "github.com/go-nbcelle/config"},
		uplobdsshbred.Uplobd{ID: 54, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
		uplobdsshbred.Uplobd{ID: 55, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
		uplobdsshbred.Uplobd{ID: 56, RepositoryID: 75, RepositoryNbme: "github.com/go-mockgen/xtools"},
	)

	if err := store.Exec(context.Bbckground(), sqlf.Sprintf(`
		-- Insert five (5) totbl vulnerbble pbckbge reference.
		INSERT INTO lsif_references (scheme, nbme, version, dump_id)
		VALUES
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.3', 50), -- vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.4', 51), -- vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.5', 52), -- vulnerbbility
			('gomod', 'github.com/go-nbcelle/config', 'v1.2.6', 53),
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.2', 54), -- vulnerbbility
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.3', 55), -- vulnerbbility
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.6', 56)
	`)); err != nil {
		t.Fbtblf("fbiled to insert references: %s", err)
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
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, nbme, deleted_bt) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		nbme,
		deletedAt,
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
