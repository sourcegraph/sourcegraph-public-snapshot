package uploadhandler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadState[T any] struct {
	uploadID         int
	suppliedUploadID bool
	numParts         int
	uploadedParts    []int
	multipart        bool
	// suppliedIndex is true if the part index was supplied in the query parameters.
	suppliedIndex bool
	// index is 0-based part number for multi-part uploads
	index            int
	done             bool
	uncompressedSize *int64
	metadata         T
}

func (uploadState *uploadState[T]) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("uploadID", uploadState.uploadID),
		attribute.Int("numParts", uploadState.numParts),
		attribute.Int("numUploadedParts", len(uploadState.uploadedParts)),
		attribute.Bool("multipart", uploadState.multipart),
		attribute.Bool("suppliedIndex", uploadState.suppliedIndex),
		attribute.Int("index", uploadState.index),
		attribute.Bool("done", uploadState.done),
		attribute.String("metadata", fmt.Sprintf("%#v", uploadState.metadata)),
	}
}

// constructUploadState reads the query args of the given HTTP request and populates an upload state object.
// This function should be used instead of reading directly from the request as the upload state's fields are
// backfilled/denormalized from the database, depending on the type of request.
func (h *UploadHandler[T]) constructUploadState(ctx context.Context, r *http.Request) (uploadState[T], int, error) {
	uploadState := uploadState[T]{
		suppliedUploadID: hasQuery(r, "uploadId"),
		uploadID:         getQueryInt(r, "uploadId"),
		suppliedIndex:    hasQuery(r, "index"),
		index:            getQueryInt(r, "index"),
		done:             hasQuery(r, "done"),
	}

	if !uploadState.suppliedUploadID {
		return h.hydrateUploadStateFromRequest(ctx, r, uploadState)
	}

	// An upload identifier was supplied; this is a subsequent request of a multi-part
	// upload. Fetch the upload record to ensure that it hasn't since been deleted by
	// the user.
	upload, exists, err := h.dbStore.GetUploadByID(ctx, uploadState.uploadID)
	if err != nil {
		return uploadState, http.StatusInternalServerError, err
	}
	if !exists {
		return uploadState, http.StatusNotFound, errors.Errorf("upload not found")
	}

	// Stash all fields given in the initial request
	uploadState.numParts = upload.NumParts
	uploadState.uploadedParts = upload.UploadedParts
	uploadState.uncompressedSize = upload.UncompressedSize
	uploadState.metadata = upload.Metadata

	return uploadState, 0, nil
}

func (h *UploadHandler[T]) hydrateUploadStateFromRequest(ctx context.Context, r *http.Request, uploadState uploadState[T]) (uploadState[T], int, error) {
	uncompressedSize := new(int64)
	if size := r.Header.Get("X-Uncompressed-Size"); size != "" {
		parsedSize, err := strconv.ParseInt(size, 10, 64)
		if err != nil {
			return uploadState, http.StatusUnprocessableEntity, errors.New("the header `X-Uncompressed-Size` must be an integer")
		}

		*uncompressedSize = parsedSize
	}

	metadata, statusCode, err := h.metadataFromRequest(ctx, r)
	if err != nil {
		return uploadState, statusCode, err
	}

	uploadState.multipart = hasQuery(r, "multiPart")
	uploadState.numParts = getQueryInt(r, "numParts")
	uploadState.uncompressedSize = uncompressedSize
	uploadState.metadata = metadata

	return uploadState, 0, nil
}

func hasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Get(name) != ""
}

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func getQueryInt(r *http.Request, name string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(name))
	return value
}
