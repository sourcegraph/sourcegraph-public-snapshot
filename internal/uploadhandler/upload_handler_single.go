package uploadhandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// handleEnqueueSinglePayload handles a non-multipart upload. This creates an upload record
// with state 'queued', proxies the data to the bundle manager, and returns the generated ID.
func (h *UploadHandler[T]) handleEnqueueSinglePayload(ctx context.Context, uploadState uploadState[T], body io.Reader) (_ any, statusCode int, err error) {
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

	id, err := tx.InsertUpload(ctx, Upload[T]{
		State:            "uploading",
		NumParts:         1,
		UploadedParts:    []int{0},
		UncompressedSize: uploadState.uncompressedSize,
		Metadata:         uploadState.metadata,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", id))

	size, err := h.uploadStore.Upload(ctx, fmt.Sprintf("upload-%d.lsif.gz", id), body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("gzippedUploadSize", int(size)))

	if err := tx.MarkQueued(ctx, id, &size); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	h.logger.Info(
		"uploadhandler: enqueued upload",
		sglog.Int("id", id),
	)

	// older versions of src-cli expect a string
	return struct {
		ID string `json:"id"`
	}{ID: strconv.Itoa(id)}, 0, nil
}
