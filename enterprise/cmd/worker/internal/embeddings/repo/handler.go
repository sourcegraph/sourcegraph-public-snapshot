package repo

import (
	"bytes"
	"context"
	"sort"

	"github.com/sourcegraph/log"

	codeintelContext "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	repoembeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/embed"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type handler struct {
	db              edb.EnterpriseDB
	uploadStore     uploadstore.Store
	gitserverClient gitserver.Client
	contextService  embed.ContextService
}

var _ workerutil.Handler[*repoembeddingsbg.RepoEmbeddingJob] = &handler{}

// The threshold to embed the entire file is slightly larger than the chunk threshold to
// avoid splitting small files unnecessarily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEarlySplitTokensThreshold = embeddingChunkTokensThreshold - 32

	defaultMaxCodeEmbeddingsPerRepo = 3_072_000
	defaultMaxTextEmbeddingsPerRepo = 512_000
)

var splitOptions = codeintelContext.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEarlySplitTokensThreshold: embeddingChunkEarlySplitTokensThreshold,
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *repoembeddingsbg.RepoEmbeddingJob) error {
	if !conf.EmbeddingsEnabled() {
		return errors.New("embeddings are not configured or disabled")
	}

	ctx = featureflag.WithFlags(ctx, h.db.FeatureFlags())

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	isDelta := false
	var lastSuccessfulJobRevision api.CommitID
	if featureflag.FromContext(ctx).GetBoolOr("sh-delta-embeddings", false) {
		// Check if we should do a delta index or a full index
		lastSuccessfulJob, err := h.db.EmbeddingsJobsStore().GetEmbeddingsJob(ctx, record.RepoID)
		if err != nil {
			logger.Info("no previous successful embeddings job found. Falling back to full index")
		} else {
			isDelta = true
			lastSuccessfulJobRevision = lastSuccessfulJob.Revision
			logger.Info(
				"found previous successful embeddings job. Attempting delta index",
				log.String("old revision", string(lastSuccessfulJobRevision)),
				log.String("new revision", string(record.Revision)),
			)
		}
	}

	embeddingsClient := embed.NewEmbeddingsClient()
	fetcher := &revisionFetcher{
		repo:      repo.Name,
		revision:  record.Revision,
		gitserver: h.gitserverClient,
	}

	config := conf.Get().Embeddings
	excludedGlobPatterns := embed.GetDefaultExcludedFilePathPatterns()
	excludedGlobPatterns = append(excludedGlobPatterns, embed.CompileGlobPatterns(config.ExcludedFilePathPatterns)...)

	opts := embed.EmbedRepoOpts{
		RepoName:          repo.Name,
		Revision:          record.Revision,
		ExcludePatterns:   excludedGlobPatterns,
		SplitOptions:      splitOptions,
		MaxCodeEmbeddings: defaultTo(config.MaxCodeEmbeddingsPerRepo, defaultMaxCodeEmbeddingsPerRepo),
		MaxTextEmbeddings: defaultTo(config.MaxTextEmbeddingsPerRepo, defaultMaxTextEmbeddingsPerRepo),
	}

	if isDelta {
		opts.IndexedRevision = lastSuccessfulJobRevision
	}

	repoEmbeddingIndex, toRemove, stats, err := embed.EmbedRepo(
		ctx,
		embeddingsClient,
		h.contextService,
		fetcher,
		getDocumentRanks,
		opts,
	)
	if err != nil {
		return err
	}

	logger.Info(
		"finished generating repo embeddings",
		log.String("repoName", string(repo.Name)),
		log.String("revision", string(record.Revision)),
		log.Object("stats", stats.ToFields()...),
	)

	// This is a bit of a hack to get around the fact that we don't use toRemove yet for anything
	if isDelta && len(toRemove) > 0 {
		logger.Debug("found outdated embeddings", log.Int("count", len(toRemove)))
	}

	// TODO (stefan): If this is a delta build, we need to update the existing index, not overwrite it.
	return embeddings.UploadRepoEmbeddingIndex(ctx, h.uploadStore, string(embeddings.GetRepoEmbeddingIndexName(repo.Name)), repoEmbeddingIndex)
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

	toRemove, changedNew, err := parseGitDiffNameStatus(b)
	if err != nil {
		return nil, nil, err
	}

	// In addition to the file names, we need the file sizes. We could ask gitserver
	// for the file size of each file, however my guess it that it is cheaper to
	// call r.List(ctx) once instead of getting this information per file.
	changedNewSet := make(map[string]struct{})
	for _, file := range changedNew {
		changedNewSet[file] = struct{}{}
	}

	// r.List() gives us the file size, which we use during indexing to determine if
	// a file should be indexed or not. We only need the file size for the files in
	// changedNewSet.
	allFiles, err := r.List(ctx)
	if err != nil {
		return nil, nil, err
	}

	for _, file := range allFiles {
		if _, ok := changedNewSet[file.Name]; ok {
			toIndex = append(toIndex, file)
		}
	}

	return
}

// TODO (stefan): share this with searcher instead of duplicating it

// parseGitDiffNameStatus returns the paths changedA and changedB for commits
// A and B respectively. It expects to be parsing the output of the command
// git diff -z --name-status --no-renames A B.
func parseGitDiffNameStatus(out []byte) (changedA, changedB []string, err error) {
	if len(out) == 0 {
		return nil, nil, nil
	}

	slices := bytes.Split(bytes.TrimRight(out, "\x00"), []byte{0})
	if len(slices)%2 != 0 {
		return nil, nil, errors.New("uneven pairs")
	}

	for i := 0; i < len(slices); i += 2 {
		path := string(slices[i+1])
		switch slices[i][0] {
		case 'D': // no longer appears in B
			changedA = append(changedA, path)
		case 'M':
			changedA = append(changedA, path)
			changedB = append(changedB, path)
		case 'A': // doesn't exist in A
			changedB = append(changedB, path)
		}
	}
	sort.Strings(changedA)
	sort.Strings(changedB)

	return changedA, changedB, nil
}
