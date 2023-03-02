package embeddings

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func DownloadIndex[T any](ctx context.Context, uploadStore uploadstore.Store, key string) (_ *T, err error) {
	file, err := uploadStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Append(err, gzipReader.Close()) }()

	var index T
	if err = json.NewDecoder(gzipReader).Decode(&index); err != nil {
		return nil, err
	}
	return &index, nil
}

func UploadIndex[T any](ctx context.Context, uploadStore uploadstore.Store, key string, index T) error {
	buffer := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(buffer)
	if err := json.NewEncoder(gzipWriter).Encode(index); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	_, err := uploadStore.Upload(ctx, key, buffer)
	return err
}
