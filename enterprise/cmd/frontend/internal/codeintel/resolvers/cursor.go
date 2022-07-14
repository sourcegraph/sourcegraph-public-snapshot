package resolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
)

// localCursor is an upload offset and a location offset within that upload.
type localCursor struct {
	UploadOffset int `json:"uploadOffset"`
	// The location offset within the associated upload.
	LocationOffset int `json:"locationOffset"`
}

// remoteCursor is an upload offset, the current batch of uploads, and a location offset within the batch of uploads.
type remoteCursor struct {
	UploadOffset   int   `json:"batchOffset"`
	UploadBatchIDs []int `json:"uploadBatchIDs"`
	// The location offset within the associated batch of uploads.
	LocationOffset int `json:"locationOffset"`
}

type cursorAdjustedUpload struct {
	DumpID               int                `json:"dumpID"`
	AdjustedPath         string             `json:"adjustedPath"`
	AdjustedPosition     lsifstore.Position `json:"adjustedPosition"`
	AdjustedPathInBundle string             `json:"adjustedPathInBundle"`
}

// adjustedUpload pairs an upload visible from the current target commit with the
// current target path and position adjusted so that it matches the data within the
// underlying index.
// type adjustedUpload struct {
// 	Upload               store.Dump
// 	AdjustedPath         string -> input path that is going t
// 	AdjustedPosition     lsifstore.Position
// 	AdjustedPathInBundle string
// }
/*
user supplied commit and path of the files



*/
