package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadHandler struct {
	db          dbutil.DB
	dbStore     DBStore
	uploadStore uploadstore.Store
	operations  *Operations
}

func NewUploadHandler(
	db dbutil.DB,
	dbStore DBStore,
	uploadStore uploadstore.Store,
	internal bool,
	authValidators AuthValidatorMap,
	operations *Operations,
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
	return authMiddleware(http.HandlerFunc(handler.handleEnqueue), db, authValidators)
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
	if payload, statusCode, err := func() (_ interface{}, statusCode int, err error) {
		ctx, endObservation := h.operations.handleEnqueue.With(r.Context(), &err, observation.Args{})
		defer func() {
			endObservation(1, observation.Args{LogFields: []log.Field{
				log.Int("statusCode", statusCode),
			}})
		}()

		uploadState, statusCode, err := h.constructUploadState(ctx, r)
		if err != nil {
			return nil, statusCode, err
		}

		if uploadHandlerFunc := h.selectUploadHandlerFunc(uploadState); uploadHandlerFunc != nil {
			return uploadHandlerFunc(ctx, uploadState, r.Body)
		}

		return nil, http.StatusBadRequest, errUnprocessableRequest
	}(); err != nil {
		log15.Error("Failed to enqueue payload", "error", err)
		http.Error(w, fmt.Sprintf("failed to enqueue payload: %s", err.Error()), statusCode)
		return
	} else if payload != nil {
		// 202 with identifier payload
		w.WriteHeader(http.StatusAccepted)
		writeJSON(w, payload)
	} else {
		// 204 with no body
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *UploadHandler) selectUploadHandlerFunc(uploadState uploadState) func(
	ctx context.Context,
	uploadState uploadState,
	body io.Reader,
) (
	payload interface{},
	statusCode int,
	err error,
) {
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

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "error", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "error", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}
