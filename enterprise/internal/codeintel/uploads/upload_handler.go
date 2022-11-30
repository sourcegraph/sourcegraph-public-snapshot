package uploads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/uploadhandler"
)

type UploadMetadata struct {
	RepositoryID      int
	Commit            string
	Root              string
	Indexer           string
	IndexerVersion    string
	AssociatedIndexID int
}

type uploadHandlerShim struct {
	store.Store
}

func (s *Service) UploadHandlerStore() uploadhandler.DBStore[UploadMetadata] {
	return &uploadHandlerShim{s.store}
}

func (s *uploadHandlerShim) Transact(ctx context.Context) (uploadhandler.DBStore[UploadMetadata], error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &uploadHandlerShim{tx}, nil
}

func (s *uploadHandlerShim) InsertUpload(ctx context.Context, upload uploadhandler.Upload[UploadMetadata]) (int, error) {
	var associatedIndexID *int
	if upload.Metadata.AssociatedIndexID != 0 {
		associatedIndexID = &upload.Metadata.AssociatedIndexID
	}

	return s.Store.InsertUpload(ctx, types.Upload{
		ID:                upload.ID,
		State:             upload.State,
		NumParts:          upload.NumParts,
		UploadedParts:     upload.UploadedParts,
		UploadSize:        upload.UploadSize,
		UncompressedSize:  upload.UncompressedSize,
		RepositoryID:      upload.Metadata.RepositoryID,
		Commit:            upload.Metadata.Commit,
		Root:              upload.Metadata.Root,
		Indexer:           upload.Metadata.Indexer,
		IndexerVersion:    upload.Metadata.IndexerVersion,
		AssociatedIndexID: associatedIndexID,
	})
}

func (s *uploadHandlerShim) GetUploadByID(ctx context.Context, uploadID int) (uploadhandler.Upload[UploadMetadata], bool, error) {
	upload, ok, err := s.Store.GetUploadByID(ctx, uploadID)
	if err != nil {
		return uploadhandler.Upload[UploadMetadata]{}, false, err
	}
	if !ok {
		return uploadhandler.Upload[UploadMetadata]{}, false, nil
	}

	u := uploadhandler.Upload[UploadMetadata]{
		ID:               upload.ID,
		State:            upload.State,
		NumParts:         upload.NumParts,
		UploadedParts:    upload.UploadedParts,
		UploadSize:       upload.UploadSize,
		UncompressedSize: upload.UncompressedSize,
		Metadata: UploadMetadata{
			RepositoryID:   upload.RepositoryID,
			Commit:         upload.Commit,
			Root:           upload.Root,
			Indexer:        upload.Indexer,
			IndexerVersion: upload.IndexerVersion,
		},
	}

	if upload.AssociatedIndexID != nil {
		u.Metadata.AssociatedIndexID = *upload.AssociatedIndexID
	}

	return u, true, nil
}
