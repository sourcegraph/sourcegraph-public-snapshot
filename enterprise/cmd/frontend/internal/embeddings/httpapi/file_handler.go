package httpapi

import (
	"net/http"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type FileHandler struct {
	logger     sglog.Logger
	operations *Operations
}

func NewFileHandler(operations *Operations) *FileHandler {
	return &FileHandler{
		logger:     sglog.Scoped("FileHandler", "Embeddings file REST API handler"),
		operations: operations,
	}
}

func (h *FileHandler) Upload() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the embedding index name from the request
		// ...

		// Parse the request body into a byte slice
		// bodyBytes, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	h.logger.Error(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// // Upload the file data to storage
		// if err := uploadStore.Upload(embeddingIndexName, bodyBytes); err != nil {
		// 	h.logger.Error(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// // Invalidate the cache for this embedding index
		// indexGetter.Invalidate(embeddingIndexName)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("It works"))
	})
}

type uploadResponse struct {
	Id string `json:"id"`
}

func (h *FileHandler) upload(r *http.Request) (resp uploadResponse, statusCode int, err error) {
	// ctx, _, endObservation := h.operations.upload.With(r.Context(), &err, observation.Args{})
	_, _, endObservation := h.operations.upload.With(r.Context(), &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("statusCode", statusCode),
		}})
	}()

	return
}
