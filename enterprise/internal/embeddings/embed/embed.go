package embed

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/binary"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5
const MAX_CODE_EMBEDDING_VECTORS = 768_000
const MAX_TEXT_EMBEDDING_VECTORS = 128_000

const EMBEDDING_BATCHES = 5
const EMBEDDING_BATCH_SIZE = 512

type filesReader func(func(fileName string, fileReader io.Reader) error) error

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	client EmbeddingsClient,
	splitOptions split.SplitOptions,
	readCodeFiles filesReader,
	readTextFiles filesReader,
	ranks types.RepoPathRanks,
) (*embeddings.RepoEmbeddingIndex, error) {
	codeIndex, err := embedFiles(readCodeFiles, client, splitOptions, MAX_CODE_EMBEDDING_VECTORS, ranks)
	if err != nil {
		return nil, err
	}

	textIndex, err := embedFiles(readTextFiles, client, splitOptions, MAX_TEXT_EMBEDDING_VECTORS, ranks)
	if err != nil {
		return nil, err
	}

	return &embeddings.RepoEmbeddingIndex{RepoName: repoName, Revision: revision, CodeIndex: codeIndex, TextIndex: textIndex}, nil
}

func createEmptyEmbeddingIndex(columnDimension int) embeddings.EmbeddingIndex {
	return embeddings.EmbeddingIndex{
		Embeddings:      []float32{},
		RowMetadata:     []embeddings.RepoEmbeddingRowMetadata{},
		ColumnDimension: columnDimension,
		Ranks:           []float32{},
	}
}

// embedFiles embeds file contents from the given file names. Since embedding models can only handle a certain amount of text (tokens) we cannot embed
// entire files. So we split the file contents into chunks and get embeddings for the chunks in batches. Functions returns an EmbeddingIndex containing
// the embeddings and metadata about the chunks the embeddings correspond to.
func embedFiles(
	fr filesReader,
	client EmbeddingsClient,
	splitOptions split.SplitOptions,
	maxEmbeddingVectors int,
	repoPathRanks types.RepoPathRanks,
) (embeddings.EmbeddingIndex, error) {
	dimensions, err := client.GetDimensions()
	if err != nil {
		return createEmptyEmbeddingIndex(dimensions), err
	}

	index := embeddings.EmbeddingIndex{
		Embeddings:      make([]float32, 0, dimensions),
		RowMetadata:     make([]embeddings.RepoEmbeddingRowMetadata, 0, 50),
		ColumnDimension: dimensions,
		Ranks:           make([]float32, 0, 50),
	}

	// addEmbeddableChunks batches embeddable chunks, gets embeddings for the batches, and appends them to the index above.
	addEmbeddableChunks := func(embeddableChunks []split.EmbeddableChunk, batchSize int) error {
		// The embeddings API operates with batches up to a certain size, so we can't send all embeddable chunks for embedding at once.
		// We batch them according to `batchSize`, and embed one by one.
		for i := 0; i < len(embeddableChunks); i += batchSize {
			end := min(len(embeddableChunks), i+batchSize)
			batch := embeddableChunks[i:end]
			batchChunks := make([]string, len(batch))
			for idx, chunk := range batch {
				batchChunks[idx] = chunk.Content
				index.RowMetadata = append(index.RowMetadata, embeddings.RepoEmbeddingRowMetadata{FileName: chunk.FileName, StartLine: chunk.StartLine, EndLine: chunk.EndLine})

				// Unknown documents have rank 0. Zoekt is a bit smarter about this, assigning 0
				// to "unimportant" files and the average for unknown files. We should probably
				// add this here, too.
				index.Ranks = append(index.Ranks, float32(repoPathRanks.Paths[chunk.FileName]))
			}

			batchEmbeddings, err := client.GetEmbeddingsWithRetries(batchChunks, GET_EMBEDDINGS_MAX_RETRIES)
			if err != nil {
				return errors.Wrap(err, "error while getting embeddings")
			}
			index.Embeddings = append(index.Embeddings, batchEmbeddings...)
		}
		return nil
	}

	embeddableChunks := []split.EmbeddableChunk{}
	err = fr(func(fileName string, fileReader io.Reader) error {
		// This is a fail-safe measure to prevent producing an extremely large index for large repositories.
		if len(index.RowMetadata) > maxEmbeddingVectors {
			return io.EOF
		}

		contentBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return errors.Wrap(err, "error while reading a file")
		}
		if binary.IsBinary(contentBytes) {
			return nil
		}
		content := string(contentBytes)
		if !IsEmbeddableFileContent(content) {
			return nil
		}

		embeddableChunks = append(embeddableChunks, split.SplitIntoEmbeddableChunks(content, fileName, splitOptions)...)

		if len(embeddableChunks) > EMBEDDING_BATCHES*EMBEDDING_BATCH_SIZE {
			err := addEmbeddableChunks(embeddableChunks, EMBEDDING_BATCH_SIZE)
			if err != nil {
				return err
			}
			embeddableChunks = []split.EmbeddableChunk{}
		}

		return nil
	})
	if err != nil && err != io.EOF {
		return createEmptyEmbeddingIndex(dimensions), err
	}

	if len(embeddableChunks) > 0 {
		err := addEmbeddableChunks(embeddableChunks, EMBEDDING_BATCH_SIZE)
		if err != nil {
			return createEmptyEmbeddingIndex(dimensions), err
		}
	}

	return index, nil
}
