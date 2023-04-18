package contextdetection

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/log"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	contextdetectionbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
}

var _ workerutil.Handler[*contextdetectionbg.ContextDetectionEmbeddingJob] = &handler{}

const MAX_EMBEDDINGS_RETRIES = 3

func (h *handler) Handle(ctx context.Context, logger log.Logger, _ *contextdetectionbg.ContextDetectionEmbeddingJob) error {
	if !conf.EmbeddingsEnabled() {
		return errors.New("embeddings are not configured or disabled")
	}

	embeddingsClient := embed.NewEmbeddingsClient()

	messagesWithAdditionalContextMeanEmbedding, err := getContextDetectionMessagesMeanEmbedding(ctx, MESSAGES_WITH_ADDITIONAL_CONTEXT, embeddingsClient)
	if err != nil {
		return err
	}

	messagesWithoutAdditionalContextMeanEmbedding, err := getContextDetectionMessagesMeanEmbedding(ctx, MESSAGES_WITHOUT_ADDITIONAL_CONTEXT, embeddingsClient)
	if err != nil {
		return err
	}

	contextDetectionIndex := embeddings.ContextDetectionEmbeddingIndex{
		MessagesWithAdditionalContextMeanEmbedding:    messagesWithAdditionalContextMeanEmbedding,
		MessagesWithoutAdditionalContextMeanEmbedding: messagesWithoutAdditionalContextMeanEmbedding,
	}

	return embeddings.UploadIndex(ctx, h.uploadStore, embeddings.CONTEXT_DETECTION_INDEX_NAME, contextDetectionIndex)
}

func getContextDetectionMessagesMeanEmbedding(ctx context.Context, messages []string, client embed.EmbeddingsClient) ([]float32, error) {
	messagesEmbeddings, err := client.GetEmbeddingsWithRetries(ctx, messages, MAX_EMBEDDINGS_RETRIES)
	if err != nil {
		return nil, err
	}

	dimensions, err := client.GetDimensions()
	if err != nil {
		return nil, err
	}
	return getMeanEmbedding(len(messages), dimensions, messagesEmbeddings), nil
}

func getMeanEmbedding(nRows int, dimensions int, embeddings []float32) []float32 {
	meanEmbedding := make([]float32, dimensions)
	for i := 0; i < nRows; i++ {
		row := embeddings[i*dimensions : (i+1)*dimensions]
		for columnIdx, columnValue := range row {
			meanEmbedding[columnIdx] += columnValue
		}
	}
	for idx := range meanEmbedding {
		meanEmbedding[idx] = meanEmbedding[idx] / float32(nRows)
	}
	return meanEmbedding
}
