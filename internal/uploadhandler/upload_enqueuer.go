package uploadhandler

import (
	"context"
	"fmt"
	"io"

	sglog "github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type UploadEnqueuer[T any] struct {
	logger      sglog.Logger
	dbStore     DBStore[T]
	uploadStore object.Storage
	operations  EnqueuerOperations
}

func NewUploadEnqueuer[T any](observationCtx *observation.Context, dbStore DBStore[T], uploadStore object.Storage) UploadEnqueuer[T] {
	return UploadEnqueuer[T]{
		logger:      observationCtx.Logger.Scoped("upload_enqueuer"),
		dbStore:     dbStore,
		uploadStore: uploadStore,
		operations:  *NewEnqueuerOperations(observationCtx),
	}
}

type UploadResult struct {
	UploadID       int
	CompressedSize int64
}

func (u *UploadEnqueuer[T]) EnqueueSinglePayload(
	ctx context.Context,
	metadata T,
	uncompressedSize *int64,
	compressedBody io.Reader) (_ UploadResult, err error) {

	ctx, trace, endObservation := u.operations.enqueueSinglePayload.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{Attrs: []attribute.KeyValue{}})
	}()

	var uploadID int
	var uploadedSize int64
	if err := u.dbStore.WithTransaction(ctx, func(tx DBStore[T]) error {
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
		trace.AddEvent("insertUpload", attribute.Int("uploadID", id))

		uploadedSize, err = u.uploadStore.Upload(ctx, fmt.Sprintf("upload-%d.lsif.gz", id), compressedBody)
		if err != nil {
			return errors.Newf("Failed to upload data to upload store (id=%d): %s", id, err)
		}
		trace.AddEvent("uploadStore.Upload", attribute.Int64("gzippedUploadSize", uploadedSize))

		if err := tx.MarkQueued(ctx, id, &uploadedSize); err != nil {
			return errors.Newf("Failed to mark upload (id=%d) as queued: %s", id, err)
		}

		uploadID = id
		return nil
	}); err != nil {
		return UploadResult{}, err
	}

	trace.Info(
		"enqueueUpload",
		sglog.Int("id", uploadID),
	)

	return UploadResult{uploadID, uploadedSize}, nil
}
