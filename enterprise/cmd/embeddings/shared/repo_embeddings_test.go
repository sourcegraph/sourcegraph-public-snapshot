package shared_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const insertEmbeddingFmtstr = `
	INSERT INTO embedding_versions (repo_id, revision)
	VALUES ($1, $2)
`

const createTempTable = `
	CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,
		version_id INTEGER NOT NULL REFERENCES embedding_versions(id),
		embedding vector(1536) NOT NULL,
		file_name TEXT NOT NULL,
		start_line INTEGER NOT NULL,
		end_line INTEGER NOT NULL,
		rank FLOAT NOT NULL
	)
`

const sampleQueriesFmtstr = `
	INSERT INTO embeddings_test (id, version_id, embedding, file_name, start_line, end_line, rank)
	SELECT id, version_id, embedding, '', 0, 0, 0.0
	FROM code_embeddings
	TABLESAMPLE SYSTEM ($1)
`

const copyRestFmtstr = `
	INSERT INTO embeddings_train (id, version_id, embedding, file_name, start_line, end_line, rank)
	SELECT id, version_id, embedding, '', 0, 0, 0.0
	FROM code_embeddings
	WHERE id NOT IN (SELECT id FROM embeddings_test)
`

const indexTrainingSet = `
	CREATE INDEX embeddings_train_idx
	ON embeddings_train
	USING ivfflat (embedding vector_cosine_ops)
	WITH (lists = %d)
`

const knn = `
	SELECT tr.id
	FROM embeddings_train AS tr
	ORDER BY tr.embedding <=> $1::vector
	LIMIT $2
`

func TestRestoreEmbeddings(t *testing.T) {
	logger := logtest.NoOp(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos().GetByName(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Handle().ExecContext(ctx, insertEmbeddingFmtstr, repo.ID, "revisionsha"); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := db.Handle().QueryRowContext(ctx, "SELECT current_database()").Scan(&name); err != nil {
		t.Fatal(err)
	}
	restoreCmd := exec.Command(
		"pg_restore",
		"--format", "custom",
		"--verbose",
		"--data-only",
		"--dbname", name,
		"--table", "code_embeddings",
		"/Users/cbart/embeddings.dump",
	)

	restoreCmd.Stdout = os.Stdout
	restoreCmd.Stderr = os.Stderr

	if err := restoreCmd.Run(); err != nil {
		t.Fatalf("Failed to restore the table dump: %v", err)
	}
	var count int
	if err := db.Handle().QueryRowContext(ctx, "SELECT COUNT(*) FROM code_embeddings").Scan(&count); err != nil {
		t.Fatalf("Failed to query the restored table: %v", err)
	}

	if count == 0 {
		t.Fatal("No records found in the restored table")
	}

	var testSetSampleSizePercent = 10 // how many embeddings to select as test set, 10 is 10% == 0.10
	var queryK = 8                    // looking for K nearest neighbors
	var numProbes = 5                 // how many indexed clusters should the query look at?
	var numTrainedClusters = 50
	var numTests = 100

	if _, err := db.Handle().ExecContext(ctx, fmt.Sprintf(createTempTable, "embeddings_test")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Handle().ExecContext(ctx, sampleQueriesFmtstr, testSetSampleSizePercent); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Handle().ExecContext(ctx, fmt.Sprintf(createTempTable, "embeddings_train")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Handle().ExecContext(ctx, copyRestFmtstr); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Handle().ExecContext(ctx, fmt.Sprintf(indexTrainingSet, numTrainedClusters)); err != nil {
		t.Fatal(err)
	}

	// Get a test set
	var queries []string
	func() {
		qs, err := db.Handle().QueryContext(ctx, `SELECT embedding FROM embeddings_test ORDER BY RANDOM() LIMIT $1`, numTests)
		if err != nil {
			t.Fatal(err)
		}
		defer qs.Close()
		for qs.Next() {
			var q string
			if err := qs.Scan(&q); err != nil {
				t.Fatal(err)
			}
			queries = append(queries, q)
		}
	}()

	var intersect int // true positives for index vs seq scan
	for _, q := range queries {
		query := q
		// KNN without an index
		actualNearestNeighbors := map[int]bool{}
		if err := db.WithTransact(ctx, func(tx database.DB) error {
			if _, err := tx.Handle().ExecContext(ctx, `SET LOCAL enable_indexscan = off`); err != nil {
				return err
			}
			if _, err := tx.Handle().ExecContext(ctx, `SET LOCAL enable_indexonlyscan = off`); err != nil {
				return err
			}
			if _, err := tx.Handle().ExecContext(ctx, `SET LOCAL enable_bitmapscan = off`); err != nil {
				return err
			}
			rs, err := tx.Handle().QueryContext(ctx, knn, query, queryK)
			if err != nil {
				return err
			}
			defer rs.Close()
			for rs.Next() {
				var id int
				if err := rs.Scan(&id); err != nil {
					return err
				}
				actualNearestNeighbors[id] = true
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		indexedNearestNeighbors := map[int]bool{}
		if err := db.WithTransact(ctx, func(tx database.DB) error {
			if _, err := tx.Handle().ExecContext(ctx, `SET LOCAL enable_seqscan = off`); err != nil {
				return err
			}
			if _, err := tx.Handle().ExecContext(ctx, fmt.Sprintf(`SET LOCAL ivfflat.probes = %d`, numProbes)); err != nil {
				return err
			}
			rs, err := tx.Handle().QueryContext(ctx, knn, query, queryK)
			if err != nil {
				return err
			}
			defer rs.Close()
			for rs.Next() {
				var id int
				if err := rs.Scan(&id); err != nil {
					return err
				}
				indexedNearestNeighbors[id] = true
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		for i := range actualNearestNeighbors {
			if indexedNearestNeighbors[i] {
				intersect++
			}
		}
	}
	t.Errorf("recall: %.2f (%d/%d)", float32(intersect)/float32(queryK*numTests), intersect, queryK*numTests)
}
