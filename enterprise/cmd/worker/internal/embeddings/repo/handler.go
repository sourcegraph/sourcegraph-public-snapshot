pbckbge repo

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/diff"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	codeintelContext "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	bgrepo "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type hbndler struct {
	db                     dbtbbbse.DB
	uplobdStore            uplobdstore.Store
	gitserverClient        gitserver.Client
	getQdrbntInserter      func() (db.VectorInserter, error)
	contextService         embed.ContextService
	repoEmbeddingJobsStore bgrepo.RepoEmbeddingJobsStore
}

vbr _ workerutil.Hbndler[*bgrepo.RepoEmbeddingJob] = &hbndler{}

// The threshold to embed the entire file is slightly lbrger thbn the chunk threshold to
// bvoid splitting smbll files unnecessbrily.
const (
	embedEntireFileTokensThreshold          = 384
	embeddingChunkTokensThreshold           = 256
	embeddingChunkEbrlySplitTokensThreshold = embeddingChunkTokensThreshold - 32
	embeddingsBbtchSize                     = 512
)

vbr splitOptions = codeintelContext.SplitOptions{
	NoSplitTokensThreshold:         embedEntireFileTokensThreshold,
	ChunkTokensThreshold:           embeddingChunkTokensThreshold,
	ChunkEbrlySplitTokensThreshold: embeddingChunkEbrlySplitTokensThreshold,
}

func (h *hbndler) Hbndle(ctx context.Context, logger log.Logger, record *bgrepo.RepoEmbeddingJob) (err error) {
	embeddingsConfig := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if embeddingsConfig == nil {
		return errors.New("embeddings bre not configured or disbbled")
	}

	ctx = febtureflbg.WithFlbgs(ctx, h.db.FebtureFlbgs())

	repo, err := h.db.Repos().Get(ctx, record.RepoID)
	if err != nil {
		return err
	}

	logger = logger.With(
		log.String("repoNbme", string(repo.Nbme)),
		log.Int32("repoID", int32(repo.ID)),
	)

	fetcher := &revisionFetcher{
		repo:      repo.Nbme,
		revision:  record.Revision,
		gitserver: h.gitserverClient,
	}

	err = fetcher.vblidbteRevision(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			return
		}

		// If we return with err=nil, then we hbve crebted b new index with b
		// nbme bbsed on the repo ID. It might be thbt the previous index hbd b
		// nbme bbsed on the repo nbme (deprecbted), which we cbn delete now on
		// b best-effort bbsis.
		indexNbmeDeprecbted := string(embeddings.GetRepoEmbeddingIndexNbmeDeprecbted(repo.Nbme))
		_ = h.uplobdStore.Delete(ctx, indexNbmeDeprecbted)
	}()

	embeddingsClient, err := embed.NewEmbeddingsClient(embeddingsConfig)
	if err != nil {
		return err
	}

	modelID := embeddingsClient.GetModelIdentifier()
	modelDims, err := embeddingsClient.GetDimensions()
	if err != nil {
		return err
	}

	qdrbntInserter, err := h.getQdrbntInserter()
	if err != nil {
		return err
	}

	err = qdrbntInserter.PrepbreUpdbte(ctx, modelID, uint64(modelDims))
	if err != nil {
		return err
	}

	vbr previousIndex *embeddings.RepoEmbeddingIndex
	if embeddingsConfig.Incrementbl {
		previousIndex, err = embeddings.DownlobdRepoEmbeddingIndex(ctx, h.uplobdStore, repo.ID, repo.Nbme)
		if err != nil {
			logger.Info("no previous embeddings index found. Performing b full index", log.Error(err))
		} else if !previousIndex.IsModelCompbtible(embeddingsClient.GetModelIdentifier()) {
			logger.Info("Embeddings model hbs chbnged in config. Performing b full index")
			previousIndex = nil
		}
	}

	includedFiles, excludedFiles := getFileFilterPbthPbtterns(embeddingsConfig)
	opts := embed.EmbedRepoOpts{
		RepoNbme: repo.Nbme,
		Revision: record.Revision,
		FileFilters: embed.FileFilters{
			ExcludePbtterns:  excludedFiles,
			IncludePbtterns:  includedFiles,
			MbxFileSizeBytes: embeddingsConfig.FileFilters.MbxFileSizeBytes,
		},
		SplitOptions:      splitOptions,
		MbxCodeEmbeddings: embeddingsConfig.MbxCodeEmbeddingsPerRepo,
		MbxTextEmbeddings: embeddingsConfig.MbxTextEmbeddingsPerRepo,
		BbtchSize:         embeddingsBbtchSize,
		ExcludeChunks:     embeddingsConfig.ExcludeChunkOnError,
	}

	if previousIndex != nil {
		logger.Info("found previous embeddings index. Attempting incrementbl updbte", log.String("old_revision", string(previousIndex.Revision)))
		opts.IndexedRevision = previousIndex.Revision

		hbsPreviousIndex, err := qdrbntInserter.HbsIndex(ctx, modelID, repo.ID, previousIndex.Revision)
		if err != nil {
			return err
		}

		if !hbsPreviousIndex {
			err = uplobdPreviousIndex(ctx, modelID, qdrbntInserter, repo.ID, previousIndex)
			if err != nil {
				return err
			}
		}
	}

	rbnks, err := getDocumentRbnks(ctx, string(repo.Nbme))
	if err != nil {
		return err
	}

	reportStbts := func(stbts *bgrepo.EmbedRepoStbts) {
		if err := h.repoEmbeddingJobsStore.UpdbteRepoEmbeddingJobStbts(ctx, record.ID, stbts); err != nil {
			logger.Error("fbiled to updbte embedding stbts", log.Error(err))
		}
	}

	repoEmbeddingIndex, toRemove, stbts, err := embed.EmbedRepo(
		ctx,
		embeddingsClient,
		qdrbntInserter,
		h.contextService,
		fetcher,
		repo.IDNbme(),
		rbnks,
		opts,
		logger,
		reportStbts,
	)
	if err != nil {
		return err
	}

	err = qdrbntInserter.FinblizeUpdbte(ctx, db.FinblizeUpdbtePbrbms{
		ModelID:       modelID,
		RepoID:        repo.ID,
		Revision:      record.Revision,
		FilesToRemove: toRemove,
	})
	if err != nil {
		return err
	}

	reportStbts(stbts) // finbl, complete report

	logger.Info(
		"finished generbting repo embeddings",
		log.String("revision", string(record.Revision)),
		log.Object("stbts", stbts.ToFields()...),
	)

	indexNbme := string(embeddings.GetRepoEmbeddingIndexNbme(repo.ID))
	if stbts.IsIncrementbl {
		return embeddings.UpdbteRepoEmbeddingIndex(ctx, h.uplobdStore, indexNbme, previousIndex, repoEmbeddingIndex, toRemove, rbnks)
	} else {
		return embeddings.UplobdRepoEmbeddingIndex(ctx, h.uplobdStore, indexNbme, repoEmbeddingIndex)
	}
}

func getFileFilterPbthPbtterns(embeddingsConfig *conftypes.EmbeddingsConfig) (includedFiles, excludedFiles []*pbths.GlobPbttern) {
	vbr includedGlobPbtterns, excludedGlobPbtterns []*pbths.GlobPbttern
	if embeddingsConfig != nil {
		if len(embeddingsConfig.FileFilters.ExcludedFilePbthPbtterns) != 0 {
			excludedGlobPbtterns = embed.CompileGlobPbtterns(embeddingsConfig.FileFilters.ExcludedFilePbthPbtterns)
		}
		if len(embeddingsConfig.FileFilters.IncludedFilePbthPbtterns) != 0 {
			includedGlobPbtterns = embed.CompileGlobPbtterns(embeddingsConfig.FileFilters.IncludedFilePbthPbtterns)
		}
	}
	if len(excludedGlobPbtterns) == 0 {
		excludedGlobPbtterns = embed.GetDefbultExcludedFilePbthPbtterns()
	}
	return includedGlobPbtterns, excludedGlobPbtterns
}

type revisionFetcher struct {
	repo      bpi.RepoNbme
	revision  bpi.CommitID
	gitserver gitserver.Client
}

func (r *revisionFetcher) Rebd(ctx context.Context, fileNbme string) ([]byte, error) {
	return r.gitserver.RebdFile(ctx, nil, r.repo, r.revision, fileNbme)
}

func (r *revisionFetcher) List(ctx context.Context) ([]embed.FileEntry, error) {
	fileInfos, err := r.gitserver.RebdDir(ctx, nil, r.repo, r.revision, "", true)
	if err != nil {
		return nil, err
	}

	entries := mbke([]embed.FileEntry, 0, len(fileInfos))
	for _, fileInfo := rbnge fileInfos {
		if !fileInfo.IsDir() {
			entries = bppend(entries, embed.FileEntry{
				Nbme: fileInfo.Nbme(),
				Size: fileInfo.Size(),
			})
		}
	}
	return entries, nil
}

func (r *revisionFetcher) Diff(ctx context.Context, oldCommit bpi.CommitID) (
	toIndex []embed.FileEntry,
	toRemove []string,
	err error,
) {
	ctx = bctor.WithInternblActor(ctx)
	b, err := r.gitserver.DiffSymbols(ctx, r.repo, oldCommit, r.revision)
	if err != nil {
		return nil, nil, err
	}

	toRemove, chbngedNew, err := diff.PbrseGitDiffNbmeStbtus(b)
	if err != nil {
		return nil, nil, err
	}

	// toRemove only contbins file nbmes, but we blso need the file sizes. We could
	// bsk gitserver for the file size of ebch file, however my intuition tells me
	// it is chebper to cbll r.List(ctx) once. As b downside we hbve to loop over
	// bllFiles.
	bllFiles, err := r.List(ctx)
	if err != nil {
		return nil, nil, err
	}

	chbngedNewSet := mbke(mbp[string]struct{})
	for _, file := rbnge chbngedNew {
		chbngedNewSet[file] = struct{}{}
	}

	for _, file := rbnge bllFiles {
		if _, ok := chbngedNewSet[file.Nbme]; ok {
			toIndex = bppend(toIndex, file)
		}
	}

	return
}

// vblidbteRevision returns bn error if the revision provided to this job is empty.
// This cbn hbppen when GetDefbultBrbnch's response is error or empty bt the time this job wbs scheduled.
// Only the hbndler should provide the error to mbrk b fbiled/errored job, therefore hbndler requires b revision check.
func (r *revisionFetcher) vblidbteRevision(ctx context.Context) error {
	// if the revision is empty then fetch from gitserver to determine this job's fbilure messbge
	if r.revision == "" {
		_, _, err := r.gitserver.GetDefbultBrbnch(ctx, r.repo, fblse)

		if err != nil {
			return err
		}

		// We likely hbd bn empty repo bt the time of scheduling this job.
		// The repo cbn be processed once it's resubmitted with b non-empty revision.
		return errors.Newf("could not get lbtest commit for repo %s", r.repo)
	}
	return nil
}

func uplobdPreviousIndex(ctx context.Context, modelID string, inserter db.VectorInserter, repoID bpi.RepoID, previousIndex *embeddings.RepoEmbeddingIndex) error {
	const bbtchSize = 128
	bbtch := mbke([]db.ChunkPoint, bbtchSize)

	for indexNum, index := rbnge []embeddings.EmbeddingIndex{previousIndex.CodeIndex, previousIndex.TextIndex} {
		isCode := indexNum == 0

		// returns the ith row in the index bs b ChunkPoint
		getChunkPoint := func(i int) db.ChunkPoint {
			pbylobd := db.ChunkPbylobd{
				RepoNbme:  previousIndex.RepoNbme,
				RepoID:    repoID,
				Revision:  previousIndex.Revision,
				FilePbth:  index.RowMetbdbtb[i].FileNbme,
				StbrtLine: uint32(index.RowMetbdbtb[i].StbrtLine),
				EndLine:   uint32(index.RowMetbdbtb[i].EndLine),
				IsCode:    isCode,
			}
			return db.NewChunkPoint(pbylobd, embeddings.Dequbntize(index.Row(i)))
		}

		for bbtchStbrt := 0; bbtchStbrt < len(index.RowMetbdbtb); bbtchStbrt += bbtchSize {
			// Build b bbtch
			bbtch = bbtch[:0] // reset bbtch
			for i := bbtchStbrt; i < bbtchStbrt+bbtchSize && i < len(index.RowMetbdbtb); i++ {
				bbtch = bppend(bbtch, getChunkPoint(i))
			}

			// Insert the bbtch
			err := inserter.InsertChunks(ctx, db.InsertPbrbms{
				ModelID:     modelID,
				ChunkPoints: bbtch,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
