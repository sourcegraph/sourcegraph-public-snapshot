package store

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestVulnerabilityMatchByID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	setupReferences(t, db)

	if _, err := store.InsertVulnerabilities(ctx, testVulnerabilities); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	if _, _, err := store.ScanMatches(ctx, 100); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	match, ok, err := store.VulnerabilityMatchByID(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error getting vulnerability match: %s", err)
	}
	if !ok {
		t.Fatalf("expected match to exist")
	}

	expectedMatch := shared.VulnerabilityMatch{
		ID:              3,
		UploadID:        52,
		VulnerabilityID: 1,
		AffectedPackage: badConfig,
	}
	if diff := cmp.Diff(expectedMatch, match); diff != "" {
		t.Errorf("unexpected vulnerability match (-want +got):\n%s", diff)
	}
}

func TestGetVulnerabilityMatches(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	/*
	 * Setup references is inserting seven (7) total references.
	 * Five (5) of them are vulnerable versions
	 * (three (3) for go-nacelle/config and two (2) for go-mockgen/xtools)
	 * the remaining two (2) of the references is of the fixed version.
	 */
	setupReferences(t, db)
	highVulnerabilityCount := 3
	mediumVulnerabilityCount := 2
	totalVulnerableVersionsInserted := highVulnerabilityCount + mediumVulnerabilityCount // 5

	highAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-nacelle/config",
		VersionConstraint: []string{"<= v1.2.5"},
	}
	mediumAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-mockgen/xtools",
		VersionConstraint: []string{"<= v1.3.5"},
	}

	mockVulnerabilities := []shared.Vulnerability{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPackages: []shared.AffectedPackage{highAffectedPackage}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPackages: []shared.AffectedPackage{mediumAffectedPackage}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
		{ID: 7, SourceID: "CVE-STU", Severity: "LOW"},
		{ID: 8, SourceID: "CVE-VWX", Severity: "LOW"},
		{ID: 9, SourceID: "CVE-Y&Z", Severity: "CRITICAL"},
	}

	if _, err := store.InsertVulnerabilities(ctx, mockVulnerabilities); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	if _, _, err := store.ScanMatches(ctx, 1000); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	/*
	 * Test
	 */

	args := shared.GetVulnerabilityMatchesArgs{Limit: 10, Offset: 0}
	matches, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	if len(matches) != totalVulnerableVersionsInserted {
		t.Errorf("unexpected total count. want=%d have=%d", len(matches), totalCount)
	}

	/*
	 * Test Severity filter
	 */

	t.Run("Test severity filter", func(t *testing.T) {
		args.Severity = "HIGH"
		high, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(high) != highVulnerabilityCount {
			t.Errorf("unexpected total count. want=%d have=%d", 3, totalCount)
		}

		args.Severity = "MEDIUM"
		medium, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(medium) != mediumVulnerabilityCount {
			t.Errorf("unexpected total count. want=%d have=%d", 2, totalCount)
		}
	})

	/*
	 * Test Language filter
	 */

	t.Run("Test language filter", func(t *testing.T) {
		args = shared.GetVulnerabilityMatchesArgs{Limit: 10, Offset: 0, Language: "go", Severity: ""}
		goMatches, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(goMatches) != totalVulnerableVersionsInserted {
			t.Errorf("unexpected total count. want=%d have=%d", 2, totalCount)
		}

		args = shared.GetVulnerabilityMatchesArgs{Limit: 10, Offset: 0, Language: "typescript", Severity: ""}
		typescriptMatches, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(typescriptMatches) != 0 {
			t.Errorf("unexpected total count. want=%d have=%d", 2, totalCount)
		}
	})

	/*
	 * Test Repository filter
	 */

	t.Run("Test repository filter", func(t *testing.T) {
		args = shared.GetVulnerabilityMatchesArgs{Limit: 10, Offset: 0, RepositoryName: "github.com/go-nacelle/config"}
		nacelleMatches, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(nacelleMatches) != highVulnerabilityCount {
			t.Errorf("unexpected total count. want=%d have=%d", 2, totalCount)
		}

		args = shared.GetVulnerabilityMatchesArgs{Limit: 10, Offset: 0, RepositoryName: "github.com/go-mockgen/xtools"}
		xToolsMatches, totalCount, err := store.GetVulnerabilityMatches(ctx, args)
		if err != nil {
			t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
		}

		if len(xToolsMatches) != mediumVulnerabilityCount {
			t.Errorf("unexpected total count. want=%d have=%d", 2, totalCount)
		}
	})
}

func TestGetVulberabilityMatchesCountByRepository(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	/*
	 * Setup references is inserting seven (7) total references.
	 * Five (5) of them are vulnerable versions
	 * (three (3) for go-nacelle/config and two (2) for go-mockgen/xtools)
	 * the remaining two (2) of the references is of the fixed version.
	 */
	setupReferences(t, db)
	var highVulnerabilityCount int32 = 3
	var mediumVulnerabilityCount int32 = 2

	highAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-nacelle/config",
		VersionConstraint: []string{"<= v1.2.5"},
	}
	mediumAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-mockgen/xtools",
		VersionConstraint: []string{"<= v1.3.5"},
	}
	mockVulnerabilities := []shared.Vulnerability{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPackages: []shared.AffectedPackage{highAffectedPackage}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPackages: []shared.AffectedPackage{mediumAffectedPackage}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
	}

	if _, err := store.InsertVulnerabilities(ctx, mockVulnerabilities); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	if _, _, err := store.ScanMatches(ctx, 1000); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	// Test
	args := shared.GetVulnerabilityMatchesCountByRepositoryArgs{Limit: 10}
	grouping, totalCount, err := store.GetVulnerabilityMatchesCountByRepository(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error getting vulnerability matches: %s", err)
	}

	expectedMatches := []shared.VulnerabilityMatchesByRepository{
		{
			ID:             2,
			RepositoryName: "github.com/go-nacelle/config",
			MatchCount:     highVulnerabilityCount,
		},
		{
			ID:             75,
			RepositoryName: "github.com/go-mockgen/xtools",
			MatchCount:     mediumVulnerabilityCount,
		},
	}

	if diff := cmp.Diff(expectedMatches, grouping); diff != "" {
		t.Errorf("unexpected vulnerability matches (-want +got):\n%s", diff)
	}

	if totalCount != len(expectedMatches) {
		t.Errorf("unexpected total count. want=%d have=%d", len(expectedMatches), totalCount)
	}
}

func TestGetVulnerabilityMatchesSummaryCount(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)
	handle := basestore.NewWithHandle(db.Handle())

	/* Insert uploads for four (4) repositories */
	insertUploads(t, db,
		uploadsshared.Upload{ID: 50, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 51, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 52, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 53, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 54, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
		uploadsshared.Upload{ID: 55, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
		uploadsshared.Upload{ID: 56, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
		uploadsshared.Upload{ID: 57, RepositoryID: 90, RepositoryName: "github.com/testify/config"},
		uploadsshared.Upload{ID: 58, RepositoryID: 90, RepositoryName: "github.com/testify/config"},
		uploadsshared.Upload{ID: 59, RepositoryID: 90, RepositoryName: "github.com/testify/config"},
		uploadsshared.Upload{ID: 60, RepositoryID: 90, RepositoryName: "github.com/testify/config"},
		uploadsshared.Upload{ID: 61, RepositoryID: 90, RepositoryName: "github.com/testify/config"},
		uploadsshared.Upload{ID: 62, RepositoryID: 200, RepositoryName: "github.com/go-sentinel/config"},
		uploadsshared.Upload{ID: 63, RepositoryID: 200, RepositoryName: "github.com/go-sentinel/config"},
	)

	/*
	 * Insert ten (10) total vulnerable package reference.
	 *  - Three (3) are high severity
	 *  - Two (2) are medium severity
	 *  - Four (4) are critical severity
	 *  - Low (1) is low severity
	 */
	if err := handle.Exec(context.Background(), sqlf.Sprintf(`
		INSERT INTO lsif_references (scheme, name, version, dump_id)
		VALUES
			('gomod', 'github.com/go-nacelle/config', 'v1.2.3', 50), -- high vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.4', 51), -- high vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.5', 52), -- high vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.6', 53),
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.2', 54), -- medium vulnerability
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.3', 55), -- medium vulnerability
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.6', 56),
			('gomod', 'github.com/testify/config', 'v1.0.1', 57), -- critical vulnerability
			('gomod', 'github.com/testify/config', 'v1.0.2', 58), -- critical vulnerability
			('gomod', 'github.com/testify/config', 'v1.0.3', 59), -- critical vulnerability
			('gomod', 'github.com/testify/config', 'v1.0.5', 60), -- critical vulnerability
			('gomod', 'github.com/testify/config', 'v1.0.6', 61),
			('gomod', 'github.com/go-sentinel/config', 'v2.3.0', 62), -- low vulnerability
			('gomod', 'github.com/go-sentinel/config', 'v2.3.6', 63)
	`)); err != nil {
		t.Fatalf("failed to insert references: %s", err)
	}

	var critical int32 = 4
	var high int32 = 3
	var medium int32 = 2
	var low int32 = 1
	var totalRepos int32 = 4

	criticalAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "testify/config",
		VersionConstraint: []string{"<= v1.0.5"},
	}
	highAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-nacelle/config",
		VersionConstraint: []string{"<= v1.2.5"},
	}
	mediumAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-mockgen/xtools",
		VersionConstraint: []string{"<= v1.3.5"},
	}
	lowAffectedPackage := shared.AffectedPackage{
		Language:          "go",
		PackageName:       "go-sentinel/config",
		VersionConstraint: []string{"<= v2.3.5"},
	}
	mockVulnerabilities := []shared.Vulnerability{
		{ID: 1, SourceID: "CVE-ABC", Severity: "HIGH", AffectedPackages: []shared.AffectedPackage{highAffectedPackage}},
		{ID: 2, SourceID: "CVE-DEF", Severity: "HIGH"},
		{ID: 3, SourceID: "CVE-GHI", Severity: "HIGH"},
		{ID: 4, SourceID: "CVE-JKL", Severity: "MEDIUM", AffectedPackages: []shared.AffectedPackage{mediumAffectedPackage}},
		{ID: 5, SourceID: "CVE-MNO", Severity: "MEDIUM"},
		{ID: 6, SourceID: "CVE-PQR", Severity: "MEDIUM"},
		{ID: 7, SourceID: "CVE-STU", Severity: "LOW", AffectedPackages: []shared.AffectedPackage{lowAffectedPackage}},
		{ID: 8, SourceID: "CVE-VWX", Severity: "LOW"},
		{ID: 9, SourceID: "CVE-Y&Z", Severity: "CRITICAL", AffectedPackages: []shared.AffectedPackage{criticalAffectedPackage}},
	}

	if _, err := store.InsertVulnerabilities(ctx, mockVulnerabilities); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	if _, _, err := store.ScanMatches(ctx, 1000); err != nil {
		t.Fatalf("unexpected error inserting vulnerabilities: %s", err)
	}

	/*
	 * Test
	 */

	summaryCount, err := store.GetVulnerabilityMatchesSummaryCount(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting vulnerability matches summary counts: %s", err)
	}

	expectedSummaryCount := shared.GetVulnerabilityMatchesSummaryCounts{
		Critical:     critical,
		High:         high,
		Medium:       medium,
		Low:          low,
		Repositories: totalRepos,
	}

	if diff := cmp.Diff(expectedSummaryCount, summaryCount); diff != "" {
		t.Errorf("unexpected vulnerability matches summary counts (-want +got):\n%s", diff)
	}
}

func setupReferences(t *testing.T, db database.DB) {
	store := basestore.NewWithHandle(db.Handle())

	insertUploads(t, db,
		uploadsshared.Upload{ID: 50, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 51, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 52, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 53, RepositoryID: 2, RepositoryName: "github.com/go-nacelle/config"},
		uploadsshared.Upload{ID: 54, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
		uploadsshared.Upload{ID: 55, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
		uploadsshared.Upload{ID: 56, RepositoryID: 75, RepositoryName: "github.com/go-mockgen/xtools"},
	)

	if err := store.Exec(context.Background(), sqlf.Sprintf(`
		-- Insert five (5) total vulnerable package reference.
		INSERT INTO lsif_references (scheme, name, version, dump_id)
		VALUES
			('gomod', 'github.com/go-nacelle/config', 'v1.2.3', 50), -- vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.4', 51), -- vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.5', 52), -- vulnerability
			('gomod', 'github.com/go-nacelle/config', 'v1.2.6', 53),
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.2', 54), -- vulnerability
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.3', 55), -- vulnerability
			('gomod', 'github.com/go-mockgen/xtools', 'v1.3.6', 56)
	`)); err != nil {
		t.Fatalf("failed to insert references: %s", err)
	}
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t testing.TB, db database.DB, uploads ...uploadsshared.Upload) {
	for _, upload := range uploads {
		if upload.Commit == "" {
			upload.Commit = makeCommit(upload.ID)
		}
		if upload.State == "" {
			upload.State = "completed"
		}
		if upload.RepositoryID == 0 {
			upload.RepositoryID = 50
		}
		if upload.Indexer == "" {
			upload.Indexer = "lsif-go"
		}
		if upload.IndexerVersion == "" {
			upload.IndexerVersion = "latest"
		}
		if upload.UploadedParts == nil {
			upload.UploadedParts = []int{}
		}

		// Ensure we have a repo for the inner join in select queries
		insertRepo(t, db, upload.RepositoryID, upload.RepositoryName)

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				id,
				commit,
				root,
				uploaded_at,
				state,
				failure_message,
				started_at,
				finished_at,
				process_after,
				num_resets,
				num_failures,
				repository_id,
				indexer,
				indexer_version,
				num_parts,
				uploaded_parts,
				upload_size,
				associated_index_id,
				content_type,
				should_reindex
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.UploadedAt,
			upload.State,
			upload.FailureMessage,
			upload.StartedAt,
			upload.FinishedAt,
			upload.ProcessAfter,
			upload.NumResets,
			upload.NumFailures,
			upload.RepositoryID,
			upload.Indexer,
			upload.IndexerVersion,
			upload.NumParts,
			pq.Array(upload.UploadedParts),
			upload.UploadSize,
			upload.AssociatedIndexID,
			upload.ContentType,
			upload.ShouldReindex,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// insertRepo creates a repository record with the given id and name. If there is already a repository
// with the given identifier, nothing happens
func insertRepo(t testing.TB, db database.DB, id int, name string) {
	if name == "" {
		name = fmt.Sprintf("n-%d", id)
	}

	deletedAt := sqlf.Sprintf("NULL")
	if strings.HasPrefix(name, "DELETED-") {
		deletedAt = sqlf.Sprintf("%s", time.Unix(1587396557, 0).UTC())
	}
	insertRepoQuery := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
	)
	if _, err := db.ExecContext(context.Background(), insertRepoQuery.Query(sqlf.PostgresBindVar), insertRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}

	status := "cloned"
	if strings.HasPrefix(name, "DELETED-") {
		status = "not_cloned"
	}
	updateGitserverRepoQuery := sqlf.Sprintf(
		`UPDATE gitserver_repos SET clone_status = %s WHERE repo_id = %s`,
		status,
		id,
	)
	if _, err := db.ExecContext(context.Background(), updateGitserverRepoQuery.Query(sqlf.PostgresBindVar), updateGitserverRepoQuery.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting gitserver repository: %s", err)
	}
}
