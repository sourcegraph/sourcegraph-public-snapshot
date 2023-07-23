package embed

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	bgrepo "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client/openai"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed/client/sourcegraph"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/paths"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewEmbeddingsClient(config *conftypes.EmbeddingsConfig) (client.EmbeddingsClient, error) {
	switch config.Provider {
	case "sourcegraph":
		return sourcegraph.NewClient(httpcli.ExternalClient, config), nil
	case "openai":
		return openai.NewClient(httpcli.ExternalClient, config), nil
	default:
		return nil, errors.Newf("invalid provider %q", config.Provider)
	}
}

const (
	getEmbeddingsMaxRetries = 5
	embeddingsBatchSize     = 512
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
	reportProgress func(*bgrepo.EmbedRepoStats),
) (*embeddings.RepoEmbeddingIndex, []string, *bgrepo.EmbedRepoStats, error) {
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

	dimensions, err := client.GetDimensions()
	if err != nil {
		return nil, nil, nil, err
	}
	newIndex := func(numFiles int) embeddings.EmbeddingIndex {
		return embeddings.EmbeddingIndex{
			Embeddings:      make([]int8, 0, numFiles*dimensions/2),
			RowMetadata:     make([]embeddings.RepoEmbeddingRowMetadata, 0, numFiles/2),
			ColumnDimension: dimensions,
			Ranks:           make([]float32, 0, numFiles/2),
		}
	}

	stats := bgrepo.EmbedRepoStats{
		CodeIndexStats: bgrepo.NewEmbedFilesStats(len(codeFileNames)),
		TextIndexStats: bgrepo.NewEmbedFilesStats(len(textFileNames)),
		IsIncremental:  isIncremental,
	}

	insertIndex := func(index *embeddings.EmbeddingIndex, metadata []embeddings.RepoEmbeddingRowMetadata, vectors []float32) {
		index.RowMetadata = append(index.RowMetadata, metadata...)
		index.Embeddings = append(index.Embeddings, embeddings.Quantize(vectors, nil)...)
		// Unknown documents have rank 0. Zoekt is a bit smarter about this, assigning 0
		// to "unimportant" files and the average for unknown files. We should probably
		// add this here, too.
		for _, md := range metadata {
			index.Ranks = append(index.Ranks, float32(ranks.Paths[md.FileName]))
		}
	}

	codeIndex := newIndex(len(codeFileNames))
	insertCode := func(md []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error {
		insertIndex(&codeIndex, md, embeddings)
		return nil
	}

	reportCodeProgress := func(codeIndexStats bgrepo.EmbedFilesStats) {
		stats.CodeIndexStats = codeIndexStats
		reportProgress(&stats)
	}

	codeIndexStats, err := embedFiles(ctx, codeFileNames, client, contextService, opts.FileFilters, opts.SplitOptions, readLister, opts.MaxCodeEmbeddings, insertCode, reportCodeProgress)
	if err != nil {
		return nil, nil, nil, err
	}
	stats.CodeIndexStats = codeIndexStats

	textIndex := newIndex(len(textFileNames))
	insertText := func(md []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error {
		insertIndex(&textIndex, md, embeddings)
		return nil
	}

	reportTextProgress := func(textIndexStats bgrepo.EmbedFilesStats) {
		stats.TextIndexStats = textIndexStats
		reportProgress(&stats)
	}

	textIndexStats, err := embedFiles(ctx, textFileNames, client, contextService, opts.FileFilters, opts.SplitOptions, readLister, opts.MaxTextEmbeddings, insertText, reportTextProgress)
	if err != nil {
		return nil, nil, nil, err
	}
	stats.TextIndexStats = textIndexStats

	embeddingsModel := client.GetModelIdentifier()
	index := &embeddings.RepoEmbeddingIndex{
		RepoName:        opts.RepoName,
		Revision:        opts.Revision,
		EmbeddingsModel: embeddingsModel,
		CodeIndex:       codeIndex,
		TextIndex:       textIndex,
	}

	return index, toRemove, &stats, nil
}

type EmbedRepoOpts struct {
	RepoName          api.RepoName
	Revision          api.CommitID
	FileFilters       FileFilters
	SplitOptions      codeintelContext.SplitOptions
	MaxCodeEmbeddings int
	MaxTextEmbeddings int

	// If set, we already have an index for a previous commit.
	IndexedRevision api.CommitID
}

type FileFilters struct {
	ExcludePatterns  []*paths.GlobPattern
	IncludePatterns  []*paths.GlobPattern
	MaxFileSizeBytes int
}

type batchInserter func(metadata []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error

// embedFiles embeds file contents from the given file names. Since embedding models can only handle a certain amount of text (tokens) we cannot embed
// entire files. So we split the file contents into chunks and get embeddings for the chunks in batches. Functions returns an EmbeddingIndex containing
// the embeddings and metadata about the chunks the embeddings correspond to.
func embedFiles(
	ctx context.Context,
	files []FileEntry,
	embeddingsClient client.EmbeddingsClient,
	contextService ContextService,
	fileFilters FileFilters,
	splitOptions codeintelContext.SplitOptions,
	reader FileReader,
	maxEmbeddingVectors int,
	insert batchInserter,
	reportProgress func(bgrepo.EmbedFilesStats),
) (bgrepo.EmbedFilesStats, error) {
	dimensions, err := embeddingsClient.GetDimensions()
	if err != nil {
		return bgrepo.EmbedFilesStats{}, err
	}

	stats := bgrepo.NewEmbedFilesStats(len(files))

	var batch []codeintelContext.EmbeddableChunk

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		batchChunks := make([]string, len(batch))
		metadata := make([]embeddings.RepoEmbeddingRowMetadata, len(batch))
		for idx, chunk := range batch {
			batchChunks[idx] = chunk.Content
			metadata[idx] = embeddings.RepoEmbeddingRowMetadata{
				FileName:  chunk.FileName,
				StartLine: chunk.StartLine,
				EndLine:   chunk.EndLine,
			}
		}

		batchEmbeddings, err := embeddingsClient.GetEmbeddings(ctx, batchChunks)
		if err != nil {
			if partErr := (client.PartialError{}); errors.As(err, &partErr) {
				return errors.Wrapf(partErr.Err, "batch failed on file %q", batch[partErr.Index].FileName)
			}
			return errors.Wrap(err, "error while getting embeddings")
		}
		if expected := len(batchChunks) * dimensions; len(batchEmbeddings) != expected {
			return errors.Newf("expected embeddings for batch to have length %d, got %d", expected, len(batchEmbeddings))
		}

		if err := insert(metadata, batchEmbeddings); err != nil {
			return err
		}

		batch = batch[:0] // reset batch
		reportProgress(stats)
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

	for _, file := range files {
		if ctx.Err() != nil {
			return bgrepo.EmbedFilesStats{}, ctx.Err()
		}

		// This is a fail-safe measure to prevent producing an extremely large index for large repositories.
		if stats.ChunksEmbedded >= maxEmbeddingVectors {
			stats.Skip(SkipReasonMaxEmbeddings, int(file.Size))
			continue
		}

		if file.Size > int64(fileFilters.MaxFileSizeBytes) {
			stats.Skip(SkipReasonLarge, int(file.Size))
			continue
		}

		if isExcludedFilePathMatch(file.Name, fileFilters.ExcludePatterns) {
			stats.Skip(SkipReasonExcluded, int(file.Size))
			continue
		}

		if !isIncludedFilePathMatch(file.Name, fileFilters.IncludePatterns) {
			stats.Skip(SkipReasonNotIncluded, int(file.Size))
			continue
		}

		contentBytes, err := reader.Read(ctx, file.Name)
		if err != nil {
			return bgrepo.EmbedFilesStats{}, errors.Wrap(err, "error while reading a file")
		}

		if embeddable, skipReason := isEmbeddableFileContent(contentBytes); !embeddable {
			stats.Skip(skipReason, len(contentBytes))
			continue
		}

		// At this point, we have determined that we want to embed this file.
		chunks, err := contextService.SplitIntoEmbeddableChunks(ctx, string(contentBytes), file.Name, splitOptions)
		if err != nil {
			return bgrepo.EmbedFilesStats{}, errors.Wrap(err, "error while splitting file")
		}
		for _, chunk := range chunks {
			if err := addToBatch(chunk); err != nil {
				return bgrepo.EmbedFilesStats{}, err
			}
			stats.AddChunk(len(chunk.Content))
		}
		stats.AddFile()
	}

	// Always do a final flush
	if err := flush(); err != nil {
		return bgrepo.EmbedFilesStats{}, err
	}

	return stats, nil
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
