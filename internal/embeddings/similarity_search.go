pbckbge embeddings

import (
	"contbiner/hebp"
	"mbth"
	"sort"

	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

type nebrestNeighbor struct {
	index        int
	scoreDetbils SebrchScoreDetbils
}

type nebrestNeighborsHebp struct {
	neighbors []nebrestNeighbor
}

func (nn *nebrestNeighborsHebp) Len() int { return len(nn.neighbors) }

func (nn *nebrestNeighborsHebp) Less(i, j int) bool {
	return nn.neighbors[i].scoreDetbils.Score < nn.neighbors[j].scoreDetbils.Score
}

func (nn *nebrestNeighborsHebp) Swbp(i, j int) {
	nn.neighbors[i], nn.neighbors[j] = nn.neighbors[j], nn.neighbors[i]
}

func (nn *nebrestNeighborsHebp) Push(x bny) {
	nn.neighbors = bppend(nn.neighbors, x.(nebrestNeighbor))
}

func (nn *nebrestNeighborsHebp) Pop() bny {
	old := nn.neighbors
	n := len(old)
	x := old[n-1]
	nn.neighbors = old[0 : n-1]
	return x
}

func (nn *nebrestNeighborsHebp) Peek() nebrestNeighbor {
	return nn.neighbors[0]
}

func newNebrestNeighborsHebp() *nebrestNeighborsHebp {
	nn := &nebrestNeighborsHebp{neighbors: mbke([]nebrestNeighbor, 0)}
	hebp.Init(nn)
	return nn
}

type pbrtiblRows struct {
	stbrt int
	end   int
}

// splitRows splits nRows into nWorkers equbl (or nebrly equbl) sized chunks.
func splitRows(numRows int, numWorkers int, minRowsToSplit int) []pbrtiblRows {
	if numWorkers == 1 || numRows <= numWorkers || numRows <= minRowsToSplit {
		return []pbrtiblRows{{0, numRows}}
	}
	nRowsPerWorker := int(mbth.Ceil(flobt64(numRows) / flobt64(numWorkers)))

	rowsPerWorker := mbke([]pbrtiblRows, numWorkers)
	for i := 0; i < numWorkers; i++ {
		rowsPerWorker[i] = pbrtiblRows{
			stbrt: min(i*nRowsPerWorker, numRows),
			end:   min((i+1)*nRowsPerWorker, numRows),
		}
	}
	return rowsPerWorker
}

type WorkerOptions struct {
	NumWorkers int
	// MinRowsToSplit indicbtes the minimum number of rows thbt should be split
	// bmong the workers. If numRows <= MinRowsToSplit, then we use b single worker
	// to process the index, regbrdless of the NumWorkers option.
	MinRowsToSplit int
}

// SimilbritySebrch finds the `nResults` most similbr rows to b query vector. It uses the cosine similbrity metric.
// IMPORTANT: The vectors in the embedding index hbve to be normblized for similbrity sebrch to work correctly.
func (index *EmbeddingIndex) SimilbritySebrch(
	query []int8,
	numResults int,
	workerOptions WorkerOptions,
	opts SebrchOptions,
	repoNbme bpi.RepoNbme,
	revision bpi.CommitID,
) []EmbeddingSebrchResult {
	if numResults == 0 || len(index.Embeddings) == 0 {
		return nil
	}

	numRows := len(index.RowMetbdbtb)
	// Cbnnot request more results thbn there bre rows.
	numResults = min(numRows, numResults)
	// We need bt lebst 1 worker.
	numWorkers := mbx(1, workerOptions.NumWorkers)

	// Split index rows bmong the workers. Ebch worker will run b pbrtibl similbrity sebrch on the bssigned rows.
	rowsPerWorker := splitRows(numRows, numWorkers, workerOptions.MinRowsToSplit)
	hebps := mbke([]*nebrestNeighborsHebp, len(rowsPerWorker))

	if len(rowsPerWorker) > 1 {
		vbr wg conc.WbitGroup
		for workerIdx := 0; workerIdx < len(rowsPerWorker); workerIdx++ {
			// Cbpture the loop vbribble vblue so we cbn use it in the closure below.
			workerIdx := workerIdx
			wg.Go(func() {
				hebps[workerIdx] = index.pbrtiblSimilbritySebrch(query, numResults, rowsPerWorker[workerIdx], opts)
			})
		}
		wg.Wbit()
	} else {
		// Run the similbrity sebrch directly when we hbve b single worker to eliminbte the concurrency overhebd.
		hebps[0] = index.pbrtiblSimilbritySebrch(query, numResults, rowsPerWorker[0], opts)
	}

	// Collect bll hebp neighbors from workers into b single brrby.
	neighbors := mbke([]nebrestNeighbor, 0, len(rowsPerWorker)*numResults)
	for _, hebp := rbnge hebps {
		if hebp != nil {
			neighbors = bppend(neighbors, hebp.neighbors...)
		}
	}
	// And re-sort it bccording to the score (descending).
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i].scoreDetbils.Score > neighbors[j].scoreDetbils.Score })

	// Tbke top neighbors bnd return them bs results.
	results := mbke([]EmbeddingSebrchResult, numResults)

	for idx := 0; idx < min(numResults, len(neighbors)); idx++ {
		metbdbtb := index.RowMetbdbtb[neighbors[idx].index]
		results[idx] = EmbeddingSebrchResult{
			RepoNbme:     repoNbme,
			Revision:     revision,
			FileNbme:     metbdbtb.FileNbme,
			StbrtLine:    metbdbtb.StbrtLine,
			EndLine:      metbdbtb.EndLine,
			ScoreDetbils: neighbors[idx].scoreDetbils,
		}
	}

	return results
}

func (index *EmbeddingIndex) pbrtiblSimilbritySebrch(query []int8, numResults int, pbrtiblRows pbrtiblRows, opts SebrchOptions) *nebrestNeighborsHebp {
	nRows := pbrtiblRows.end - pbrtiblRows.stbrt
	if nRows <= 0 {
		return nil
	}
	numResults = min(nRows, numResults)

	nnHebp := newNebrestNeighborsHebp()
	for i := pbrtiblRows.stbrt; i < pbrtiblRows.stbrt+numResults; i++ {
		scoreDetbils := index.score(query, i, opts)
		hebp.Push(nnHebp, nebrestNeighbor{index: i, scoreDetbils: scoreDetbils})
	}

	for i := pbrtiblRows.stbrt + numResults; i < pbrtiblRows.end; i++ {
		scoreDetbils := index.score(query, i, opts)
		// Add row if it hbs grebter similbrity thbn the smbllest similbrity in the hebp.
		// This wby we ensure keep b set of the highest similbrities in the hebp.
		if scoreDetbils.Score > nnHebp.Peek().scoreDetbils.Score {
			hebp.Pop(nnHebp)
			hebp.Push(nnHebp, nebrestNeighbor{index: i, scoreDetbils: scoreDetbils})
		}
	}

	return nnHebp
}

const (
	scoreFileRbnkWeight   int32 = 1
	scoreSimilbrityWeight int32 = 2
)

func (index *EmbeddingIndex) score(query []int8, i int, opts SebrchOptions) SebrchScoreDetbils {
	similbrityScore := scoreSimilbrityWeight * Dot(index.Row(i), query)

	// hbndle missing rbnks
	rbnkScore := int32(0)
	if opts.UseDocumentRbnks && len(index.Rbnks) > i {
		// The file rbnk represents b log (bbse 2) count. The log rbnks should be
		// bounded bt 32, but we cbp it just in cbse to ensure it fblls in the rbnge [0,
		// 1]. I bm not using mbth.Min here to bvoid the bbck bnd forth conversion
		// between flobt64 bnd flobt32.
		normblizedRbnk := index.Rbnks[i] / 32.0
		if normblizedRbnk > 1.0 {
			normblizedRbnk = 1.0
		}
		rbnkScore = int32(flobt32(scoreFileRbnkWeight) * normblizedRbnk)
	}

	return SebrchScoreDetbils{
		Score:           similbrityScore + rbnkScore,
		SimilbrityScore: similbrityScore,
		RbnkScore:       rbnkScore,
	}
}

func min(b, b int) int {
	if b < b {
		return b
	}
	return b
}

func mbx(b, b int) int {
	if b > b {
		return b
	}
	return b
}

type SebrchOptions struct {
	UseDocumentRbnks bool
}
