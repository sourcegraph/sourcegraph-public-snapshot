package embeddings

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func DownloadIndex[T any](ctx context.Context, uploadStore uploadstore.Store, key string) (_ *T, err error) {
	file, err := uploadStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Append(err, file.Close()) }()

	var index T
	if err = gob.NewDecoder(file).Decode(&index); err != nil {
		return nil, err
	}
	return &index, nil
}

func UploadIndex[T any](ctx context.Context, uploadStore uploadstore.Store, key string, index T) error {
	buffer := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buffer).Encode(index); err != nil {
		return err
	}

	_, err := uploadStore.Upload(ctx, key, buffer)
	return err
}
