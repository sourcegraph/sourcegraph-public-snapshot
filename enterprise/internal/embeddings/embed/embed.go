package embed

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5
const MAX_CODE_EMBEDDING_VECTORS = 768_000
const MAX_TEXT_EMBEDDING_VECTORS = 128_000

const EMBEDDING_BATCHES = 5
const EMBEDDING_BATCH_SIZE = 512

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const EMBED_ENTIRE_FILE_TOKENS_THRESHOLD = 384
const EMBEDDING_CHUNK_TOKENS_THRESHOLD = 256
const EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD = EMBEDDING_CHUNK_TOKENS_THRESHOLD - 32

var SPLIT_OPTIONS = split.SplitOptions{
	ChunkTokensThreshold:           EMBEDDING_CHUNK_TOKENS_THRESHOLD,
	ChunkEarlySplitTokensThreshold: EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD,
}

type readFile func(fileName string) ([]byte, error)

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	fileNames []string,
	client EmbeddingsClient,
	readFile readFile,
) (*embeddings.RepoEmbeddingIndex, error) {
	codeFileNames, textFileNames := []string{}, []string{}
	for _, fileName := range fileNames {
		if isValidTextFile(fileName) {
			textFileNames = append(textFileNames, fileName)
		} else if isValidCodeFile(fileName) {
			codeFileNames = append(codeFileNames, fileName)
		}
	}

	codeIndex, err := embedFiles(codeFileNames, client, readFile, MAX_CODE_EMBEDDING_VECTORS)
	if err != nil {
		return nil, err
	}

	textIndex, err := embedFiles(textFileNames, client, readFile, MAX_TEXT_EMBEDDING_VECTORS)
	if err != nil {
		return nil, err
	}

	return &embeddings.RepoEmbeddingIndex{RepoName: repoName, Revision: revision, CodeIndex: codeIndex, TextIndex: textIndex}, nil
}

// embedFiles embeds file contents from the given file names. Since embedding models can only handle a certain amount of text (tokens) we cannot embed
// entire files. So we split the file contents into chunks and get embeddings for the chunks in batches. Functions returns an EmbeddingIndex containing
// the embeddings and metadata about the chunks the embeddings correspond to.
func embedFiles(
	fileNames []string,
	client EmbeddingsClient,
	readFile readFile,
	maxEmbeddingVectors int,
) (*embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata], error) {
	if len(fileNames) == 0 {
		return nil, nil
	}

	dimensions, err := client.GetDimensions()
	if err != nil {
		return nil, err
	}
	index := &embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata]{
		Embeddings:      make([]float32, 0, len(fileNames)*dimensions),
		RowMetadata:     make([]embeddings.RepoEmbeddingRowMetadata, 0, len(fileNames)),
		ColumnDimension: dimensions,
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
	for _, fileName := range fileNames {
		// This is a fail-safe measure to prevent producing an extremely large index for large repositories.
		if len(index.RowMetadata) > maxEmbeddingVectors {
			break
		}

		contentBytes, err := readFile(fileName)
		if err != nil {
			return nil, errors.Wrap(err, "error while reading a file")
		}
		if !utf8.Valid(contentBytes) {
			continue
		}
		content := string(contentBytes)
		if !isEmbeddableFile(fileName, content) {
			continue
		}

		// If the file is small enough, embed the entire file rather than splitting it into chunks.
		if embeddings.EstimateTokens(content) < EMBED_ENTIRE_FILE_TOKENS_THRESHOLD {
			embeddableChunks = append(embeddableChunks, split.EmbeddableChunk{FileName: fileName, StartLine: 0, EndLine: strings.Count(content, "\n") + 1, Content: content})
		} else {
			embeddableChunks = append(embeddableChunks, split.SplitIntoEmbeddableChunks(content, fileName, SPLIT_OPTIONS)...)
		}

		if len(embeddableChunks) > EMBEDDING_BATCHES*EMBEDDING_BATCH_SIZE {
			err := addEmbeddableChunks(embeddableChunks, EMBEDDING_BATCH_SIZE)
			if err != nil {
				return nil, err
			}
			embeddableChunks = []split.EmbeddableChunk{}
		}
	}

	if len(embeddableChunks) > 0 {
		err := addEmbeddableChunks(embeddableChunks, EMBEDDING_BATCH_SIZE)
		if err != nil {
			return nil, err
		}
	}

	return index, nil
}
