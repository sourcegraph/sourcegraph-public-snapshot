package uploadhandler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

func SingleUpload[T any](ctx context.Context,
	dbStore DBStore[T],
	uploadStore uploadstore.Store,
	metadata T, uncompressedSize *int64, body io.Reader) (_ int, err error) {
	// ctx, trace, _ := h.operations.handleEnqueueSinglePayload.With(ctx, &err, observation.Args{})
	// defer func() {
	// 	endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
	// 		attribute.Int("statusCode", statusCode),
	// 	}})
	// }()

	var uploadID int
	if err := dbStore.WithTransaction(ctx, func(tx DBStore[T]) error {
		id, err := tx.InsertUpload(ctx, Upload[T]{
			State:            "uploading",
			NumParts:         1,
			UploadedParts:    []int{0},
			UncompressedSize: uncompressedSize,
			Metadata:         metadata,
		})
		if err != nil {
			return err
		}
		// trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", id))

		size, err := uploadStore.Upload(ctx, fmt.Sprintf("upload-%d.lsif.gz", id), body)
		if err != nil {
			return err
		}
		// trace.AddEvent("TODO Domain Owner", attribute.Int("gzippedUploadSize", int(size)))

		if err := tx.MarkQueued(ctx, id, &size); err != nil {
			return err
		}

		uploadID = id
		return nil
	}); err != nil {
		return 0, err
	}

	// h.logger.Info(
	// 	"uploadhandler: enqueued upload",
	// 	sglog.Int("id", uploadID),
	// )
	//
	// older versions of src-cli expect a string
	return uploadID, nil
}

// handleEnqueueSinglePayload handles a non-multipart upload. This creates an upload record
// with state 'queued', proxies the data to the bundle manager, and returns the generated ID.
func (h *UploadHandler[T]) handleEnqueueSinglePayload(ctx context.Context, uploadState uploadState[T], body io.Reader) (_ any, statusCode int, err error) {
	ctx, _, endObservation := h.operations.handleEnqueueSinglePayload.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{
			attribute.Int("statusCode", statusCode),
		}})
	}()

	// var uploadID int
	// if err := h.dbStore.WithTransaction(ctx, func(tx DBStore[T]) error {
	// 	id, err := tx.InsertUpload(ctx, Upload[T]{
	// 		State:            "uploading",
	// 		NumParts:         1,
	// 		UploadedParts:    []int{0},
	// 		UncompressedSize: uploadState.uncompressedSize,
	// 		Metadata:         uploadState.metadata,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	trace.AddEvent("TODO Domain Owner", attribute.Int("uploadID", id))
	//
	// 	size, err := h.uploadStore.Upload(ctx, fmt.Sprintf("upload-%d.lsif.gz", id), body)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	trace.AddEvent("TODO Domain Owner", attribute.Int("gzippedUploadSize", int(size)))
	//
	// 	if err := tx.MarkQueued(ctx, id, &size); err != nil {
	// 		return err
	// 	}
	//
	// 	uploadID = id
	// 	return nil
	// }); err != nil {
	// 	return nil, http.StatusInternalServerError, err
	// }

	uploadID, err := SingleUpload(ctx, h.dbStore, h.uploadStore, uploadState.metadata, uploadState.uncompressedSize, body)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	h.logger.Info(
		"uploadhandler: enqueued upload",
		sglog.Int("id", uploadID),
	)

	// older versions of src-cli expect a string
	return struct {
		ID string `json:"id"`
	}{ID: strconv.Itoa(uploadID)}, 0, nil
}
