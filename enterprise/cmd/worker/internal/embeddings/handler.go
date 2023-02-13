package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/sourcegraph/log"

	"github.com/grafana/regexp"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	embeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
}

var _ workerutil.Handler[*embeddingsbg.RepoEmbeddingJob] = &handler{}

var matchEverythingRegexp = regexp.MustCompile(``)

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *embeddingsbg.RepoEmbeddingJob) error {
	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	files, err := h.gitserverClient.ListFiles(ctx, nil, repo.Name, record.Revision, matchEverythingRegexp)
	if err != nil {
		return err
	}
	fmt.Println("FILES!", len(files), files[:10])

	rowMetadata := make([]embeddings.RepoEmbeddingRowMetadata, 0, 10)
	for _, file := range files[:10] {
		rowMetadata = append(rowMetadata, embeddings.RepoEmbeddingRowMetadata{FileName: file, StartLine: 0, EndLine: 10})
	}
	// TODO: Separate code & text indices
	// TODO: Up to a 1M embeddings
	dimension := 512
	index := embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata]{
		// RepoName:        repo.Name,
		// Revision:        revision,
		Embeddings:      getRandomEmbeddings(1000, dimension),
		ColumnDimension: dimension,
		RowMetadata:     rowMetadata,
	}

	indexJsonBytes, err := json.Marshal(index)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(indexJsonBytes)
	_, err = h.uploadStore.Upload(ctx, "index", bytesReader)
	return err
}

func getRandomEmbeddings(n int, d int) []float32 {
	embeddings := make([]float32, n*d)
	for i := 0; i < n*d; i++ {
		embeddings[i] = rand.Float32()
	}
	return embeddings
}
