package file

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
	codeintelTypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	bgfile "github.com/sourcegraph/sourcegraph/internal/embeddings/background/file"
	bgrepo "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	db                     database.DB
	uploadStore            uploadstore.Store
	gitserverClient        gitserver.Client
	contextService         embed.ContextService
	fileEmbeddingJobsStore bgfile.FileEmbeddingJobsStore
}

var _ workerutil.Handler[*bgfile.FileEmbeddingJob] = &handler{}

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEarlySplitTokensThreshold = embeddingChunkTokensThreshold - 32
)

var splitOptions = codeintelContext.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEarlySplitTokensThreshold: embeddingChunkEarlySplitTokensThreshold,
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *bgfile.FileEmbeddingJob) error {
	embeddingsConfig := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if embeddingsConfig == nil {
		return errors.New("embeddings are not configured or disabled")
	}

	ctx = featureflag.WithFlags(ctx, h.db.FeatureFlags())

	// TODO: GetEmbeddingPluginID
	plugin, err := h.db.EmbeddingPlugins().Get(ctx, record.EmbeddingPluginID)
	if err != nil {
		return err
	}

	embeddingsClient, err := embed.NewEmbeddingsClient(embeddingsConfig)
	if err != nil {
		return err
	}

	fetcher := &fileFetcher{
		plugin:      plugin.Name,
	}
	includedFiles, excludedFiles := getFileFilterPathPatterns(embeddingsConfig)
	opts := embed.EmbedFilesOpts{
		PluginName: plugin.Name,
		FileFilters: embed.FileFilters{
			ExcludePatterns:  excludedFiles,
			IncludePatterns:  includedFiles,
			MaxFileSizeBytes: embeddingsConfig.FileFilters.MaxFileSizeBytes,
		},
		MaxEmbeddings: embeddingsConfig.MaxTextEmbeddingsPerRepo,
	}

	reportStats := func(stats *bgrepo.EmbedRepoStats) {
		if err := h.fileEmbeddingJobsStore.UpdateFileEmbeddingJobStats(ctx, record.ID, stats); err != nil {
			logger.Error("failed to update embedding stats", log.Error(err))
		}
	}

	emptyRanks := codeintelTypes.RepoPathRanks{
		MeanRank: 0,
		Paths:    make(map[string]float64),
	}

	fileEmbeddingIndex, _, stats, err := embed.EmbedFiles(
		ctx,
		embeddingsClient,
		h.contextService,
		fetcher,
		emptyRanks,
		opts,
		logger,
		reportStats,
	)
	if err != nil {
		return err
	}

	reportStats(stats) // final, complete report

	logger.Info(
		"finished generating plugin embeddings",
		log.String("pluginName", string(plugin.Name)),
		log.Object("stats", stats.ToFields()...),
	)

	indexName := string(embeddings.GetFileEmbeddingIndexName(plugin.Name))
	return embeddings.UploadFileEmbeddingIndex(ctx, h.uploadStore, indexName, fileEmbeddingIndex)
}

func getFileFilterPathPatterns(embeddingsConfig *conftypes.EmbeddingsConfig) (includedFiles, excludedFiles []*paths.GlobPattern) {
	var includedGlobPatterns, excludedGlobPatterns []*paths.GlobPattern
	if embeddingsConfig != nil {
		if len(embeddingsConfig.FileFilters.ExcludedFilePathPatterns) != 0 {
			excludedGlobPatterns = embed.CompileGlobPatterns(embeddingsConfig.FileFilters.ExcludedFilePathPatterns)
		}
		if len(embeddingsConfig.FileFilters.IncludedFilePathPatterns) != 0 {
			includedGlobPatterns = embed.CompileGlobPatterns(embeddingsConfig.FileFilters.IncludedFilePathPatterns)
		}
	}
	if len(excludedGlobPatterns) == 0 {
		excludedGlobPatterns = embed.GetDefaultExcludedFilePathPatterns()
	}
	return includedGlobPatterns, excludedGlobPatterns
}

type fileFetcher struct {
	plugin      string
}

func (r *fileFetcher) Read(ctx context.Context, fileName string) ([]byte, error) {
	// TODO
	return nil, nil
}

func (r *fileFetcher) List(ctx context.Context) ([]embed.FileEntry, error) {
	// TODO
	return nil, nil
}

func (r *fileFetcher) Diff(ctx context.Context, oldCommit api.CommitID) (
	toIndex []embed.FileEntry,
	toRemove []string,
	err error,
) {
	// TODO
	return
}
