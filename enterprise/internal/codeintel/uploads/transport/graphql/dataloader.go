package graphql

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

func NewUploadLoaderFactory(uploadService UploadsService) *dataloader.DataloaderFactory[int, shared.Upload] {
	return dataloader.NewDataloaderFactory[int, shared.Upload](dataloader.BackingServiceFunc[int, shared.Upload](uploadService.GetUploadsByIDs))
}

func NewIndexLoaderFactory(uploadService UploadsService) *dataloader.DataloaderFactory[int, shared.Index] {
	return dataloader.NewDataloaderFactory[int, shared.Index](dataloader.BackingServiceFunc[int, shared.Index](uploadService.GetIndexesByIDs))
}
