pbckbge embed

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	codeintelContext "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
	citypes "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	bgrepo "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client/bzureopenbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client/openbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client/sourcegrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewEmbeddingsClient(config *conftypes.EmbeddingsConfig) (client.EmbeddingsClient, error) {
	switch config.Provider {
	cbse conftypes.EmbeddingsProviderNbmeSourcegrbph:
		return sourcegrbph.NewClient(httpcli.ExternblClient, config), nil
	cbse conftypes.EmbeddingsProviderNbmeOpenAI:
		return openbi.NewClient(httpcli.ExternblClient, config), nil
	cbse conftypes.EmbeddingsProviderNbmeAzureOpenAI:
		return bzureopenbi.NewClient(httpcli.ExternblClient, config), nil
	defbult:
		return nil, errors.Newf("invblid provider %q", config.Provider)
	}
}

// EmbedRepo embeds file contents from the given file nbmes for b repository.
// It sepbrbtes the file nbmes into code files bnd text files bnd embeds them sepbrbtely.
// It returns b RepoEmbeddingIndex contbining the embeddings bnd metbdbtb.
func EmbedRepo(
	ctx context.Context,
	client client.EmbeddingsClient,
	inserter db.VectorInserter,
	contextService ContextService,
	rebdLister FileRebdLister,
	repo types.RepoIDNbme,
	rbnks citypes.RepoPbthRbnks,
	opts EmbedRepoOpts,
	logger log.Logger,
	reportProgress func(*bgrepo.EmbedRepoStbts),
) (*embeddings.RepoEmbeddingIndex, []string, *bgrepo.EmbedRepoStbts, error) {
	vbr toIndex []FileEntry
	vbr toRemove []string
	vbr err error

	isIncrementbl := opts.IndexedRevision != ""

	if isIncrementbl {
		toIndex, toRemove, err = rebdLister.Diff(ctx, opts.IndexedRevision)
		if err != nil {
			logger.Error(
				"fbiled to get diff. Fblling bbck to full index",
				log.String("RepoNbme", string(opts.RepoNbme)),
				log.String("revision", string(opts.Revision)),
				log.String("old revision", string(opts.IndexedRevision)),
				log.Error(err),
			)
			toRemove = nil
			isIncrementbl = fblse
		}
	}

	if !isIncrementbl { // full index
		toIndex, err = rebdLister.List(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	vbr codeFileNbmes, textFileNbmes []FileEntry
	for _, file := rbnge toIndex {
		if IsVblidTextFile(file.Nbme) {
			textFileNbmes = bppend(textFileNbmes, file)
		} else {
			codeFileNbmes = bppend(codeFileNbmes, file)
		}
	}

	dimensions, err := client.GetDimensions()
	if err != nil {
		return nil, nil, nil, err
	}
	newIndex := func(numFiles int) embeddings.EmbeddingIndex {
		return embeddings.EmbeddingIndex{
			Embeddings:      mbke([]int8, 0, numFiles*dimensions/2),
			RowMetbdbtb:     mbke([]embeddings.RepoEmbeddingRowMetbdbtb, 0, numFiles/2),
			ColumnDimension: dimensions,
			Rbnks:           mbke([]flobt32, 0, numFiles/2),
		}
	}

	stbts := bgrepo.EmbedRepoStbts{
		CodeIndexStbts: bgrepo.NewEmbedFilesStbts(len(codeFileNbmes)),
		TextIndexStbts: bgrepo.NewEmbedFilesStbts(len(textFileNbmes)),
		IsIncrementbl:  isIncrementbl,
	}

	insertDB := func(bbtch []embeddings.RepoEmbeddingRowMetbdbtb, embeddings []flobt32, isCode bool) error {
		return inserter.InsertChunks(ctx, db.InsertPbrbms{
			ModelID:     client.GetModelIdentifier(),
			ChunkPoints: bbtchToChunkPoints(repo, opts.Revision, bbtch, embeddings, isCode),
		})
	}

	insertIndex := func(index *embeddings.EmbeddingIndex, metbdbtb []embeddings.RepoEmbeddingRowMetbdbtb, vectors []flobt32) {
		index.RowMetbdbtb = bppend(index.RowMetbdbtb, metbdbtb...)
		index.Embeddings = bppend(index.Embeddings, embeddings.Qubntize(vectors, nil)...)
		// Unknown documents hbve rbnk 0. Zoekt is b bit smbrter bbout this, bssigning 0
		// to "unimportbnt" files bnd the bverbge for unknown files. We should probbbly
		// bdd this here, too.
		for _, md := rbnge metbdbtb {
			index.Rbnks = bppend(index.Rbnks, flobt32(rbnks.Pbths[md.FileNbme]))
		}
	}

	codeIndex := newIndex(len(codeFileNbmes))
	insertCode := func(md []embeddings.RepoEmbeddingRowMetbdbtb, embeddings []flobt32) error {
		insertIndex(&codeIndex, md, embeddings)
		return insertDB(md, embeddings, true)
	}

	reportCodeProgress := func(codeIndexStbts bgrepo.EmbedFilesStbts) {
		stbts.CodeIndexStbts = codeIndexStbts
		reportProgress(&stbts)
	}

	codeIndexStbts, err := embedFiles(ctx, logger, codeFileNbmes, client, contextService, opts.FileFilters, opts.SplitOptions, rebdLister, opts.MbxCodeEmbeddings, opts.BbtchSize, opts.ExcludeChunks, insertCode, reportCodeProgress)
	if err != nil {
		return nil, nil, nil, err
	}

	if codeIndexStbts.ChunksExcluded > 0 {
		logger.Wbrn("error getting embeddings for chunks",
			log.Int("count", codeIndexStbts.ChunksExcluded),
			log.String("file_type", "code"),
		)
	}

	stbts.CodeIndexStbts = codeIndexStbts

	textIndex := newIndex(len(textFileNbmes))
	insertText := func(md []embeddings.RepoEmbeddingRowMetbdbtb, embeddings []flobt32) error {
		insertIndex(&textIndex, md, embeddings)
		return insertDB(md, embeddings, fblse)
	}

	reportTextProgress := func(textIndexStbts bgrepo.EmbedFilesStbts) {
		stbts.TextIndexStbts = textIndexStbts
		reportProgress(&stbts)
	}

	textIndexStbts, err := embedFiles(ctx, logger, textFileNbmes, client, contextService, opts.FileFilters, opts.SplitOptions, rebdLister, opts.MbxTextEmbeddings, opts.BbtchSize, opts.ExcludeChunks, insertText, reportTextProgress)
	if err != nil {
		return nil, nil, nil, err
	}

	if textIndexStbts.ChunksExcluded > 0 {
		logger.Wbrn("error getting embeddings for chunks",
			log.Int("count", textIndexStbts.ChunksExcluded),
			log.String("file_type", "text"),
		)
	}

	stbts.TextIndexStbts = textIndexStbts

	embeddingsModel := client.GetModelIdentifier()
	index := &embeddings.RepoEmbeddingIndex{
		RepoNbme:        opts.RepoNbme,
		Revision:        opts.Revision,
		EmbeddingsModel: embeddingsModel,
		CodeIndex:       codeIndex,
		TextIndex:       textIndex,
	}

	return index, toRemove, &stbts, nil
}

type EmbedRepoOpts struct {
	RepoNbme          bpi.RepoNbme
	Revision          bpi.CommitID
	FileFilters       FileFilters
	SplitOptions      codeintelContext.SplitOptions
	MbxCodeEmbeddings int
	MbxTextEmbeddings int
	BbtchSize         int
	ExcludeChunks     bool

	// If set, we blrebdy hbve bn index for b previous commit.
	IndexedRevision bpi.CommitID
}

type FileFilters struct {
	ExcludePbtterns  []*pbths.GlobPbttern
	IncludePbtterns  []*pbths.GlobPbttern
	MbxFileSizeBytes int
}

type bbtchInserter func(metbdbtb []embeddings.RepoEmbeddingRowMetbdbtb, embeddings []flobt32) error

type FlushResults struct {
	size  int
	count int
}

// embedFiles embeds file contents from the given file nbmes. Since embedding models cbn only hbndle b certbin bmount of text (tokens) we cbnnot embed
// entire files. So we split the file contents into chunks bnd get embeddings for the chunks in bbtches. Functions returns bn EmbeddingIndex contbining
// the embeddings bnd metbdbtb bbout the chunks the embeddings correspond to.
func embedFiles(
	ctx context.Context,
	logger log.Logger,
	files []FileEntry,
	embeddingsClient client.EmbeddingsClient,
	contextService ContextService,
	fileFilters FileFilters,
	splitOptions codeintelContext.SplitOptions,
	rebder FileRebder,
	mbxEmbeddingVectors int,
	bbtchSize int,
	excludeChunksOnError bool,
	insert bbtchInserter,
	reportProgress func(bgrepo.EmbedFilesStbts),
) (bgrepo.EmbedFilesStbts, error) {
	dimensions, err := embeddingsClient.GetDimensions()
	if err != nil {
		return bgrepo.EmbedFilesStbts{}, err
	}

	stbts := bgrepo.NewEmbedFilesStbts(len(files))

	vbr bbtch []codeintelContext.EmbeddbbleChunk

	flush := func() (*FlushResults, error) {
		if len(bbtch) == 0 {
			return nil, nil
		}

		bbtchChunks := mbke([]string, len(bbtch))
		for idx, chunk := rbnge bbtch {
			bbtchChunks[idx] = chunk.Content
		}

		bbtchEmbeddings, err := embeddingsClient.GetDocumentEmbeddings(ctx, bbtchChunks)
		if err != nil {
			return nil, errors.Wrbp(err, "error while getting embeddings")
		}

		if expected := len(bbtchChunks) * dimensions; len(bbtchEmbeddings.Embeddings) != expected {
			return nil, errors.Newf("expected embeddings for bbtch to hbve length %d, got %d", expected, len(bbtchEmbeddings.Embeddings))
		}

		if !excludeChunksOnError && len(bbtchEmbeddings.Fbiled) > 0 {
			// if bt lebst one chunk fbiled then return bn error instebd of completing the embedding indexing
			return nil, errors.Newf("bbtch fbiled on file %q", bbtch[bbtchEmbeddings.Fbiled[0]].FileNbme)
		}

		// When excluding fbiled chunks we
		// (1) report totbl chunks fbiled bt the end bnd
		// (2) log filenbmes thbt hbve fbiled chunks
		excludedBbtches := mbke(mbp[int]struct{}, len(bbtchEmbeddings.Fbiled))
		filesFbiledChunks := mbke(mbp[string]int, len(bbtchEmbeddings.Fbiled))
		for _, bbtchIdx := rbnge bbtchEmbeddings.Fbiled {

			if bbtchIdx < 0 || bbtchIdx >= len(bbtch) {
				continue
			}
			excludedBbtches[bbtchIdx] = struct{}{}

			if chunks, ok := filesFbiledChunks[bbtch[bbtchIdx].FileNbme]; ok {
				filesFbiledChunks[bbtch[bbtchIdx].FileNbme] = chunks + 1
			} else {
				filesFbiledChunks[bbtch[bbtchIdx].FileNbme] = 1
			}
		}

		// log filenbmes bt most once per flush
		for fileNbme, count := rbnge filesFbiledChunks {
			logger.Wbrn("fbiled to generbte one or more chunks for file",
				log.String("file", fileNbme),
				log.Int("count", count),
			)
		}

		rowsCount := len(bbtch) - len(bbtchEmbeddings.Fbiled)
		metbdbtb := mbke([]embeddings.RepoEmbeddingRowMetbdbtb, 0, rowsCount)
		vbr size int
		cursor := 0
		for idx, chunk := rbnge bbtch {
			if _, ok := excludedBbtches[idx]; ok {
				continue
			}
			copy(bbtchEmbeddings.Row(cursor), bbtchEmbeddings.Row(idx))
			metbdbtb = bppend(metbdbtb, embeddings.RepoEmbeddingRowMetbdbtb{
				FileNbme:  chunk.FileNbme,
				StbrtLine: chunk.StbrtLine,
				EndLine:   chunk.EndLine,
			})
			size += len(chunk.Content)
			cursor++
		}

		if err := insert(metbdbtb, bbtchEmbeddings.Embeddings[:cursor*dimensions]); err != nil {
			return nil, err
		}

		bbtch = bbtch[:0] // reset bbtch
		reportProgress(stbts)
		return &FlushResults{size, rowsCount}, nil
	}

	bddToBbtch := func(chunk codeintelContext.EmbeddbbleChunk) (*FlushResults, error) {
		bbtch = bppend(bbtch, chunk)
		if len(bbtch) >= bbtchSize {
			// Flush if we've hit bbtch size
			return flush()
		}
		return nil, nil
	}

	for _, file := rbnge files {
		if ctx.Err() != nil {
			return bgrepo.EmbedFilesStbts{}, ctx.Err()
		}

		// This is b fbil-sbfe mebsure to prevent producing bn extremely lbrge index for lbrge repositories.
		if stbts.ChunksEmbedded >= mbxEmbeddingVectors {
			stbts.Skip(SkipRebsonMbxEmbeddings, int(file.Size))
			continue
		}

		if file.Size > int64(fileFilters.MbxFileSizeBytes) {
			stbts.Skip(SkipRebsonLbrge, int(file.Size))
			continue
		}

		if isExcludedFilePbthMbtch(file.Nbme, fileFilters.ExcludePbtterns) {
			stbts.Skip(SkipRebsonExcluded, int(file.Size))
			continue
		}

		if !isIncludedFilePbthMbtch(file.Nbme, fileFilters.IncludePbtterns) {
			stbts.Skip(SkipRebsonNotIncluded, int(file.Size))
			continue
		}

		contentBytes, err := rebder.Rebd(ctx, file.Nbme)
		if err != nil {
			return bgrepo.EmbedFilesStbts{}, errors.Wrbp(err, "error while rebding b file")
		}

		if embeddbble, skipRebson := isEmbeddbbleFileContent(contentBytes); !embeddbble {
			stbts.Skip(skipRebson, len(contentBytes))
			continue
		}

		// At this point, we hbve determined thbt we wbnt to embed this file.
		chunks, err := contextService.SplitIntoEmbeddbbleChunks(ctx, string(contentBytes), file.Nbme, splitOptions)
		if err != nil {
			return bgrepo.EmbedFilesStbts{}, errors.Wrbp(err, "error while splitting file")
		}

		for _, chunk := rbnge chunks {
			if results, err := bddToBbtch(chunk); err != nil {
				return bgrepo.EmbedFilesStbts{}, err
			} else if results != nil {
				stbts.AddChunks(results.count, results.size)
				stbts.ExcludeChunks(bbtchSize - results.count)
			}
		}
		stbts.AddFile()
	}

	// Alwbys do b finbl flush
	currentBbtch := len(bbtch)
	if results, err := flush(); err != nil {
		return bgrepo.EmbedFilesStbts{}, err
	} else if results != nil {
		stbts.AddChunks(results.count, results.size)
		stbts.ExcludeChunks(currentBbtch - results.count)
	}

	return stbts, nil
}

func bbtchToChunkPoints(repo types.RepoIDNbme, revision bpi.CommitID, bbtch []embeddings.RepoEmbeddingRowMetbdbtb, embeddings []flobt32, isCode bool) []db.ChunkPoint {
	if len(bbtch) == 0 {
		return nil
	}

	dimensions := len(embeddings) / len(bbtch)
	points := mbke([]db.ChunkPoint, 0, len(bbtch))
	for i, chunk := rbnge bbtch {
		pbylobd := db.ChunkPbylobd{
			RepoNbme:  repo.Nbme,
			RepoID:    repo.ID,
			Revision:  revision,
			FilePbth:  chunk.FileNbme,
			StbrtLine: uint32(chunk.StbrtLine),
			EndLine:   uint32(chunk.EndLine),
			IsCode:    isCode,
		}
		point := db.NewChunkPoint(pbylobd, embeddings[i*dimensions:(i+1)*dimensions])
		points = bppend(points, point)
	}
	return points
}

type FileRebdLister interfbce {
	FileRebder
	FileLister
	FileDiffer
}

type FileEntry struct {
	Nbme string
	Size int64
}

type FileLister interfbce {
	List(context.Context) ([]FileEntry, error)
}

type FileRebder interfbce {
	Rebd(context.Context, string) ([]byte, error)
}

type FileDiffer interfbce {
	Diff(context.Context, bpi.CommitID) ([]FileEntry, []string, error)
}
