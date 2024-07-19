package repo

import (
	"context"
	"io"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelContext "github.com/sourcegraph/sourcegraph/internal/codeintel/context"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	bgrepo "github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	db                     database.DB
	uploadStore            object.Storage
	gitserverClient        gitserver.Client
	contextService         embed.ContextService
	repoEmbeddingJobsStore bgrepo.RepoEmbeddingJobsStore
	rankingService         *ranking.Service
}

var _ workerutil.Handler[*bgrepo.RepoEmbeddingJob] = &handler{}

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEarlySplitTokensThreshold = embeddingChunkTokensThreshold - 32
	embeddingsTolerableFailureRatio         = 0.10
)

var embeddingsBatchSize = env.MustGetInt("SRC_EMBEDDINGS_BATCH_SIZE", 512, "Number of chunks to embed at a time.")

var splitOptions = codeintelContext.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEarlySplitTokensThreshold: embeddingChunkEarlySplitTokensThreshold,
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *bgrepo.RepoEmbeddingJob) (err error) {
	embeddingsConfig := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if embeddingsConfig == nil {
		return nil
	}

	ctx = featureflag.WithFlags(ctx, h.db.FeatureFlags())

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	logger = logger.With(
		log.String("repoName", string(repo.Name)),
		log.Int32("repoID", int32(repo.ID)),
	)

	fetcher := &revisionFetcher{
		repo:      repo.Name,
		revision:  record.Revision,
		gitserver: h.gitserverClient,
	}

	err = fetcher.validateRevision(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			return
		}

		// If we return with err=nil, then we have created a new index with a
		// name based on the repo ID. It might be that the previous index had a
		// name based on the repo name (deprecated), which we can delete now on
		// a best-effort basis.
		indexNameDeprecated := string(embeddings.GetRepoEmbeddingIndexNameDeprecated(repo.Name))
		_ = h.uploadStore.Delete(ctx, indexNameDeprecated)
	}()

	embeddingsClient, err := embed.NewEmbeddingsClient(embeddingsConfig)
	if err != nil {
		return err
	}

	var previousIndex *embeddings.RepoEmbeddingIndex
	if embeddingsConfig.Incremental {
		previousIndex, err = embeddings.DownloadRepoEmbeddingIndex(ctx, h.uploadStore, repo.ID, repo.Name)
		if err != nil {
			logger.Info("no previous embeddings index found. Performing a full index", log.Error(err))
		} else if !previousIndex.IsModelCompatible(embeddingsClient.GetModelIdentifier()) {
			logger.Info("Embeddings model has changed in config. Performing a full index")
			previousIndex = nil
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
		SplitOptions:          splitOptions,
		MaxCodeEmbeddings:     embeddingsConfig.MaxCodeEmbeddingsPerRepo,
		MaxTextEmbeddings:     embeddingsConfig.MaxTextEmbeddingsPerRepo,
		BatchSize:             embeddingsBatchSize,
		ExcludeChunks:         embeddingsConfig.ExcludeChunkOnError,
		TolerableFailureRatio: embeddingsTolerableFailureRatio,
	}

	if previousIndex != nil {
		logger.Info("found previous embeddings index. Attempting incremental update", log.String("old_revision", string(previousIndex.Revision)))
		opts.IndexedRevision = previousIndex.Revision
	}

	ranks, err := h.rankingService.GetDocumentRanks(ctx, repo.Name)
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
		repo.IDName(),
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
		log.String("revision", string(record.Revision)),
		log.Object("stats", stats.ToFields()...),
	)

	indexName := string(embeddings.GetRepoEmbeddingIndexName(repo.ID))
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

type revisionFetcher struct {
	repo      api.RepoName
	revision  api.CommitID
	gitserver gitserver.Client
}

func (r *revisionFetcher) Read(ctx context.Context, fileName string) ([]byte, error) {
	fr, err := r.gitserver.NewFileReader(ctx, r.repo, r.revision, fileName)
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	return io.ReadAll(fr)
}

func (r *revisionFetcher) List(ctx context.Context) ([]embed.FileEntry, error) {
	it, err := r.gitserver.ReadDir(ctx, r.repo, r.revision, "", true)
	if err != nil {
		return nil, err
	}
	defer it.Close()

	entries := make([]embed.FileEntry, 0)
	for {
		fileInfo, err := it.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

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
	filesToRemove []string,
	err error,
) {
	ctx = actor.WithInternalActor(ctx)
	changedFilesIterator, err := r.gitserver.ChangedFiles(ctx, r.repo, string(oldCommit), string(r.revision))
	if err != nil {
		return nil, nil, err
	}
	defer changedFilesIterator.Close()

	var toRemove []string
	var changedNew []string

	for {
		f, err := changedFilesIterator.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, errors.Wrap(err, "iterating over changed files in git diff")
		}

		switch f.Status {
		case gitdomain.StatusDeleted:
			// Deleted since "oldCommit"
			toRemove = append(toRemove, f.Path)
		case gitdomain.StatusModified:
			// Modified in "r.revision"
			toRemove = append(toRemove, f.Path)
			changedNew = append(changedNew, f.Path)
		case gitdomain.StatusAdded:
			// Added in "r.revision"
			changedNew = append(changedNew, f.Path)
		case gitdomain.StatusTypeChanged:
			// a type change does not change the contents of a file,
			// so this is safe to ignore.
		}
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

	return toIndex, toRemove, nil
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
