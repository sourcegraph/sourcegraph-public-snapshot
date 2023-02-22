package contextdetection

import (
	"bytes"
	"context"
	"encoding/json"

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
	config := conf.Get().Embeddings
	if config == nil || !config.Enabled {
		return errors.New("embeddings are not configured or disabled")
	}

	embeddingsClient := embed.NewEmbeddingsClient()

	messagesWithAdditionalContextIndex, err := getContextDetectionMessagesEmbeddingIndex(MESSAGES_WITH_ADDITIONAL_CONTEXT, embeddingsClient)
	if err != nil {
		return err
	}

	messagesWithoutAdditionalContextIndex, err := getContextDetectionMessagesEmbeddingIndex(MESSAGES_WITHOUT_ADDITIONAL_CONTEXT, embeddingsClient)
	if err != nil {
		return err
	}

	contextDetectionIndex := embeddings.ContextDetectionEmbeddingIndex{
		MessagesWithAdditionalContextIndex:    messagesWithAdditionalContextIndex,
		MessagesWithoutAdditionalContextIndex: messagesWithoutAdditionalContextIndex,
	}

	indexJsonBytes, err := json.Marshal(contextDetectionIndex)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(indexJsonBytes)
	_, err = h.uploadStore.Upload(ctx, embeddings.CONTEXT_DETECTION_INDEX_NAME, bytesReader)
	return err
}

func getContextDetectionMessagesEmbeddingIndex(messages []string, client embed.EmbeddingsClient) (embeddings.EmbeddingIndex[embeddings.ContextDetectionEmbeddingRowMetadata], error) {
	metadata := make([]embeddings.ContextDetectionEmbeddingRowMetadata, len(messages))
	for idx, message := range messages {
		metadata[idx] = embeddings.ContextDetectionEmbeddingRowMetadata{Message: message}
	}

	messagesEmbeddings, err := client.GetEmbeddingsWithRetries(messages, MAX_EMBEDDINGS_RETRIES)
	if err != nil {
		return embeddings.EmbeddingIndex[embeddings.ContextDetectionEmbeddingRowMetadata]{}, err
	}

	dimensions, err := client.GetDimensions()
	if err != nil {
		return embeddings.EmbeddingIndex[embeddings.ContextDetectionEmbeddingRowMetadata]{}, err
	}
	return embeddings.EmbeddingIndex[embeddings.ContextDetectionEmbeddingRowMetadata]{
		Embeddings:      messagesEmbeddings,
		RowMetadata:     metadata,
		ColumnDimension: dimensions,
	}, nil
}
