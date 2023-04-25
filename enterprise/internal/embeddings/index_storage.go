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

func UploadRepoEmbeddingIndex(ctx context.Context, uploadStore uploadstore.Store, key string, index *RepoEmbeddingIndex) error {
	pr, pw := io.Pipe()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer pw.Close()

		enc := gob.NewEncoder(pw)
		return encodeRepoEmbeddingIndex(enc, index, embeddingsChunkSize)
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
	defer file.Close()

	dec := gob.NewDecoder(file)

	rei, err := decodeRepoEmbeddingIndex(dec)
	// If decoding fails, assume it is an old index and decode with a generic decoder.
	if err != nil {
		originalErr := err
		rei, err = DownloadIndex[RepoEmbeddingIndex](ctx, uploadStore, key)
		if err != nil {
			// Return both errors in case the first one is the one we care about
			return nil, errors.Append(originalErr, err)
		}
	}

	return rei, nil
}

const embeddingsChunkSize = 10_000

// encodeRepoEmbeddingIndex is a specialized encoder for repo embedding indexes. Instead of GOB-encoding
// the entire RepoEmbeddingIndex, we encode each field separately, and we encode the embeddings array by chunks.
// This way we avoid allocating a separate very large slice for the embeddings.
func encodeRepoEmbeddingIndex(enc *gob.Encoder, rei *RepoEmbeddingIndex, chunkSize int) error {
	if err := enc.Encode(rei.RepoName); err != nil {
		return err
	}

	if err := enc.Encode(rei.Revision); err != nil {
		return err
	}

	for _, ei := range []EmbeddingIndex{rei.CodeIndex, rei.TextIndex} {
		if err := enc.Encode(ei.ColumnDimension); err != nil {
			return err
		}

		if err := enc.Encode(ei.RowMetadata); err != nil {
			return err
		}

		if err := enc.Encode(ei.Ranks); err != nil {
			return err
		}

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

			if err := enc.Encode(Dequantize(ei.Embeddings[start:end])); err != nil {
				return err
			}
		}
	}

	return nil
}

func decodeRepoEmbeddingIndex(dec *gob.Decoder) (*RepoEmbeddingIndex, error) {
	rei := &RepoEmbeddingIndex{}

	if err := dec.Decode(&rei.RepoName); err != nil {
		return nil, err
	}

	if err := dec.Decode(&rei.Revision); err != nil {
		return nil, err
	}

	for _, ei := range []*EmbeddingIndex{&rei.CodeIndex, &rei.TextIndex} {
		if err := dec.Decode(&ei.ColumnDimension); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.RowMetadata); err != nil {
			return nil, err
		}

		if err := dec.Decode(&ei.Ranks); err != nil {
			return nil, err
		}

		var numChunks int
		if err := dec.Decode(&numChunks); err != nil {
			return nil, err
		}

		ei.Embeddings = make([]int8, 0, numChunks*ei.ColumnDimension)
		for i := 0; i < numChunks; i++ {
			var embeddingSlice []float32
			if err := dec.Decode(&embeddingSlice); err != nil {
				return nil, err
			}
			ei.Embeddings = append(ei.Embeddings, Quantize(embeddingSlice)...)
		}
	}

	return rei, nil
}
