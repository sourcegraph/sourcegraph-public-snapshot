package repo

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/diff"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	bgrepo "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	db                     database.DB
	uploadStore            uploadstore.Store
	gitserverClient        gitserver.Client
	contextService         embed.ContextService
	repoEmbeddingJobsStore bgrepo.RepoEmbeddingJobsStore
}

var _ workerutil.Handler[*bgrepo.RepoEmbeddingJob] = &handler{}

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

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *bgrepo.RepoEmbeddingJob) error {
	embeddingsConfig := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if embeddingsConfig == nil {
		return errors.New("embeddings are not configured or disabled")
	}

	ctx = featureflag.WithFlags(ctx, h.db.FeatureFlags())

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	fetcher := &revisionFetcher{
		repo:      repo.Name,
		revision:  record.Revision,
		gitserver: h.gitserverClient,
	}

	err = fetcher.validateRevision(ctx)
	if err != nil {
		return err
	}

	embeddingsClient, err := embed.NewEmbeddingsClient(embeddingsConfig)
	if err != nil {
		return err
	}

	// lastSuccessfulJobRevision is the revision of the last successful embeddings
	// job for this repo. If we can find one, we'll attempt an incremental index,
	// otherwise we fall back to a full index.
	var lastSuccessfulJobRevision api.CommitID
	var previousIndex *embeddings.RepoEmbeddingIndex
	if embeddingsConfig.Incremental {
		lastSuccessfulJobRevision, previousIndex = h.getPreviousEmbeddingIndex(ctx, logger, repo)

		if previousIndex != nil && !previousIndex.IsModelCompatible(embeddingsClient.GetModelIdentifier()) {
			logger.Info("Embeddings model has changed in config. Performing a full index")
			lastSuccessfulJobRevision, previousIndex = "", nil
		}
	}

	includedFiles, excludedFiles := getFileFilterPathPatterns(embeddingsConfig)
	opts := embed.EmbedRepoOpts{
		RepoName: repo.Name,
		Revision: record.Revision,
		FileFilters: embed.FileFilters{
			ExcludePatterns:  excludedFiles,
			IncludePatterns:  includedFiles,
			MaxFileSizeBytes: embeddingsConfig.FileFilters.MaxFileSizeBytes,
		},
		SplitOptions:      splitOptions,
		MaxCodeEmbeddings: embeddingsConfig.MaxCodeEmbeddingsPerRepo,
		MaxTextEmbeddings: embeddingsConfig.MaxTextEmbeddingsPerRepo,
		IndexedRevision:   lastSuccessfulJobRevision,
	}

	ranks, err := getDocumentRanks(ctx, string(repo.Name))
	if err != nil {
		return err
	}

	reportStats := func(stats *bgrepo.EmbedRepoStats) {
		if err := h.repoEmbeddingJobsStore.UpdateRepoEmbeddingJobStats(ctx, record.ID, stats); err != nil {
			logger.Error("failed to update embedding stats", log.Error(err))
		}
	}

	repoEmbeddingIndex, toRemove, stats, err := embed.EmbedRepo(
		ctx,
		embeddingsClient,
		h.contextService,
		fetcher,
		ranks,
		opts,
		logger,
		reportStats,
	)
	if err != nil {
		return err
	}

	reportStats(stats) // final, complete report

	logger.Info(
		"finished generating repo embeddings",
		log.String("repoName", string(repo.Name)),
		log.String("revision", string(record.Revision)),
		log.Object("stats", stats.ToFields()...),
	)

	indexName := string(embeddings.GetRepoEmbeddingIndexName(repo.Name))
	if stats.IsIncremental {
		return embeddings.UpdateRepoEmbeddingIndex(ctx, h.uploadStore, indexName, previousIndex, repoEmbeddingIndex, toRemove, ranks)
	} else {
		return embeddings.UploadRepoEmbeddingIndex(ctx, h.uploadStore, indexName, repoEmbeddingIndex)
	}
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

// getPreviousEmbeddingIndex checks the last successfully indexed revision and returns its embeddings index. If there
// is no previous revision, or if there's a problem downloading the index, then it returns a nil index. This means we
// need to do a full (non-incremental) reindex.
func (h *handler) getPreviousEmbeddingIndex(ctx context.Context, logger log.Logger, repo *types.Repo) (api.CommitID, *embeddings.RepoEmbeddingIndex) {
	lastSuccessfulJob, err := h.repoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJob(ctx, repo.ID)
	if err != nil {
		logger.Info("No previous successful embeddings job found. Falling back to full index")
		return "", nil
	}

	indexName := string(embeddings.GetRepoEmbeddingIndexName(repo.Name))
	index, err := embeddings.DownloadRepoEmbeddingIndex(ctx, h.uploadStore, indexName)
	if err != nil {
		logger.Error("Error downloading previous embeddings index. Falling back to full index")
		return "", nil
	}

	logger.Info(
		"Found previous successful embeddings job. Attempting incremental index",
		log.String("old revision", string(lastSuccessfulJob.Revision)),
	)
	return lastSuccessfulJob.Revision, index
}

type revisionFetcher struct {
	repo      api.RepoName
	revision  api.CommitID
	gitserver gitserver.Client
}

func (r *revisionFetcher) Read(ctx context.Context, fileName string) ([]byte, error) {
	return r.gitserver.ReadFile(ctx, nil, r.repo, r.revision, fileName)
}

func (r *revisionFetcher) List(ctx context.Context) ([]embed.FileEntry, error) {
	fileInfos, err := r.gitserver.ReadDir(ctx, nil, r.repo, r.revision, "", true)
	if err != nil {
		return nil, err
	}

	entries := make([]embed.FileEntry, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			entries = append(entries, embed.FileEntry{
				Name: fileInfo.Name(),
				Size: fileInfo.Size(),
			})
		}
	}
	return entries, nil
}

func (r *revisionFetcher) Diff(ctx context.Context, oldCommit api.CommitID) (
	toIndex []embed.FileEntry,
	toRemove []string,
	err error,
) {
	ctx = actor.WithInternalActor(ctx)
	b, err := r.gitserver.DiffSymbols(ctx, r.repo, oldCommit, r.revision)
	if err != nil {
		return nil, nil, err
	}

	toRemove, changedNew, err := diff.ParseGitDiffNameStatus(b)
	if err != nil {
		return nil, nil, err
	}

	// toRemove only contains file names, but we also need the file sizes. We could
	// ask gitserver for the file size of each file, however my intuition tells me
	// it is cheaper to call r.List(ctx) once. As a downside we have to loop over
	// allFiles.
	allFiles, err := r.List(ctx)
	if err != nil {
		return nil, nil, err
	}

	changedNewSet := make(map[string]struct{})
	for _, file := range changedNew {
		changedNewSet[file] = struct{}{}
	}

	for _, file := range allFiles {
		if _, ok := changedNewSet[file.Name]; ok {
			toIndex = append(toIndex, file)
		}
	}

	return
}

// validateRevision returns an error if the revision provided to this job is empty.
// This can happen when GetDefaultBranch's response is error or empty at the time this job was scheduled.
// Only the handler should provide the error to mark a failed/errored job, therefore handler requires a revision check.
func (r *revisionFetcher) validateRevision(ctx context.Context) error {
	// if the revision is empty then fetch from gitserver to determine this job's failure message
	if r.revision == "" {
		_, _, err := r.gitserver.GetDefaultBranch(ctx, r.repo, false)

		if err != nil {
			return err
		}

		// We likely had an empty repo at the time of scheduling this job.
		// The repo can be processed once it's resubmitted with a non-empty revision.
		return errors.Newf("could not get latest commit for repo %s", r.repo)
	}
	return nil
}
