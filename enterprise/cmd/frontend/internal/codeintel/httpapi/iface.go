package httpapi

import (
	"context"
)

type Upload struct {
	ID                int
	State             string
	NumParts          int
	UploadedParts     []int
	UploadSize        *int64
	UncompressedSize  *int64
	AssociatedIndexID *int
	Metadata          codeintelUploadMetadata
}

type DBStore interface {
	Transact(ctx context.Context) (DBStore, error)
	Done(err error) error

	GetUploadByID(ctx context.Context, uploadID int) (Upload, bool, error)
	InsertUpload(ctx context.Context, upload Upload) (int, error)
	AddUploadPart(ctx context.Context, uploadID, partIndex int) error
	MarkQueued(ctx context.Context, id int, uploadSize *int64) error
	MarkFailed(ctx context.Context, id int, reason string) error
}

// type DBStoreShim struct {
// 	*dbstore.Store
// }

// func (s *DBStoreShim) Transact(ctx context.Context) (DBStore, error) {
// 	tx, err := s.Store.Transact(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &DBStoreShim{tx}, nil
// }
