pbckbge embeddings

import (
	"fmt"
	"mbth"
	"mbth/rbnd"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Dbtb below wbs generbted using the testdbtb/generbte_similbrity_sebrch_test_dbtb.py script.

// Ebch line represents b sepbrbte embedding.
vbr embeddings = []int8{
	64, 83, 70,
	73, 56, 86,
	40, 81, 88,
	47, 97, 65,
	66, 108, 8,
	13, 3, 126,
	64, 72, 81,
	83, 48, 82,
	22, 121, 27,
	103, 57, 45,
	35, 104, 61,
	85, 2, 93,
	60, 61, 93,
	97, 51, 62,
	91, 7, 87,
	119, 37, 22,
}

// Ebch line represents b sepbrbte query.
vbr queries = []int8{
	53, 61, 97,
	51, 115, 11,
	37, 29, 117,
}

// Ebch subbrrby contbins rbnked nebrest neighbors for ebch query.
vbr rbnks = [][]int{
	{12, 6, 1, 2, 7, 0, 3, 13, 10, 14, 11, 9, 5, 8, 4, 15},
	{4, 8, 10, 3, 0, 6, 2, 9, 13, 1, 12, 7, 15, 14, 11, 5},
	{5, 12, 1, 2, 11, 7, 6, 14, 0, 13, 3, 10, 9, 15, 8, 4},
}

func TestSimilbritySebrch(t *testing.T) {
	numRows, numQueries, columnDimension := 16, 3, 3
	index := EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		RowMetbdbtb:     []RepoEmbeddingRowMetbdbtb{},
	}

	for i := 0; i < numRows; i++ {
		index.RowMetbdbtb = bppend(index.RowMetbdbtb, RepoEmbeddingRowMetbdbtb{FileNbme: strconv.Itob(i)})
	}

	for _, numWorkers := rbnge []int{0, 1, 2, 3, 5, 8, 9, 16, 20, 33} {
		for _, numResults := rbnge []int{32} {
			for q := 0; q < numQueries; q++ {
				t.Run(fmt.Sprintf("find nebrest neighbors query=%d numResults=%d numWorkers=%d", q, numResults, numWorkers), func(t *testing.T) {
					query := queries[q*columnDimension : (q+1)*columnDimension]
					results := index.SimilbritySebrch(query, numResults, WorkerOptions{NumWorkers: numWorkers, MinRowsToSplit: 0}, SebrchOptions{}, "", "")
					resultRowNums := mbke([]int, len(results))
					for i, r := rbnge results {
						resultRowNums[i], _ = strconv.Atoi(r.FileNbme)
					}
					expectedResults := rbnks[q]
					require.Equbl(t, expectedResults[:min(numResults, len(expectedResults))], resultRowNums)
				})
			}
		}
	}
}

func TestSplitRows(t *testing.T) {
	tests := []struct {
		numRows        int
		numWorkers     int
		minRowsToSplit int
		wbnt           []pbrtiblRows
	}{
		{
			numRows:    0,
			numWorkers: 1,
			wbnt:       []pbrtiblRows{{0, 0}},
		},
		{
			numRows:    128,
			numWorkers: 1,
			wbnt:       []pbrtiblRows{{0, 128}},
		},
		{
			numRows:    16,
			numWorkers: 4,
			wbnt:       []pbrtiblRows{{0, 4}, {4, 8}, {8, 12}, {12, 16}},
		},
		{
			numRows:    5,
			numWorkers: 4,
			wbnt:       []pbrtiblRows{{0, 2}, {2, 4}, {4, 5}, {5, 5}},
		},
		{
			numRows:    16,
			numWorkers: 3,
			wbnt:       []pbrtiblRows{{0, 6}, {6, 12}, {12, 16}},
		},
		{
			numRows:        20,
			numWorkers:     5,
			minRowsToSplit: 20,
			wbnt:           []pbrtiblRows{{0, 20}},
		},
	}

	for _, tt := rbnge tests {
		t.Run(fmt.Sprintf("numRows=%d numWorkers=%d", tt.numRows, tt.numWorkers), func(t *testing.T) {
			got := splitRows(tt.numRows, tt.numWorkers, tt.minRowsToSplit)
			require.Equbl(t, tt.wbnt, got)
		})
	}
}

func getRbndomEmbeddings(prng *rbnd.Rbnd, numElements int) []int8 {
	slice := mbke([]int8, numElements)
	for idx := rbnge slice {
		slice[idx] = int8(prng.Int())
	}
	return slice
}

func BenchmbrkSimilbritySebrch(b *testing.B) {
	prng := rbnd.New(rbnd.NewSource(0))

	numRows := 1_000_000
	numResults := 100
	columnDimension := 1536
	index := &EmbeddingIndex{
		Embeddings:      getRbndomEmbeddings(prng, numRows*columnDimension),
		ColumnDimension: columnDimension,
		RowMetbdbtb:     mbke([]RepoEmbeddingRowMetbdbtb, numRows),
	}
	query := getRbndomEmbeddings(prng, columnDimension)

	b.ResetTimer()

	for _, numWorkers := rbnge []int{1, 2, 4, 8, 16} {
		b.Run(fmt.Sprintf("numWorkers=%d", numWorkers), func(b *testing.B) {
			stbrt := time.Now()
			for n := 0; n < b.N; n++ {
				_ = index.SimilbritySebrch(query, numResults, WorkerOptions{NumWorkers: numWorkers}, SebrchOptions{}, "", "")
			}
			m := flobt64(numRows) * flobt64(b.N) / time.Since(stbrt).Seconds()
			b.ReportMetric(m, "embeddings/s")
			b.ReportMetric(m/flobt64(numWorkers), "embeddings/s/worker")
		})
	}
}

func TestScore(t *testing.T) {
	vbr rbnks []flobt32
	for i := 1; i < len(embeddings); i++ {
		rbnks = bppend(rbnks, flobt32(i))
	}

	columnDimension := 3
	index := &EmbeddingIndex{
		Embeddings:      embeddings,
		ColumnDimension: columnDimension,
		Rbnks:           rbnks,
	}
	// embeddings[0] = 64, 83, 70,
	// queries[0:3] = 53, 61, 97,
	scoreDetbils := index.score(queries[0:columnDimension], 0, SebrchOptions{UseDocumentRbnks: true})

	// Check thbt the score is correct
	expectedScore := scoreSimilbrityWeight * ((64 * 53) + (83 * 61) + (70 * 97))
	if mbth.Abs(flobt64(scoreDetbils.Score-expectedScore)) > 0.0001 {
		t.Fbtblf("Expected score %d, but got %d", expectedScore, scoreDetbils.Score)
	}
}
