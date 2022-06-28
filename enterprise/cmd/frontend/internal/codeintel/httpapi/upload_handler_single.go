package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// handleEnqueueSinglePayload handles a non-multipart upload. This creates an upload record
// with state 'queued', proxies the data to the bundle manager, and returns the generated ID.
func (h *UploadHandler) handleEnqueueSinglePayload(ctx context.Context, uploadState uploadState, body io.Reader) (_ any, statusCode int, err error) {
	ctx, trace, endObservation := h.operations.handleEnqueueSinglePayload.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("statusCode", statusCode),
		}})
	}()

	tx, err := h.dbStore.Transact(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer func() { err = tx.Done(err) }()

	id, err := tx.InsertUpload(ctx, dbstore.Upload{
		Commit:            uploadState.commit,
		Root:              uploadState.root,
		RepositoryID:      uploadState.repositoryID,
		Indexer:           uploadState.indexer,
		IndexerVersion:    uploadState.indexerVersion,
		AssociatedIndexID: &uploadState.associatedIndexID,
		State:             "uploading",
		NumParts:          1,
		UploadedParts:     []int{0},
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	trace.Log(log.Int("uploadID", id))

	size, err := h.uploadStore.Upload(ctx, fmt.Sprintf("upload-%d.lsif.gz", id), body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	trace.Log(log.Int("gzippedUploadSize", int(size)))

	if err := tx.MarkQueued(ctx, id, &size); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	log15.Info(
		"codeintel.httpapi: enqueued upload",
		"id", id,
		"repository_id", uploadState.repositoryID,
		"commit", uploadState.commit,
	)

	// older versions of src-cli expect a string
	return struct {
		ID string `json:"id"`
	}{ID: strconv.Itoa(id)}, 0, nil
}
