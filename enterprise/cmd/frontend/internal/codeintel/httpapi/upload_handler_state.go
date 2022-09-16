package httpapi

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadState struct {
	uploadID         int
	numParts         int
	uploadedParts    []int
	multipart        bool
	suppliedIndex    bool
	index            int
	done             bool
	uncompressedSize *int64
	metadata         codeintelUploadMetadata
}

type codeintelUploadMetadata struct {
	repositoryID      int
	commit            string
	root              string
	indexer           string
	indexerVersion    string
	associatedIndexID int
}

var revhashPattern = lazyregexp.New(`^[a-z0-9]{40}$`)

// constructUploadState reads the query args of the given HTTP request and populates an upload state object.
// This function should be used instead of reading directly from the request as the upload state's fields are
// backfilled/denormalized from the database, depending on the type of request.
func (h *UploadHandler) constructUploadState(ctx context.Context, r *http.Request) (uploadState, int, error) {
	uploadState := uploadState{
		uploadID:      getQueryInt(r, "uploadId"),
		suppliedIndex: hasQuery(r, "index"),
		index:         getQueryInt(r, "index"),
		done:          hasQuery(r, "done"),
	}

	if uploadState.uploadID == 0 {
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

	return h.hydrateUploadStateFromRecord(ctx, r, uploadState, upload)
}

func (h *UploadHandler) hydrateUploadStateFromRequest(ctx context.Context, r *http.Request, uploadState uploadState) (uploadState, int, error) {
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

func (h *UploadHandler) hydrateUploadStateFromRecord(_ context.Context, r *http.Request, uploadState uploadState, upload dbstore.Upload) (uploadState, int, error) {
	metadata, statusCode, err := h.metadataFromRecord(upload)
	if err != nil {
		return uploadState, statusCode, err
	}

	// Stash all fields given in the initial request
	uploadState.numParts = upload.NumParts
	uploadState.uploadedParts = upload.UploadedParts
	uploadState.uncompressedSize = upload.UncompressedSize
	uploadState.metadata = metadata

	return uploadState, 0, nil
}
