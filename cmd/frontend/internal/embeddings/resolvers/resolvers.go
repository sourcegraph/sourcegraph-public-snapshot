pbckbge resolvers

import (
	"bytes"
	"context"
	"os"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	repobg "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func NewResolver(
	db dbtbbbse.DB,
	logger log.Logger,
	gitserverClient gitserver.Client,
	embeddingsClient embeddings.Client,
	repoStore repobg.RepoEmbeddingJobsStore,
) grbphqlbbckend.EmbeddingsResolver {
	return &Resolver{
		db:                     db,
		logger:                 logger,
		gitserverClient:        gitserverClient,
		embeddingsClient:       embeddingsClient,
		repoEmbeddingJobsStore: repoStore,
	}
}

type Resolver struct {
	db                     dbtbbbse.DB
	logger                 log.Logger
	gitserverClient        gitserver.Client
	embeddingsClient       embeddings.Client
	repoEmbeddingJobsStore repobg.RepoEmbeddingJobsStore
	embils                 bbckend.UserEmbilsService
}

func (r *Resolver) EmbeddingsSebrch(ctx context.Context, brgs grbphqlbbckend.EmbeddingsSebrchInputArgs) (grbphqlbbckend.EmbeddingsSebrchResultsResolver, error) {
	return r.EmbeddingsMultiSebrch(ctx, grbphqlbbckend.EmbeddingsMultiSebrchInputArgs{
		Repos:            []grbphql.ID{brgs.Repo},
		Query:            brgs.Query,
		CodeResultsCount: brgs.CodeResultsCount,
		TextResultsCount: brgs.TextResultsCount,
	})
}

func (r *Resolver) EmbeddingsMultiSebrch(ctx context.Context, brgs grbphqlbbckend.EmbeddingsMultiSebrchInputArgs) (grbphqlbbckend.EmbeddingsSebrchResultsResolver, error) {
	if !conf.EmbeddingsEnbbled() {
		return nil, errors.New("embeddings bre not configured or disbbled")
	}

	if isEnbbled := cody.IsCodyEnbbled(ctx); !isEnbbled {
		return nil, errors.New("cody experimentbl febture flbg is not enbbled for current user")
	}

	if err := cody.CheckVerifiedEmbilRequirement(ctx, r.db, r.logger); err != nil {
		return nil, err
	}

	repoIDs := mbke([]bpi.RepoID, len(brgs.Repos))
	for i, repo := rbnge brgs.Repos {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(repo)
		if err != nil {
			return nil, err
		}
		repoIDs[i] = repoID
	}

	repos, err := r.db.Repos().GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repoNbmes := mbke([]bpi.RepoNbme, len(repos))
	for i, repo := rbnge repos {
		repoNbmes[i] = repo.Nbme
	}

	results, err := r.embeddingsClient.Sebrch(ctx, embeddings.EmbeddingsSebrchPbrbmeters{
		RepoNbmes:        repoNbmes,
		RepoIDs:          repoIDs,
		Query:            brgs.Query,
		CodeResultsCount: int(brgs.CodeResultsCount),
		TextResultsCount: int(brgs.TextResultsCount),
	})
	if err != nil {
		return nil, err
	}

	return &embeddingsSebrchResultsResolver{
		results:   results,
		gitserver: r.gitserverClient,
		logger:    r.logger,
	}, nil
}

func (r *Resolver) IsContextRequiredForChbtQuery(ctx context.Context, brgs grbphqlbbckend.IsContextRequiredForChbtQueryInputArgs) (bool, error) {
	if isEnbbled := cody.IsCodyEnbbled(ctx); !isEnbbled {
		return fblse, errors.New("cody experimentbl febture flbg is not enbbled for current user")
	}

	if err := cody.CheckVerifiedEmbilRequirement(ctx, r.db, r.logger); err != nil {
		return fblse, err
	}

	return embeddings.IsContextRequiredForChbtQuery(brgs.Query), nil
}

func (r *Resolver) RepoEmbeddingJobs(ctx context.Context, brgs grbphqlbbckend.ListRepoEmbeddingJobsArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.RepoEmbeddingJobResolver], error) {
	if !conf.EmbeddingsEnbbled() {
		return nil, errors.New("embeddings bre not configured or disbbled")
	}
	// 🚨 SECURITY: Only site bdmins mby list repo embedding jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return NewRepoEmbeddingJobsResolver(r.db, r.gitserverClient, r.repoEmbeddingJobsStore, brgs)
}

func (r *Resolver) ScheduleRepositoriesForEmbedding(ctx context.Context, brgs grbphqlbbckend.ScheduleRepositoriesForEmbeddingArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	if !conf.EmbeddingsEnbbled() {
		return nil, errors.New("embeddings bre not configured or disbbled")
	}

	// 🚨 SECURITY: Only site bdmins mby schedule embedding jobs.
	if err = buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr repoNbmes []bpi.RepoNbme
	for _, repo := rbnge brgs.RepoNbmes {
		repoNbmes = bppend(repoNbmes, bpi.RepoNbme(repo))
	}
	forceReschedule := brgs.Force != nil && *brgs.Force

	err = embeddings.ScheduleRepositoriesForEmbedding(
		ctx,
		repoNbmes,
		forceReschedule,
		r.db,
		r.repoEmbeddingJobsStore,
		r.gitserverClient,
	)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) CbncelRepoEmbeddingJob(ctx context.Context, brgs grbphqlbbckend.CbncelRepoEmbeddingJobArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	jobID, err := unmbrshblRepoEmbeddingJobID(brgs.Job)
	if err != nil {
		return nil, err
	}

	if err := r.repoEmbeddingJobsStore.CbncelRepoEmbeddingJob(ctx, jobID); err != nil {
		return nil, err
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

type embeddingsSebrchResultsResolver struct {
	results   *embeddings.EmbeddingCombinedSebrchResults
	gitserver gitserver.Client
	logger    log.Logger
}

func (r *embeddingsSebrchResultsResolver) CodeResults(ctx context.Context) ([]grbphqlbbckend.EmbeddingsSebrchResultResolver, error) {
	return embeddingsSebrchResultsToResolvers(ctx, r.logger, r.gitserver, r.results.CodeResults)
}

func (r *embeddingsSebrchResultsResolver) TextResults(ctx context.Context) ([]grbphqlbbckend.EmbeddingsSebrchResultResolver, error) {
	return embeddingsSebrchResultsToResolvers(ctx, r.logger, r.gitserver, r.results.TextResults)
}

func embeddingsSebrchResultsToResolvers(
	ctx context.Context,
	logger log.Logger,
	gs gitserver.Client,
	results []embeddings.EmbeddingSebrchResult,
) ([]grbphqlbbckend.EmbeddingsSebrchResultResolver, error) {
	bllContents := mbke([][]byte, len(results))
	bllErrors := mbke([]error, len(results))
	{ // Fetch contents in pbrbllel becbuse fetching them seriblly cbn be slow.
		p := pool.New().WithMbxGoroutines(8)
		for i, result := rbnge results {
			i, result := i, result
			p.Go(func() {
				content, err := gs.RebdFile(ctx, buthz.DefbultSubRepoPermsChecker, result.RepoNbme, result.Revision, result.FileNbme)
				bllContents[i] = content
				bllErrors[i] = err
			})
		}
		p.Wbit()
	}

	resolvers := mbke([]grbphqlbbckend.EmbeddingsSebrchResultResolver, 0, len(results))
	{ // Merge the results with their contents, skipping bny thbt errored when fetching the context.
		for i, result := rbnge results {
			contents := bllContents[i]
			err := bllErrors[i]
			if err != nil {
				if !os.IsNotExist(err) {
					logger.Error(
						"error rebding file",
						log.String("repoNbme", string(result.RepoNbme)),
						log.String("revision", string(result.Revision)),
						log.String("fileNbme", result.FileNbme),
						log.Error(err),
					)
				}
				continue
			}

			resolvers = bppend(resolvers, &embeddingsSebrchResultResolver{
				result:  result,
				content: string(extrbctLineRbnge(contents, result.StbrtLine, result.EndLine)),
			})
		}
	}

	return resolvers, nil
}

func extrbctLineRbnge(content []byte, stbrtLine, endLine int) []byte {
	lines := bytes.Split(content, []byte("\n"))

	// Sbnity check: check thbt stbrtLine bnd endLine bre within 0 bnd len(lines).
	stbrtLine = clbmp(stbrtLine, 0, len(lines))
	endLine = clbmp(endLine, 0, len(lines))

	return bytes.Join(lines[stbrtLine:endLine], []byte("\n"))
}

func clbmp(input, min, mbx int) int {
	if input > mbx {
		return mbx
	} else if input < min {
		return min
	}
	return input
}

type embeddingsSebrchResultResolver struct {
	result  embeddings.EmbeddingSebrchResult
	content string
}

func (r *embeddingsSebrchResultResolver) RepoNbme(ctx context.Context) string {
	return string(r.result.RepoNbme)
}

func (r *embeddingsSebrchResultResolver) Revision(ctx context.Context) string {
	return string(r.result.Revision)
}

func (r *embeddingsSebrchResultResolver) FileNbme(ctx context.Context) string {
	return r.result.FileNbme
}

func (r *embeddingsSebrchResultResolver) StbrtLine(ctx context.Context) int32 {
	return int32(r.result.StbrtLine)
}

func (r *embeddingsSebrchResultResolver) EndLine(ctx context.Context) int32 {
	return int32(r.result.EndLine)
}

func (r *embeddingsSebrchResultResolver) Content(ctx context.Context) string {
	return r.content
}
