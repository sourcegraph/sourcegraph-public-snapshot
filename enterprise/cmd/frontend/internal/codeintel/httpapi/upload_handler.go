package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type UploadHandler struct {
	db          database.DB
	dbStore     DBStore
	uploadStore uploadstore.Store
	operations  *Operations
}

func NewUploadHandler(
	db database.DB,
	dbStore DBStore,
	uploadStore uploadstore.Store,
	internal bool,
	authValidators AuthValidatorMap,
	operations *Operations,
	hub *sentry.Hub,
) http.Handler {
	handler := &UploadHandler{
		db:          db,
		dbStore:     dbStore,
		uploadStore: uploadStore,
		operations:  operations,
	}

	if internal {
		return http.HandlerFunc(handler.handleEnqueue)
	}

	// 🚨 SECURITY: Non-internal installations of this handler will require a user/repo
	// visibility check with the remote code host (if enabled via site configuration).
	return authMiddleware(http.HandlerFunc(handler.handleEnqueue), db, authValidators, operations.authMiddleware)
}

var errUnprocessableRequest = errors.New("unprocessable request: missing expected query arguments (uploadId, index, or done)")

// POST /upload
//
// handleEnqueue dispatches to the correct handler function based on the request's query args. Running
// the `src lsif upload` command will cause one of two sequences of requests to occur. For uploads that
// are small enough repos (that can be uploaded in one-shot), only one request will be made:
//
//    - POST `/upload?repositoryId,commit,root,indexerName`
//
// For larger uploads, the requests are broken up into a setup request, a serires of upload requests,
// and a finalization request:
//
//   - POST `/upload?repositoryId,commit,root,indexerName,multiPart=true,numParts={n}`
//   - POST `/upload?uploadId={id},index={i}`
//   - POST `/upload?uploadId={id},done=true`
//
// See the functions the following functions for details on how each request is handled:
//
//   - handleEnqueueSinglePayload
//   - handleEnqueueMultipartSetup'
//   - handleEnqueueMultipartUpload
//   - handleEnqueueMultipartFinalize
func (h *UploadHandler) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	// Wrap the interesting bits of this in a function literal that's immediately
	// executed so that we can instrument the duration and the resulting error more
	// easily. The remainder of the function simply serializes the result to the
	// HTTP response writer.
	payload, statusCode, err := func() (_ any, statusCode int, err error) {
		ctx, trace, endObservation := h.operations.handleEnqueue.With(r.Context(), &err, observation.Args{})
		defer func() {
			endObservation(1, observation.Args{LogFields: []log.Field{
				log.Int("statusCode", statusCode),
			}})
		}()

		uploadState, statusCode, err := h.constructUploadState(ctx, r)
		if err != nil {
			return nil, statusCode, err
		}
		trace.Log(
			log.Int("repositoryID", uploadState.repositoryID),
			log.Int("uploadID", uploadState.uploadID),
			log.String("commit", uploadState.commit),
			log.String("root", uploadState.root),
			log.String("indexer", uploadState.indexer),
			log.String("indexerVersion", uploadState.indexerVersion),
			log.Int("associatedIndexID", uploadState.associatedIndexID),
			log.Int("numParts", uploadState.numParts),
			log.Int("numUploadedParts", len(uploadState.uploadedParts)),
			log.Bool("multipart", uploadState.multipart),
			log.Bool("suppliedIndex", uploadState.suppliedIndex),
			log.Int("index", uploadState.index),
			log.Bool("done", uploadState.done),
		)

		if uploadHandlerFunc := h.selectUploadHandlerFunc(uploadState); uploadHandlerFunc != nil {
			return uploadHandlerFunc(ctx, uploadState, r.Body)
		}

		return nil, http.StatusBadRequest, errUnprocessableRequest
	}()
	if err != nil {
		if statusCode >= 500 {
			log15.Error("codeintel.httpapi: failed to enqueue payload", "error", err)
		}

		http.Error(w, fmt.Sprintf("failed to enqueue payload: %s", err.Error()), statusCode)
		return
	}

	if payload == nil {
		// 204 with no body
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("codeintel.httpapi: failed to serialize result", "error", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// 202 with identifier payload
	w.WriteHeader(http.StatusAccepted)

	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		log15.Error("codeintel.httpapi: failed to write payload to client", "error", err)
	}
}

type uploadHandlerFunc = func(context.Context, uploadState, io.Reader) (any, int, error)

func (h *UploadHandler) selectUploadHandlerFunc(uploadState uploadState) uploadHandlerFunc {
	if uploadState.uploadID == 0 {
		if uploadState.multipart {
			return h.handleEnqueueMultipartSetup
		}

		return h.handleEnqueueSinglePayload
	}

	if uploadState.suppliedIndex {
		return h.handleEnqueueMultipartUpload
	}

	if uploadState.done {
		return h.handleEnqueueMultipartFinalize
	}

	return nil
}
