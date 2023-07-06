package httpapi

import (
	"net/http"

	sglog "github.com/sourcegraph/log"
)

type FileHandler struct {
	logger sglog.Logger
}

func NewFileHandler() *FileHandler {
	return &FileHandler{
		logger: sglog.Scoped("FileHandler", "Embeddings file REST API handler"),
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
