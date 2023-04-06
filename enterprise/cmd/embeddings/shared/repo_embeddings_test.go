package shared_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
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

const insertEmbeddingsFromIndexFmtstr = `
	INSERT INTO code_embeddings (version_id, embedding, file_name, start_line, end_line, rank)
	VALUES                      (1,          $1,        '',        0,          0,        0)
`

func decodeRepoEmbeddingIndex(dec *gob.Decoder) (*embeddings.RepoEmbeddingIndex, error) {
	rei := &embeddings.RepoEmbeddingIndex{}

	if err := dec.Decode(&rei.RepoName); err != nil {
		return nil, err
	}

	if err := dec.Decode(&rei.Revision); err != nil {
		return nil, err
	}

	for _, ei := range []*embeddings.EmbeddingIndex{&rei.CodeIndex, &rei.TextIndex} {
		if err := dec.Decode(&ei.ColumnDimension); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.RowMetadata); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.Ranks); err != nil {
			return nil, err
		}

		var numChunks int
		if err := dec.Decode(&numChunks); err != nil {
			return nil, err
		}

		ei.Embeddings = make([]float32, 0, numChunks*ei.ColumnDimension)
		for i := 0; i < numChunks; i++ {
			var embeddingSlice []float32
			if err := dec.Decode(&embeddingSlice); err != nil {
				return nil, err
			}
			ei.Embeddings = append(ei.Embeddings, embeddingSlice...)
		}
	}

	return rei, nil
}

// ~/.asdf/shims/go test -timeout 600s -run ^TestRestoreEmbeddings$ -v github.com/sourcegraph/sourcegraph/enterprise/cmd/embeddings/shared
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
	// var name string
	// if err := db.Handle().QueryRowContext(ctx, "SELECT current_database()").Scan(&name); err != nil {
	// 	t.Fatal(err)
	// }
	// restoreCmd := exec.Command(
	// 	"pg_restore",
	// 	"--format", "custom",
	// 	"--verbose",
	// 	"--data-only",
	// 	"--dbname", name,
	// 	"--table", "code_embeddings",
	// 	// pg_dump --format custom --blobs --verbose --file embeddings.dump --table code_embeddings sourcegraph
	// 	"/Users/cbart/embeddings.dump",
	// )

	// restoreCmd.Stdout = os.Stdout
	// restoreCmd.Stderr = os.Stderr

	// if err := restoreCmd.Run(); err != nil {
	// 	t.Fatalf("Failed to restore the table dump: %v", err)
	// }
	var index *embeddings.RepoEmbeddingIndex
	func() {
		// MEGAREPO!
		//f, err := os.Open("/Users/cbart/.sourcegraph-dev/data/blobstore-go/buckets/embeddings/github_com_sgtest_megarepo_b825625a0feb5dc46656b300e77b7d8e.embeddingindex")
		// SOURCEGRAPH
		f, err := os.Open("/Users/cbart/.sourcegraph-dev/data/blobstore-go/buckets/embeddings/github_com_sourcegraph_sourcegraph_cf360e12ff91b2fc199e75aef4ff6744.embeddingindex")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if index, err = decodeRepoEmbeddingIndex(gob.NewDecoder(f)); err != nil {
			t.Fatal(err)
		}
	}()
	c := index.CodeIndex
	for i := range c.RowMetadata {
		var b bytes.Buffer
		fmt.Fprintf(&b, "[")
		var notFirst bool
		for _, f := range c.Embeddings[i*c.ColumnDimension : (i+1)*c.ColumnDimension] {
			if notFirst {
				fmt.Fprint(&b, ",")
			}
			notFirst = true
			fmt.Fprintf(&b, "%.9f", f)
		}
		fmt.Fprintf(&b, "]")
		if _, err := db.Handle().ExecContext(ctx, insertEmbeddingsFromIndexFmtstr, b.String()); err != nil {
			t.Fatal(err)
		}
	}

	var count int
	if err := db.Handle().QueryRowContext(ctx, "SELECT COUNT(*) FROM code_embeddings").Scan(&count); err != nil {
		t.Fatalf("Failed to query the restored table: %v", err)
	}

	if count == 0 {
		t.Fatal("No records found in the restored table")
	}

	// PARAMETERS
	var testSetSampleSizePercent = 10 // how many embeddings to select as test set, 10 is 10% == 0.10
	var queryK = 8                    // looking for K nearest neighbors
	var numProbes = 5                 // how many indexed clusters should the query look at?
	var numTrainedClusters = 50       // how many clusters does index train, this should be vectors/1000 if vectors < 1000 else sqrt(vectors)
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

	// For each of the sampled queries count true positives. True positive / total neighbors is recall.
	var truePositives int
	var seqScanTimesMs []int
	var indexScanTimesMs []int
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
			start := time.Now()
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
			seqScanTimesMs = append(seqScanTimesMs, int(time.Now().Sub(start).Milliseconds()))
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
			start := time.Now()
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
			indexScanTimesMs = append(indexScanTimesMs, int(time.Now().Sub(start).Milliseconds()))
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		for i := range actualNearestNeighbors {
			if indexedNearestNeighbors[i] {
				truePositives++
			}
		}
	}
	t.Logf("parameters: k=%d probes=%d clusters=%d", queryK, numProbes, numTrainedClusters)
	t.Logf("recall: %.2f (%d/%d)", float32(truePositives)/float32(queryK*numTests), truePositives, queryK*numTests)
	t.Logf("using seq scan (ms): med = %d, 90th = %d, max = %d", median(seqScanTimesMs), percentile(seqScanTimesMs, 0.9), max(seqScanTimesMs))
	t.Logf("using idx scan (ms): med = %d, 90th = %d, max = %d", median(indexScanTimesMs), percentile(indexScanTimesMs, 0.9), max(indexScanTimesMs))
}

func median(nums []int) int {
	sort.Ints(nums)
	return nums[len(nums)/2]
}

func percentile(nums []int, p float64) int {
	sort.Ints(nums)
	index := int(float64(len(nums)-1) * p)
	return nums[index]
}

func max(nums []int) int {
	max := int(math.MinInt)
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}
