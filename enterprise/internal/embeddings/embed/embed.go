package embed

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	codeintelContext "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed/client/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed/client/sourcegraph"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewEmbeddingsClient(siteConfig *schema.SiteConfiguration) (client.EmbeddingsClient, error) {
	c := siteConfig.Embeddings
	if c == nil || !c.Enabled {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	switch c.Provider {
	case "sourcegraph", "":
		return sourcegraph.NewClient(siteConfig), nil
	case "openai":
		return openai.NewClient(c), nil
	default:
		return nil, errors.Newf("invalid provider %q", c.Provider)
	}
}

const (
	getEmbeddingsMaxRetries = 5
	embeddingsBatchSize     = 512
	maxFileSize             = 1_000_000 // 1MB
)

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	client client.EmbeddingsClient,
	contextService ContextService,
	readLister FileReadLister,
	ranks types.RepoPathRanks,
	opts EmbedRepoOpts,
	logger log.Logger,
) (*embeddings.RepoEmbeddingIndex, []string, *embeddings.EmbedRepoStats, error) {
	start := time.Now()

	var toIndex []FileEntry
	var toRemove []string
	var err error

	isIncremental := opts.IndexedRevision != ""

	if isIncremental {
		toIndex, toRemove, err = readLister.Diff(ctx, opts.IndexedRevision)
		if err != nil {
			logger.Error(
				"failed to get diff. Falling back to full index",
				log.String("RepoName", string(opts.RepoName)),
				log.String("revision", string(opts.Revision)),
				log.String("old revision", string(opts.IndexedRevision)),
				log.Error(err),
			)
			toRemove = nil
			isIncremental = false
		}
	}

	if !isIncremental { // full index
		toIndex, err = readLister.List(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	var codeFileNames, textFileNames []FileEntry
	for _, file := range toIndex {
		if IsValidTextFile(file.Name) {
			textFileNames = append(textFileNames, file)
		} else {
			codeFileNames = append(codeFileNames, file)
		}
	}

	codeIndex, codeIndexStats, err := embedFiles(ctx, codeFileNames, client, contextService, opts.ExcludePatterns, opts.SplitOptions, readLister, opts.MaxCodeEmbeddings, ranks)
	if err != nil {
		return nil, nil, nil, err
	}

	textIndex, textIndexStats, err := embedFiles(ctx, textFileNames, client, contextService, opts.ExcludePatterns, opts.SplitOptions, readLister, opts.MaxTextEmbeddings, ranks)
	if err != nil {
		return nil, nil, nil, err

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
		IsIncremental:  isIncremental,
	}

	return index, toRemove, stats, nil
}

type EmbedRepoOpts struct {
	RepoName          api.RepoName
	Revision          api.CommitID
	ExcludePatterns   []*paths.GlobPattern
	SplitOptions      codeintelContext.SplitOptions
	MaxCodeEmbeddings int
	MaxTextEmbeddings int

	// If set, we already have an index for a previous commit.
	IndexedRevision api.CommitID
}

// embedFiles embeds file contents from the given file names. Since embedding models can only handle a certain amount of text (tokens) we cannot embed
// entire files. So we split the file contents into chunks and get embeddings for the chunks in batches. Functions returns an EmbeddingIndex containing
// the embeddings and metadata about the chunks the embeddings correspond to.
func embedFiles(
	ctx context.Context,
	files []FileEntry,
	client client.EmbeddingsClient,
	contextService ContextService,
	excludePatterns []*paths.GlobPattern,
	splitOptions codeintelContext.SplitOptions,
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

	var batch []codeintelContext.EmbeddableChunk

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

		batchEmbeddings, err := client.GetEmbeddingsWithRetries(ctx, batchChunks, getEmbeddingsMaxRetries)
		if err != nil {
			return errors.Wrap(err, "error while getting embeddings")
		}
		index.Embeddings = append(index.Embeddings, embeddings.Quantize(batchEmbeddings)...)

		batch = batch[:0] // reset batch
		return nil
	}

	addToBatch := func(chunk codeintelContext.EmbeddableChunk) error {
		batch = append(batch, chunk)
		if len(batch) >= embeddingsBatchSize {
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
		if ctx.Err() != nil {
			return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, ctx.Err()
		}

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
		chunks, err := contextService.SplitIntoEmbeddableChunks(ctx, string(contentBytes), file.Name, splitOptions)
		if err != nil {
			return embeddings.EmbeddingIndex{}, embeddings.EmbedFilesStats{}, errors.Wrap(err, "error while splitting file")
		}
		for _, chunk := range chunks {
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
	FileDiffer
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

type FileDiffer interface {
	Diff(context.Context, api.CommitID) ([]FileEntry, []string, error)
}
