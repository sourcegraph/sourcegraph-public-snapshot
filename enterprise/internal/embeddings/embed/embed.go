package embed

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5

const EMBEDDING_BATCHES = 5
const EMBEDDING_BATCH_SIZE = 512

const maxFileSize = 1000000 // 1MB

type ranksGetter func(ctx context.Context, repoName string) (types.RepoPathRanks, error)

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	client EmbeddingsClient,
	readLister FileReadLister,
	getDocumentRanks ranksGetter,
	opts EmbedRepoOpts,
) (*embeddings.RepoEmbeddingIndex, *embeddings.EmbedRepoStats, error) {
	start := time.Now()

	allFiles, err := readLister.List(ctx)
	if err != nil {
		return nil, nil, err
	}

	var codeFileNames, textFileNames []FileEntry
	for _, file := range allFiles {
		if isValidTextFile(file.Name) {
			textFileNames = append(textFileNames, file)
		} else {
			codeFileNames = append(codeFileNames, file)
		}
	}

	ranks, err := getDocumentRanks(ctx, string(opts.RepoName))
	if err != nil {
		return nil, nil, err
	}

	codeIndex, codeIndexStats, err := embedFiles(ctx, codeFileNames, client, opts.ExcludePatterns, opts.SplitOptions, readLister, opts.MaxCodeEmbeddings, ranks)
	if err != nil {
		return nil, nil, err
	}

	textIndex, textIndexStats, err := embedFiles(ctx, textFileNames, client, opts.ExcludePatterns, opts.SplitOptions, readLister, opts.MaxTextEmbeddings, ranks)
	if err != nil {
		return nil, nil, err

	}

	index := &embeddings.RepoEmbeddingIndex{
		RepoName:  opts.RepoName,
		Revision:  opts.Revision,
		CodeIndex: codeIndex,
		TextIndex: textIndex,
	}

	stats := &embeddings.EmbedRepoStats{
		Duration:       time.Since(start),
		HasRanks:       len(ranks.Paths) > 0,
		CodeIndexStats: codeIndexStats,
		TextIndexStats: textIndexStats,
	}

	return index, stats, nil
}

type EmbedRepoOpts struct {
	RepoName          api.RepoName
	Revision          api.CommitID
	ExcludePatterns   []*paths.GlobPattern
	SplitOptions      split.SplitOptions
	MaxCodeEmbeddings int
	MaxTextEmbeddings int
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
) (embeddings.EmbeddingIndex, embeddings.EmbedFilesStats, error) {
	start := time.Now()

	dimensions, err := client.GetDimensions()
	if err != nil {
		return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, err
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

	var (
		statsEmbeddedByteCount  int
		statsEmbeddedFileCount  int
		statsEmbeddedChunkCount int
		statsSkipped            SkipStats
	)
	for _, file := range files {
		// This is a fail-safe measure to prevent producing an extremely large index for large repositories.
		if statsEmbeddedChunkCount >= maxEmbeddingVectors {
			statsSkipped.Add(SkipReasonMaxEmbeddings, int(file.Size))
			continue
		}

		if file.Size > maxFileSize {
			statsSkipped.Add(SkipReasonLarge, int(file.Size))
			continue
		}

		if isExcludedFilePath(file.Name, excludePatterns) {
			statsSkipped.Add(SkipReasonExcluded, int(file.Size))
			continue
		}

		contentBytes, err := reader.Read(ctx, file.Name)
		if err != nil {
			return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, errors.Wrap(err, "error while reading a file")
		}

		if embeddable, skipReason := isEmbeddableFileContent(contentBytes); !embeddable {
			statsSkipped.Add(skipReason, len(contentBytes))
			continue
		}

		// At this point, we have determined that we want to embed this file.

		for _, chunk := range split.SplitIntoEmbeddableChunks(string(contentBytes), file.Name, splitOptions) {
			if err := addToBatch(chunk); err != nil {
				return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, err
			}
			statsEmbeddedByteCount += len(chunk.Content)
			statsEmbeddedChunkCount += 1
		}
		statsEmbeddedFileCount += 1
	}

	// Always do a final flush
	if err := flush(); err != nil {
		return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, err
	}

	stats := embeddings.EmbedFilesStats{
		Duration:           time.Since(start),
		EmbeddedBytes:      statsEmbeddedByteCount,
		EmbeddedFileCount:  statsEmbeddedFileCount,
		EmbeddedChunkCount: statsEmbeddedChunkCount,
		SkippedCounts:      statsSkipped.Counts(),
		SkippedByteCounts:  statsSkipped.ByteCounts(),
	}

	return index, stats, nil
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
