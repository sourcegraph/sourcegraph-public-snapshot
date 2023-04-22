package embed

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5
const MAX_CODE_EMBEDDING_VECTORS = 768_000
const MAX_TEXT_EMBEDDING_VECTORS = 128_000

const EMBEDDING_BATCHES = 5
const EMBEDDING_BATCH_SIZE = 512

const maxFileSize = 1000000 // 1MB

type ranksGetter func(ctx context.Context, repoName string) (types.RepoPathRanks, error)

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	excludePatterns []*paths.GlobPattern,
	client EmbeddingsClient,
	splitOptions split.SplitOptions,
	readLister FileReadLister,
	getDocumentRanks ranksGetter,
) (*embeddings.RepoEmbeddingIndex, error) {
	allFiles, err := readLister.List(ctx)
	if err != nil {
		return nil, err
	}

	var codeFileNames, textFileNames []FileEntry
	for _, file := range allFiles {
		if isValidTextFile(file.Name) {
			textFileNames = append(textFileNames, file)
		} else {
			codeFileNames = append(codeFileNames, file)
		}
	}

	ranks, err := getDocumentRanks(ctx, string(repoName))
	if err != nil {
		return nil, err
	}

	codeIndex, err := embedFiles(ctx, codeFileNames, client, excludePatterns, splitOptions, readLister, MAX_CODE_EMBEDDING_VECTORS, ranks)
	if err != nil {
		return nil, err
	}

	textIndex, err := embedFiles(ctx, textFileNames, client, excludePatterns, splitOptions, readLister, MAX_TEXT_EMBEDDING_VECTORS, ranks)
	if err != nil {
		return nil, err
	}

	return &embeddings.RepoEmbeddingIndex{RepoName: repoName, Revision: revision, CodeIndex: codeIndex, TextIndex: textIndex}, nil
}

// embedFiles embeds file contents from the given file names. Since embedding models can only handle a certain amount of text (tokens) we cannot embed
// entire files. So we split the file contents into chunks and get embeddings for the chunks in batches. Functions returns an EmbeddingIndex containing
// the embeddings and metadata about the chunks the embeddings correspond to.
func embedFiles(
	ctx context.Context,
	files []FileEntry,
	client EmbeddingsClient,
	excludePatterns []*paths.GlobPattern,
	splitOptions split.SplitOptions,
	reader FileReader,
	maxEmbeddingVectors int,
	repoPathRanks types.RepoPathRanks,
) (embeddings.EmbeddingIndex, error) {
	dimensions, err := client.GetDimensions()
	if err != nil {
		return embeddings.EmbeddingIndex{}, err
	}

	index := embeddings.EmbeddingIndex{
		Embeddings:      make([]int8, 0, len(files)*dimensions/2),
		RowMetadata:     make([]embeddings.RepoEmbeddingRowMetadata, 0, len(files)/2),
		ColumnDimension: dimensions,
		Ranks:           make([]float32, 0, len(files)/2),
	}

	var batch []split.EmbeddableChunk

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		batchChunks := make([]string, len(batch))
		for idx, chunk := range batch {
			batchChunks[idx] = chunk.Content
			index.RowMetadata = append(index.RowMetadata, embeddings.RepoEmbeddingRowMetadata{FileName: chunk.FileName, StartLine: chunk.StartLine, EndLine: chunk.EndLine})

			// Unknown documents have rank 0. Zoekt is a bit smarter about this, assigning 0
			// to "unimportant" files and the average for unknown files. We should probably
			// add this here, too.
			index.Ranks = append(index.Ranks, float32(repoPathRanks.Paths[chunk.FileName]))
		}

		batchEmbeddings, err := client.GetEmbeddingsWithRetries(ctx, batchChunks, GET_EMBEDDINGS_MAX_RETRIES)
		if err != nil {
			return errors.Wrap(err, "error while getting embeddings")
		}
		index.Embeddings = append(index.Embeddings, embeddings.Quantize(batchEmbeddings)...)
		batch = batch[:0] // reset batch
		return nil
	}

	addToBatch := func(chunk split.EmbeddableChunk) error {
		batch = append(batch, chunk)
		if len(batch) >= EMBEDDING_BATCH_SIZE {
			// Flush if we've hit batch size
			return flush()
		}
		return nil
	}

	for _, file := range files {
		// This is a fail-safe measure to prevent producing an extremely large index for large repositories.
		if len(index.RowMetadata) > maxEmbeddingVectors {
			break
		}

		if file.Size > maxFileSize {
			continue
		}

		if isExcludedFilePath(file.Name, excludePatterns) {
			continue
		}

		contentBytes, err := reader.Read(ctx, file.Name)
		if err != nil {
			return embeddings.EmbeddingIndex{}, errors.Wrap(err, "error while reading a file")
		}

		if embeddable, _ := isEmbeddableFileContent(contentBytes); !embeddable {
			continue
		}

		for _, chunk := range split.SplitIntoEmbeddableChunks(string(contentBytes), file.Name, splitOptions) {
			if err := addToBatch(chunk); err != nil {
				return embeddings.EmbeddingIndex{}, err
			}
		}
	}

	// Always do a final flush
	if err := flush(); err != nil {
		return embeddings.EmbeddingIndex{}, err
	}

	return index, nil
}

type FileReadLister interface {
	FileReader
	FileLister
}

type FileEntry struct {
	Name string
	Size int64
}

type FileLister interface {
	List(context.Context) ([]FileEntry, error)
}

type FileReader interface {
	Read(context.Context, string) ([]byte, error)
}
