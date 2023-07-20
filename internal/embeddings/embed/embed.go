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

	truncateIndex := func(index *embeddings.EmbeddingIndex, count int) {
		cutoff := len(index.RowMetadata) - count

		index.RowMetadata = index.RowMetadata[:cutoff]
		index.Embeddings = index.Embeddings[:cutoff*index.ColumnDimension]
		index.Ranks = index.Ranks[:cutoff]
	}

	codeIndex := newIndex(len(codeFileNames))
	insertCode := func(md []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error {
		insertIndex(&codeIndex, md, embeddings)
		return nil
	}

	truncateCode := func(count int) error {
		truncateIndex(&codeIndex, count)
		return nil
	}

	reportCodeProgress := func(codeIndexStats bgrepo.EmbedFilesStats) {
		stats.CodeIndexStats = codeIndexStats
		reportProgress(&stats)
	}

	codeIndexStats, err := embedFiles(ctx, codeFileNames, client, contextService, opts.FileFilters, opts.SplitOptions, readLister, opts.MaxCodeEmbeddings, opts.BatchSize, insertCode, truncateCode, reportCodeProgress, logger)
	if err != nil {
		return nil, nil, nil, err
	}
	stats.CodeIndexStats = codeIndexStats

	textIndex := newIndex(len(textFileNames))
	insertText := func(md []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error {
		insertIndex(&textIndex, md, embeddings)
		return nil
	}

	truncateText := func(count int) error {
		truncateIndex(&textIndex, count)
		return nil
	}

	reportTextProgress := func(textIndexStats bgrepo.EmbedFilesStats) {
		stats.TextIndexStats = textIndexStats
		reportProgress(&stats)
	}

	textIndexStats, err := embedFiles(ctx, textFileNames, client, contextService, opts.FileFilters, opts.SplitOptions, readLister, opts.MaxTextEmbeddings, opts.BatchSize, insertText, truncateText, reportTextProgress, logger)
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
	RepoName           api.RepoName
	Revision           api.CommitID
	FileFilters        FileFilters
	SplitOptions       codeintelContext.SplitOptions
	MaxCodeEmbeddings  int
	MaxTextEmbeddings  int
	BatchSize          int
	ExcludeFileOnError bool

	// If set, we already have an index for a previous commit.
	IndexedRevision api.CommitID
}

type FileFilters struct {
	ExcludePatterns  []*paths.GlobPattern
	IncludePatterns  []*paths.GlobPattern
	MaxFileSizeBytes int
}

type batchInserter func(metadata []embeddings.RepoEmbeddingRowMetadata, embeddings []float32) error
type truncate func(count int) error

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
	batchSize int,
	insert batchInserter,
	truncate truncate,
	reportProgress func(bgrepo.EmbedFilesStats),
	logger log.Logger,
) (bgrepo.EmbedFilesStats, error) {
	dimensions, err := embeddingsClient.GetDimensions()
	if err != nil {
		return bgrepo.EmbedFilesStats{}, err
	}

	stats := bgrepo.NewEmbedFilesStats(len(files))

	var batch []codeintelContext.EmbeddableChunk

	// how many chunks were added to batches for the current file
	var currentFileBatchedCount int

	// how many embedding rows were indexed for the most recently indexed file
	var partiallyIndexedFileCount int

	flush := func() (bool, error) {
		if len(batch) == 0 {
			return false, nil
		}

		batchChunks := make([]string, len(batch))
		for idx, chunk := range batch {
			batchChunks[idx] = chunk.Content
		}

		batchEmbeddings, err := embeddingsClient.GetEmbeddings(ctx, batchChunks)
		if err != nil {
			if partErr := (client.PartialError{}); errors.As(err, &partErr) {
				return false, errors.Wrapf(partErr.Err, "batch failed on file %q", batch[partErr.Index].FileName)
			}
			return false, errors.Wrap(err, "error while getting embeddings")
		}

		if expected := len(batchChunks) * dimensions; len(batchEmbeddings.Embeddings) != expected {
			return false, errors.Newf("expected embeddings for batch to have length %d, got %d", expected, len(batchEmbeddings.Embeddings))
		}

		// files with partial failures need to be excluded from the index entirely
		excludedFileNames := make(map[string]struct{}, len(batchEmbeddings.Failed))
		for _, batchIdx := range batchEmbeddings.Failed {
			if batchIdx < 0 || batchIdx >= len(batch) {
				continue
			}
			if _, ok := excludedFileNames[batch[batchIdx].FileName]; !ok {
				excludedFileNames[batch[batchIdx].FileName] = struct{}{}
			}
		}

		// if the first file of this batch had a partial failure and already has had chunks
		// indexed by the immediately preceding flush operation(s) then truncate the index to exclude that file
		if _, ok := excludedFileNames[batch[0].FileName]; ok {
			if err := truncate(partiallyIndexedFileCount); err != nil {
				return false, err
			}
		}

		for fileName, _ := range excludedFileNames {
			logger.Debug("error getting embeddings for file",
				log.String("file", fileName),
			)
		}

		metadata := make([]embeddings.RepoEmbeddingRowMetadata, 0, (len(batchEmbeddings.Embeddings)-len(batchEmbeddings.Failed))*dimensions)

		cursor := 0
		for idx, chunk := range batch {
			if _, ok := excludedFileNames[chunk.FileName]; ok {
				continue
			}
			copy(batchEmbeddings.Row(cursor), batchEmbeddings.Row(idx))
			metadata = append(metadata, embeddings.RepoEmbeddingRowMetadata{
				FileName:  chunk.FileName,
				StartLine: chunk.StartLine,
				EndLine:   chunk.EndLine,
			})

			cursor++
		}

		if err := insert(metadata, batchEmbeddings.Embeddings[:cursor*dimensions]); err != nil {
			return false, err
		}

		// if the file currently being processed (i.e. the file for the most recently batched chunk)
		// is flagged for exclusion then notify caller to skip the file's subsequent chunks
		var excludeCurrentFile bool

		if _, excludeCurrentFile = excludedFileNames[batch[len(batch)-1].FileName]; excludeCurrentFile {
			// current file was not indexed this flush and was truncated from index (if indexed by a previous flush)
			partiallyIndexedFileCount = 0
		} else {
			// current file could still fail later
			partiallyIndexedFileCount = currentFileBatchedCount
		}

		batch = batch[:0] // reset batch
		reportProgress(stats)
		return excludeCurrentFile, nil
	}

	addToBatch := func(chunk codeintelContext.EmbeddableChunk) (bool, error) {
		batch = append(batch, chunk)
		if len(batch) >= batchSize {
			// Flush if we've hit batch size
			return flush()
		}
		return false, nil
	}

fileLoop:
	for _, file := range files {
		currentFileBatchedCount = 0
		partiallyIndexedFileCount = 0

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
		var size int
		for _, chunk := range chunks {
			if excludeCurrentFile, err := addToBatch(chunk); err != nil {
				return bgrepo.EmbedFilesStats{}, err
			} else if excludeCurrentFile {
				// current file received an embedding error in the last flush()
				continue fileLoop
			}

			size += len(chunk.Content)
			currentFileBatchedCount++
		}

		stats.AddChunks(len(chunks), size)
		stats.AddFile()
	}

	// Always do a final flush
	if _, err := flush(); err != nil {
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
