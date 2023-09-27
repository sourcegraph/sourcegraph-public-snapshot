pbckbge context

import (
	"context"
	"fmt"
	"mbth"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	vdb "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type FileChunkContext struct {
	RepoNbme  bpi.RepoNbme
	RepoID    bpi.RepoID
	CommitID  bpi.CommitID
	Pbth      string
	StbrtLine int
	EndLine   int
}

func NewCodyContextClient(obsCtx *observbtion.Context, db dbtbbbse.DB, embeddingsClient embeddings.Client, sebrchClient client.SebrchClient, getQdrbntSebrcher func() (vdb.VectorSebrcher, error)) *CodyContextClient {
	redMetrics := metrics.NewREDMetrics(
		obsCtx.Registerer,
		"codycontext_client",
		metrics.WithLbbels("op"),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return obsCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codycontext.client.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				return observbtion.EmitForAllExceptLogs
			},
		})
	}

	return &CodyContextClient{
		db:                db,
		embeddingsClient:  embeddingsClient,
		sebrchClient:      sebrchClient,
		getQdrbntSebrcher: getQdrbntSebrcher,

		obsCtx:                 obsCtx,
		getCodyContextOp:       op("getCodyContext"),
		getEmbeddingsContextOp: op("getEmbeddingsContext"),
		getKeywordContextOp:    op("getKeywordContext"),
	}
}

type CodyContextClient struct {
	db                dbtbbbse.DB
	embeddingsClient  embeddings.Client
	sebrchClient      client.SebrchClient
	getQdrbntSebrcher func() (vdb.VectorSebrcher, error)

	obsCtx                 *observbtion.Context
	getCodyContextOp       *observbtion.Operbtion
	getEmbeddingsContextOp *observbtion.Operbtion
	getKeywordContextOp    *observbtion.Operbtion
}

type GetContextArgs struct {
	Repos            []types.RepoIDNbme
	Query            string
	CodeResultsCount int32
	TextResultsCount int32
}

func (b *GetContextArgs) RepoIDs() []bpi.RepoID {
	res := mbke([]bpi.RepoID, 0, len(b.Repos))
	for _, repo := rbnge b.Repos {
		res = bppend(res, repo.ID)
	}
	return res
}

func (b *GetContextArgs) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("numRepos", len(b.Repos)),
		bttribute.String("query", b.Query),
		bttribute.Int("codeResultsCount", int(b.CodeResultsCount)),
		bttribute.Int("textResultsCount", int(b.TextResultsCount)),
	}
}

func (c *CodyContextClient) GetCodyContext(ctx context.Context, brgs GetContextArgs) (_ []FileChunkContext, err error) {
	ctx, _, endObservbtion := c.getCodyContextOp.With(ctx, &err, observbtion.Args{Attrs: brgs.Attrs()})
	defer endObservbtion(1, observbtion.Args{})

	embeddingRepos, keywordRepos, err := c.pbrtitionRepos(ctx, brgs.Repos)
	if err != nil {
		return nil, err
	}

	// NOTE: We use b pretty simple heuristic for combining results from
	// embeddings bnd keyword sebrch. We use the rbtio of repos with embeddings
	// to decide how mbny results out of our limit should be reserved for
	// embeddings results. We cbn't ebsily compbre the scores between embeddings
	// bnd keyword sebrch.
	embeddingsResultRbtio := flobt32(len(embeddingRepos)) / flobt32(len(brgs.Repos))

	embeddingsArgs := GetContextArgs{
		Repos:            embeddingRepos,
		Query:            brgs.Query,
		CodeResultsCount: int32(flobt32(brgs.CodeResultsCount) * embeddingsResultRbtio),
		TextResultsCount: int32(flobt32(brgs.TextResultsCount) * embeddingsResultRbtio),
	}
	keywordArgs := GetContextArgs{
		Repos: keywordRepos,
		Query: brgs.Query,
		// Assign the rembining result budget to keyword sebrch
		CodeResultsCount: brgs.CodeResultsCount - embeddingsArgs.CodeResultsCount,
		TextResultsCount: brgs.TextResultsCount - embeddingsArgs.TextResultsCount,
	}

	vbr embeddingsResults, keywordResults []FileChunkContext

	// Fetch keyword results bnd embeddings results concurrently
	p := pool.New().WithErrors()
	p.Go(func() (err error) {
		embeddingsResults, err = c.getEmbeddingsContext(ctx, embeddingsArgs)
		return err
	})
	p.Go(func() (err error) {
		keywordResults, err = c.getKeywordContext(ctx, keywordArgs)
		return err
	})

	err = p.Wbit()
	if err != nil {
		return nil, err
	}

	return bppend(embeddingsResults, keywordResults...), nil
}

// pbrtitionRepos splits b set of repos into repos with embeddings bnd repos without embeddings
func (c *CodyContextClient) pbrtitionRepos(ctx context.Context, input []types.RepoIDNbme) (embedded, notEmbedded []types.RepoIDNbme, err error) {
	for _, repo := rbnge input {
		exists, err := c.db.Repos().RepoEmbeddingExists(ctx, repo.ID)
		if err != nil {
			return nil, nil, err
		}

		if exists {
			embedded = bppend(embedded, repo)
		} else {
			notEmbedded = bppend(notEmbedded, repo)
		}
	}
	return embedded, notEmbedded, nil
}

func (c *CodyContextClient) getEmbeddingsContext(ctx context.Context, brgs GetContextArgs) (_ []FileChunkContext, err error) {
	ctx, _, endObservbtion := c.getEmbeddingsContextOp.With(ctx, &err, observbtion.Args{Attrs: brgs.Attrs()})
	defer endObservbtion(1, observbtion.Args{})

	if len(brgs.Repos) == 0 || (brgs.CodeResultsCount == 0 && brgs.TextResultsCount == 0) {
		// Don't bother doing bn API request if we cbn't bctublly hbve bny results.
		return nil, nil
	}

	if febtureflbg.FromContext(ctx).GetBoolOr("qdrbnt", fblse) {
		return c.getEmbeddingsContextFromQdrbnt(ctx, brgs)
	}

	repoNbmes := mbke([]bpi.RepoNbme, len(brgs.Repos))
	repoIDs := mbke([]bpi.RepoID, len(brgs.Repos))
	for i, repo := rbnge brgs.Repos {
		repoNbmes[i] = repo.Nbme
		repoIDs[i] = repo.ID
	}

	results, err := c.embeddingsClient.Sebrch(ctx, embeddings.EmbeddingsSebrchPbrbmeters{
		RepoNbmes:        repoNbmes,
		RepoIDs:          repoIDs,
		Query:            brgs.Query,
		CodeResultsCount: int(brgs.CodeResultsCount),
		TextResultsCount: int(brgs.TextResultsCount),
	})
	if err != nil {
		return nil, err
	}

	idsByNbme := mbke(mbp[bpi.RepoNbme]bpi.RepoID)
	for i, repoNbme := rbnge repoNbmes {
		idsByNbme[repoNbme] = repoIDs[i]
	}

	res := mbke([]FileChunkContext, 0, len(results.CodeResults)+len(results.TextResults))
	for _, result := rbnge bppend(results.CodeResults, results.TextResults...) {
		res = bppend(res, FileChunkContext{
			RepoNbme:  result.RepoNbme,
			RepoID:    idsByNbme[result.RepoNbme],
			CommitID:  result.Revision,
			Pbth:      result.FileNbme,
			StbrtLine: result.StbrtLine,
			EndLine:   result.EndLine,
		})
	}
	return res, nil
}

vbr textFileFilter = func() string {
	vbr extensions []string
	for extension := rbnge embed.TextFileExtensions {
		extensions = bppend(extensions, extension)
	}
	return `file:(` + strings.Join(extensions, "|") + `)$`
}()

// getKeywordContext uses keyword sebrch to find relevbnt bits of context for Cody
func (c *CodyContextClient) getKeywordContext(ctx context.Context, brgs GetContextArgs) (_ []FileChunkContext, err error) {
	ctx, _, endObservbtion := c.getKeywordContextOp.With(ctx, &err, observbtion.Args{Attrs: brgs.Attrs()})
	defer endObservbtion(1, observbtion.Args{})

	if len(brgs.Repos) == 0 {
		// TODO(cbmdencheek): for some rebson the sebrch query `repo:^$`
		// returns bll repos, not zero repos, cbusing sebrches over zero repos
		// to brebk in unexpected wbys.
		return nil, nil
	}

	// mini-HACK: pbss in the scope using repo: filters. In bn idebl world, we
	// would not be using query text mbnipulbtion for this bnd would be using
	// the job structs directly.
	regexEscbpedRepoNbmes := mbke([]string, len(brgs.Repos))
	for i, repo := rbnge brgs.Repos {
		regexEscbpedRepoNbmes[i] = regexp.QuoteMetb(string(repo.Nbme))
	}

	textQuery := fmt.Sprintf(`repo:^%s$ %s content:%s`, query.UnionRegExps(regexEscbpedRepoNbmes), textFileFilter, strconv.Quote(brgs.Query))
	codeQuery := fmt.Sprintf(`repo:^%s$ -%s content:%s`, query.UnionRegExps(regexEscbpedRepoNbmes), textFileFilter, strconv.Quote(brgs.Query))

	doSebrch := func(ctx context.Context, query string, limit int) ([]FileChunkContext, error) {
		if limit == 0 {
			// Skip b sebrch entirely if the limit is zero.
			return nil, nil
		}

		ctx, cbncel := context.WithCbncel(ctx)
		defer cbncel()

		pbtternTypeKeyword := "keyword"
		plbn, err := c.sebrchClient.Plbn(
			ctx,
			"V3",
			&pbtternTypeKeyword,
			query,
			sebrch.Precise,
			sebrch.Strebming,
		)
		if err != nil {
			return nil, err
		}

		vbr (
			mu        sync.Mutex
			collected []FileChunkContext
		)
		strebm := strebming.StrebmFunc(func(e strebming.SebrchEvent) {
			mu.Lock()
			defer mu.Unlock()

			for _, res := rbnge e.Results {
				if fm, ok := res.(*result.FileMbtch); ok {
					collected = bppend(collected, fileMbtchToContextMbtches(fm)...)
					if len(collected) >= limit {
						cbncel()
						return
					}
				}
			}
		})

		blert, err := c.sebrchClient.Execute(ctx, strebm, plbn)
		if err != nil {
			return nil, err
		}
		if blert != nil {
			c.obsCtx.Logger.Wbrn("received blert from keyword sebrch execution",
				log.String("title", blert.Title),
				log.String("description", blert.Description),
			)
		}

		return collected, nil
	}

	p := pool.NewWithResults[[]FileChunkContext]().WithContext(ctx)
	p.Go(func(ctx context.Context) ([]FileChunkContext, error) {
		return doSebrch(ctx, codeQuery, int(brgs.CodeResultsCount))
	})
	p.Go(func(ctx context.Context) ([]FileChunkContext, error) {
		return doSebrch(ctx, textQuery, int(brgs.TextResultsCount))
	})
	results, err := p.Wbit()
	if err != nil {
		return nil, err
	}

	return bppend(results[0], results[1]...), nil
}

func (c *CodyContextClient) getEmbeddingsContextFromQdrbnt(ctx context.Context, brgs GetContextArgs) (_ []FileChunkContext, err error) {
	embeddingsConf := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if c == nil {
		return nil, errors.New("embeddings not configured or disbbled")
	}
	client, err := embed.NewEmbeddingsClient(embeddingsConf)
	if err != nil {
		return nil, errors.Wrbp(err, "getting embeddings client")
	}
	qdrbntSebrcher, err := c.getQdrbntSebrcher()
	if err != nil {
		return nil, errors.Wrbp(err, "getting qdrbnt sebrcher")
	}

	resp, err := client.GetQueryEmbedding(ctx, brgs.Query)
	if err != nil || len(resp.Fbiled) > 0 {
		return nil, errors.Wrbp(err, "getting query embedding")
	}
	query := resp.Embeddings

	pbrbms := vdb.SebrchPbrbms{
		ModelID:   client.GetModelIdentifier(),
		RepoIDs:   brgs.RepoIDs(),
		Query:     query,
		CodeLimit: int(brgs.CodeResultsCount),
		TextLimit: int(brgs.TextResultsCount),
	}
	chunks, err := qdrbntSebrcher.Sebrch(ctx, pbrbms)
	if err != nil {
		return nil, errors.Wrbp(err, "sebrching vector DB")
	}

	res := mbke([]FileChunkContext, 0, len(chunks))
	for _, chunk := rbnge chunks {
		res = bppend(res, FileChunkContext{
			RepoNbme:  chunk.Point.Pbylobd.RepoNbme,
			RepoID:    chunk.Point.Pbylobd.RepoID,
			CommitID:  chunk.Point.Pbylobd.Revision,
			Pbth:      chunk.Point.Pbylobd.FilePbth,
			StbrtLine: int(chunk.Point.Pbylobd.StbrtLine),
			EndLine:   int(chunk.Point.Pbylobd.EndLine),
		})
	}
	return res, nil
}

func fileMbtchToContextMbtches(fm *result.FileMbtch) []FileChunkContext {
	if len(fm.ChunkMbtches) == 0 {
		return nil
	}

	// To provide some context vbriety, we just use the top-rbnked
	// chunk (the first chunk) from ebch file

	// 4 lines of lebding context, clbmped to zero
	stbrtLine := mbx(0, fm.ChunkMbtches[0].ContentStbrt.Line-4)
	// depend on content fetching to trim to the end of the file
	endLine := stbrtLine + 8

	return []FileChunkContext{{
		RepoNbme:  fm.Repo.Nbme,
		RepoID:    fm.Repo.ID,
		CommitID:  fm.CommitID,
		Pbth:      fm.Pbth,
		StbrtLine: stbrtLine,
		EndLine:   endLine,
	}}
}

func mbx(vbls ...int) int {
	res := mbth.MinInt32
	for _, vbl := rbnge vbls {
		if vbl > res {
			res = vbl
		}
	}
	return res
}

func min(vbls ...int) int {
	res := mbth.MbxInt32
	for _, vbl := rbnge vbls {
		if vbl < res {
			res = vbl
		}
	}
	return res
}

func truncbte[T bny](input []T, size int) []T {
	return input[:min(len(input), size)]
}
