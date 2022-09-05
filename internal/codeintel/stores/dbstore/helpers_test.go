package dbstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type printableRank struct{ value *int }

func (r printableRank) String() string {
	if r.value == nil {
		return "nil"
	}
	return strconv.Itoa(*r.value)
}

// makeCommit formats an integer as a 40-character git commit hash.
func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t testing.TB, db database.DB, uploads ...Upload) {
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
				associated_index_id
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
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
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting upload: %s", err)
		}
	}
}

func updateUploads(t testing.TB, db database.DB, uploads ...Upload) {
	for _, upload := range uploads {
		query := sqlf.Sprintf(`
			UPDATE lsif_uploads
			SET
				commit = COALESCE(NULLIF(%s, ''), commit),
				root = COALESCE(NULLIF(%s, ''), root),
				uploaded_at = COALESCE(NULLIF(%s, '0001-01-01 00:00:00+00'::timestamptz), uploaded_at),
				state = COALESCE(NULLIF(%s, ''), state),
				failure_message  = COALESCE(%s, failure_message),
				started_at = COALESCE(%s, started_at),
				finished_at = COALESCE(%s, finished_at),
				process_after = COALESCE(%s, process_after),
				num_resets = COALESCE(NULLIF(%s, 0), num_resets),
				num_failures = COALESCE(NULLIF(%s, 0), num_failures),
				repository_id = COALESCE(NULLIF(%s, 0), repository_id),
				indexer = COALESCE(NULLIF(%s, ''), indexer),
				indexer_version = COALESCE(NULLIF(%s, ''), indexer_version),
				num_parts = COALESCE(NULLIF(%s, 0), num_parts),
				uploaded_parts = COALESCE(NULLIF(%s, '{}'::integer[]), uploaded_parts),
				upload_size = COALESCE(%s, upload_size),
				associated_index_id = COALESCE(%s, associated_index_id)
			WHERE id = %s
		`,
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
			upload.ID)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while updating upload: %s", err)
		}
	}
}

func deleteUploads(t testing.TB, db database.DB, uploads ...int) {
	for _, upload := range uploads {
		query := sqlf.Sprintf(`DELETE FROM lsif_uploads WHERE id = %s`, upload)
		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while deleting upload: %s", err)
		}
	}
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

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name, deleted_at) VALUES (%s, %s, %s) ON CONFLICT (id) DO NOTHING`,
		id,
		name,
		deletedAt,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while upserting repository: %s", err)
	}
}

// insertPackages populates the lsif_packages table with the given packages.
func insertPackages(t testing.TB, store *Store, packages []shared.Package) {
	for _, pkg := range packages {
		if err := store.UpdatePackages(context.Background(), pkg.DumpID, []precise.Package{
			{
				Scheme:  pkg.Scheme,
				Name:    pkg.Name,
				Version: pkg.Version,
			},
		}); err != nil {
			t.Fatalf("unexpected error updating packages: %s", err)
		}
	}
}

// insertPackageReferences populates the lsif_references table with the given package references.
func insertPackageReferences(t testing.TB, store *Store, packageReferences []shared.PackageReference) {
	for _, packageReference := range packageReferences {
		if err := store.UpdatePackageReferences(context.Background(), packageReference.DumpID, []precise.PackageReference{
			{
				Package: precise.Package{
					Scheme:  packageReference.Scheme,
					Name:    packageReference.Name,
					Version: packageReference.Version,
				},
			},
		}); err != nil {
			t.Fatalf("unexpected error updating package references: %s", err)
		}
	}
}

// insertVisibleAtTip populates rows of the lsif_uploads_visible_at_tip table for the given repository
// with the given identifiers. Each upload is assumed to refer to the tip of the default branch. To mark
// an upload as protected (visible to _some_ branch) butn ot visible from the default branch, use the
// insertVisibleAtTipNonDefaultBranch method instead.
func insertVisibleAtTip(t testing.TB, db database.DB, repositoryID int, uploadIDs ...int) {
	insertVisibleAtTipInternal(t, db, repositoryID, true, uploadIDs...)
}

// insertVisibleAtTipNonDefaultBranch populates rows of the lsif_uploads_visible_at_tip table for the
// given repository with the given identifiers. Each upload is assumed to refer to the tip of a branch
// distinct from the default branch or a tag.
func insertVisibleAtTipNonDefaultBranch(t testing.TB, db database.DB, repositoryID int, uploadIDs ...int) {
	insertVisibleAtTipInternal(t, db, repositoryID, false, uploadIDs...)
}

func insertVisibleAtTipInternal(t testing.TB, db database.DB, repositoryID int, isDefaultBranch bool, uploadIDs ...int) {
	var rows []*sqlf.Query
	for _, uploadID := range uploadIDs {
		rows = append(rows, sqlf.Sprintf("(%s, %s, %s)", repositoryID, uploadID, isDefaultBranch))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_uploads_visible_at_tip (repository_id, upload_id, is_default_branch) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating uploads visible at tip: %s", err)
	}
}

// insertNearestUploads populates the lsif_nearest_uploads table with the given upload metadata.
func insertNearestUploads(t testing.TB, db database.DB, repositoryID int, uploads map[string][]commitgraph.UploadMeta) {
	var rows []*sqlf.Query
	for commit, uploadMetas := range uploads {
		uploadsByLength := make(map[int]int, len(uploadMetas))
		for _, uploadMeta := range uploadMetas {
			uploadsByLength[uploadMeta.UploadID] = int(uploadMeta.Distance)
		}

		serializedUploadMetas, err := json.Marshal(uploadsByLength)
		if err != nil {
			t.Fatalf("unexpected error marshalling uploads: %s", err)
		}

		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s)",
			repositoryID,
			dbutil.CommitBytea(commit),
			serializedUploadMetas,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nearest_uploads (repository_id, commit_bytea, uploads) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating commit graph: %s", err)
	}
}

//nolint:unparam // unparam complains that `repositoryID` always has same value across call-sites, but that's OK
func insertLinks(t testing.TB, db database.DB, repositoryID int, links map[string]commitgraph.LinkRelationship) {
	if len(links) == 0 {
		return
	}

	var rows []*sqlf.Query
	for commit, link := range links {
		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s, %s)",
			repositoryID,
			dbutil.CommitBytea(commit),
			dbutil.CommitBytea(link.AncestorCommit),
			link.Distance,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nearest_uploads_links (repository_id, commit_bytea, ancestor_commit_bytea, distance) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while updating links: %s %s", err, query.Query(sqlf.PostgresBindVar))
	}
}

func toCommitGraphView(uploads []Upload) *commitgraph.CommitGraphView {
	commitGraphView := commitgraph.NewCommitGraphView()
	for _, upload := range uploads {
		commitGraphView.Add(commitgraph.UploadMeta{UploadID: upload.ID}, upload.Commit, fmt.Sprintf("%s:%s", upload.Root, upload.Indexer))
	}

	return commitGraphView
}

func normalizeVisibleUploads(uploadMetas map[string][]commitgraph.UploadMeta) map[string][]commitgraph.UploadMeta {
	for _, uploads := range uploadMetas {
		sort.Slice(uploads, func(i, j int) bool {
			return uploads[i].UploadID-uploads[j].UploadID < 0
		})
	}

	return uploadMetas
}

func getUploadStates(db database.DB, ids ...int) (map[int]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(
		`SELECT id, state FROM lsif_uploads WHERE id IN (%s)`,
		sqlf.Join(intsToQueries(ids), ", "),
	)

	return scanStates(db.QueryContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...))
}

// scanStates scans pairs of id/states from the return value of `*Store.query`.
func scanStates(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	states := map[int]string{}
	for rows.Next() {
		var id int
		var state string
		if err := rows.Scan(&id, &state); err != nil {
			return nil, err
		}

		states[id] = strings.ToLower(state)
	}

	return states, nil
}

func dumpToUpload(expected Dump) Upload {
	return Upload{
		ID:                expected.ID,
		Commit:            expected.Commit,
		Root:              expected.Root,
		UploadedAt:        expected.UploadedAt,
		State:             expected.State,
		FailureMessage:    expected.FailureMessage,
		StartedAt:         expected.StartedAt,
		FinishedAt:        expected.FinishedAt,
		ProcessAfter:      expected.ProcessAfter,
		NumResets:         expected.NumResets,
		NumFailures:       expected.NumFailures,
		RepositoryID:      expected.RepositoryID,
		RepositoryName:    expected.RepositoryName,
		Indexer:           expected.Indexer,
		IndexerVersion:    expected.IndexerVersion,
		AssociatedIndexID: expected.AssociatedIndexID,
	}
}

func assertReferenceCounts(t *testing.T, store *Store, expectedReferenceCountsByID map[int]int) {
	referenceCountsByID, err := scanIntPairs(store.Query(context.Background(), sqlf.Sprintf(`SELECT id, reference_count FROM lsif_uploads`)))
	if err != nil {
		t.Fatalf("unexpected error querying reference counts: %s", err)
	}

	if diff := cmp.Diff(expectedReferenceCountsByID, referenceCountsByID); diff != "" {
		t.Errorf("unexpected reference count (-want +got):\n%s", diff)
	}
}
