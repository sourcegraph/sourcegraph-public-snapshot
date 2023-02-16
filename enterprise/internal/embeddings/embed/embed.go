package embed

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5
const MAX_CODE_EMBEDDING_VECTORS = 768_000
const MAX_TEXT_EMBEDDING_VECTORS = 128_000

const EMBEDDING_BATCHES = 5
const EMBEDDING_BATCH_SIZE = 512

const EMBED_ENTIRE_FILE_TOKENS_THRESHOLD = 768
const EMBEDDING_CHUNK_TOKENS_THRESHOLD = 512
const EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD = EMBEDDING_CHUNK_TOKENS_THRESHOLD - 64

var SPLIT_OPTIONS = split.SplitOptions{
	ChunkTokensThreshold:           EMBEDDING_CHUNK_TOKENS_THRESHOLD,
	ChunkEarlySplitTokensThreshold: EMBEDDING_CHUNK_EARLY_SPLIT_TOKENS_THRESHOLD,
}

type readFile func(fileName string) ([]byte, error)

func EmbedRepo(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileNames []string, embeddingsConfig *schema.Embeddings, readFile readFile) (*embeddings.RepoEmbeddingIndex, error) {
	codeFileNames, textFileNames := []string{}, []string{}
	for _, fileName := range fileNames {
		if isTextFile(fileName) {
			textFileNames = append(textFileNames, fileName)
		} else if isValidCodeFile(fileName) {
			codeFileNames = append(codeFileNames, fileName)
		}
	}

	codeIndex, err := embedFiles(codeFileNames, readFile, embeddingsConfig, MAX_CODE_EMBEDDING_VECTORS)
	if err != nil {
		return nil, err
	}

	textIndex, err := embedFiles(textFileNames, readFile, embeddingsConfig, MAX_TEXT_EMBEDDING_VECTORS)
	if err != nil {
		return nil, err
	}

	return &embeddings.RepoEmbeddingIndex{
		RepoName:  repoName,
		Revision:  revision,
		CodeIndex: codeIndex,
		TextIndex: textIndex,
	}, nil
}

func embedFiles(fileNames []string, readFile readFile, config *schema.Embeddings, maxEmbeddingVectors int) (*embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata], error) {
	if len(fileNames) == 0 {
		return nil, nil
	}

	index := &embeddings.EmbeddingIndex[embeddings.RepoEmbeddingRowMetadata]{
		Embeddings:      make([]float32, 0, len(fileNames)*config.Dimensions),
		RowMetadata:     make([]embeddings.RepoEmbeddingRowMetadata, 0, len(fileNames)),
		ColumnDimension: config.Dimensions,
	}

	batchAddEmbeddings := func(embeddableChunks []split.EmbeddableChunk, batchSize int) error {
		for i := 0; i < len(embeddableChunks); i += batchSize {
			end := min(len(embeddableChunks), i+batchSize)
			batch := embeddableChunks[i:end]

			batchChunks := make([]string, len(batch))
			for idx, chunk := range batch {
				batchChunks[idx] = chunk.Content
				index.RowMetadata = append(index.RowMetadata, embeddings.RepoEmbeddingRowMetadata{
					FileName:  chunk.FileName,
					StartLine: chunk.StartLine,
					EndLine:   chunk.EndLine,
				})
			}

			batchEmbeddings, err := GetEmbeddingsWithRetries(batchChunks, config, GET_EMBEDDINGS_MAX_RETRIES)
			if err != nil {
				return errors.Wrap(err, "error while getting embeddings")
			}
			index.Embeddings = append(index.Embeddings, batchEmbeddings...)
		}
		return nil
	}

	embeddableChunks := []split.EmbeddableChunk{}
	for _, fileName := range fileNames {
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
		if !isEmbeddableFile(content) {
			continue
		}

		if embeddings.CountTokens(content) < EMBED_ENTIRE_FILE_TOKENS_THRESHOLD {
			embeddableChunks = append(embeddableChunks, split.EmbeddableChunk{FileName: fileName, StartLine: 0, EndLine: strings.Count(content, "\n") + 1, Content: content})
		} else {
			embeddableChunks = append(embeddableChunks, split.SplitIntoEmbeddableChunks(content, fileName, SPLIT_OPTIONS)...)
		}

		if len(embeddableChunks) > EMBEDDING_BATCHES*EMBEDDING_BATCH_SIZE {
			err := batchAddEmbeddings(embeddableChunks, EMBEDDING_BATCH_SIZE)
			if err != nil {
				return nil, err
			}
			embeddableChunks = []split.EmbeddableChunk{}
		}
	}

	if len(embeddableChunks) > 0 {
		err := batchAddEmbeddings(embeddableChunks, EMBEDDING_BATCH_SIZE)
		if err != nil {
			return nil, err
		}
	}

	return index, nil
}
