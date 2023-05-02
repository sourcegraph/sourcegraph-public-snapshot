package embed

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/split"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

const GET_EMBEDDINGS_MAX_RETRIES = 5
const EMBEDDING_BATCH_SIZE = 512

const maxFileSize = 1000000 // 1MB

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEarlySplitTokensThreshold = embeddingChunkTokensThreshold - 32

	defaultMaxCodeEmbeddingsPerRepo = 3_072_000
	defaultMaxTextEmbeddingsPerRepo = 512_000
)

var splitOptions = split.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEarlySplitTokensThreshold: embeddingChunkEarlySplitTokensThreshold,
}

type ranksGetter func(ctx context.Context, repoName string) (types.RepoPathRanks, error)

// EmbedRepo embeds file contents from the given file names for a repository.
// It separates the file names into code files and text files and embeds them separately.
// It returns a RepoEmbeddingIndex containing the embeddings and metadata.
func EmbedRepo(
	ctx context.Context,
	repoName api.RepoName,
	revision api.CommitID,
	client EmbeddingsClient,
	gitserverClient gitserver.Client,
) (*embeddings.RepoEmbeddingIndex, *embeddings.EmbedRepoStats, error) {
	fetcher := &revisionFetcher{
		repo:      repoName,
		revision:  revision,
		gitserver: gitserverClient,
	}

	config := conf.Get().Embeddings
	excludedGlobPatterns := GetDefaultExcludedFilePathPatterns()
	excludedGlobPatterns = append(excludedGlobPatterns, CompileGlobPatterns(config.ExcludedFilePathPatterns)...)

	opts := embedRepoOptions{
		RepoName:          repoName,
		Revision:          revision,
		ExcludePatterns:   excludedGlobPatterns,
		SplitOptions:      splitOptions,
		MaxCodeEmbeddings: defaultTo(config.MaxCodeEmbeddingsPerRepo, defaultMaxCodeEmbeddingsPerRepo),
		MaxTextEmbeddings: defaultTo(config.MaxTextEmbeddingsPerRepo, defaultMaxTextEmbeddingsPerRepo),
	}

	return embedRepo(ctx, client, fetcher, getDocumentRanks, opts)
}

// embedRepo is a private helper method that accepts helper objects and options to control the
// embeddings process. It is useful for unit testing.
func embedRepo(
	ctx context.Context,
	client EmbeddingsClient,
	readLister FileReadLister,
	getDocumentRanks ranksGetter,
	opts embedRepoOptions,
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

type embedRepoOptions struct {
	RepoName          api.RepoName
	Revision          api.CommitID
	ExcludePatterns   []*paths.GlobPattern
	SplitOptions      split.SplitOptions
	MaxCodeEmbeddings int
	MaxTextEmbeddings int
}

func defaultTo(input, def int) int {
	if input == 0 {
		return def
	}
	return input
}

type revisionFetcher struct {
	repo      api.RepoName
	revision  api.CommitID
	gitserver gitserver.Client
}

func (r *revisionFetcher) Read(ctx context.Context, fileName string) ([]byte, error) {
	return r.gitserver.ReadFile(ctx, nil, r.repo, r.revision, fileName)
}

func (r *revisionFetcher) List(ctx context.Context) ([]FileEntry, error) {
	fileInfos, err := r.gitserver.ReadDir(ctx, nil, r.repo, r.revision, "", true)
	if err != nil {
		return nil, err
	}

	entries := make([]FileEntry, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			entries = append(entries, FileEntry{
				Name: fileInfo.Name(),
				Size: fileInfo.Size(),
			})
		}
	}
	return entries, nil
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
