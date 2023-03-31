package embeddings

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"

	"golang.org/x/sync/errgroup"

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

func encodeRepoEmbeddingIndex(pw io.Writer, rei *RepoEmbeddingIndex, chunkSize int) error {
	enc := gob.NewEncoder(pw)

	// Write RepoName field
	if err := enc.Encode(rei.RepoName); err != nil {
		return err
	}

	// Write Revision field
	if err := enc.Encode(rei.Revision); err != nil {
		return err
	}

	// Encode CodeIndex and TextIndex
	for _, ei := range []EmbeddingIndex{rei.CodeIndex, rei.TextIndex} {
		// Write Embeddings field
		numChunks := (len(ei.Embeddings) + chunkSize - 1) / chunkSize
		if err := enc.Encode(numChunks); err != nil {
			return err
		}
		for i := 0; i < numChunks; i++ {
			start := i * chunkSize
			end := start + chunkSize

			if end > len(ei.Embeddings) {
				end = len(ei.Embeddings)
			}

			if err := enc.Encode(ei.Embeddings[start:end]); err != nil {
				return err
			}
		}

		if err := enc.Encode(ei.ColumnDimension); err != nil {
			return err
		}

		if err := enc.Encode(ei.RowMetadata); err != nil {
			return err
		}

		if err := enc.Encode(ei.Ranks); err != nil {
			return err
		}
	}

	return nil
}

func UploadRepoEmbeddingIndex(ctx context.Context, uploadStore uploadstore.Store, key string, index *RepoEmbeddingIndex) error {
	pr, pw := io.Pipe()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer pw.Close()

		return encodeRepoEmbeddingIndex(pw, index, 1000)
	})

	eg.Go(func() error {
		defer pr.Close()

		_, err := uploadStore.Upload(ctx, key, pr)
		return err
	})

	return eg.Wait()
}

func DownloadRepoEmbeddingIndex(ctx context.Context, uploadStore uploadstore.Store, key string) (*RepoEmbeddingIndex, error) {
	file, err := uploadStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Append(err, file.Close()) }()

	rei := &RepoEmbeddingIndex{}
	dec := gob.NewDecoder(file)

	if err := dec.Decode(&rei.RepoName); err != nil {
		return nil, err
	}

	if err := dec.Decode(&rei.Revision); err != nil {
		return nil, err
	}

	// Decode CodeIndex and TextIndex
	for _, ei := range []*EmbeddingIndex{&rei.CodeIndex, &rei.TextIndex} {
		// Write Embeddings field
		var numChunks int
		if err := dec.Decode(&numChunks); err != nil {
			return nil, err
		}

		for i := 0; i < numChunks; i++ {
			var embeddingSlice []float32
			if err := dec.Decode(&embeddingSlice); err != nil {
				return nil, err
			}
			ei.Embeddings = append(ei.Embeddings, embeddingSlice...)
		}

		if err := dec.Decode(&ei.ColumnDimension); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.RowMetadata); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.Ranks); err != nil {
			return nil, err
		}
	}

	return rei, nil
}
