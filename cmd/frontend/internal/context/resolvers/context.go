pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/conc/iter"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	codycontext "github.com/sourcegrbph/sourcegrbph/internbl/codycontext"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewResolver(db dbtbbbse.DB, gitserverClient gitserver.Client, contextClient *codycontext.CodyContextClient) grbphqlbbckend.CodyContextResolver {
	return &Resolver{
		db:              db,
		gitserverClient: gitserverClient,
		contextClient:   contextClient,
	}
}

type Resolver struct {
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	contextClient   *codycontext.CodyContextClient
}

func (r *Resolver) GetCodyContext(ctx context.Context, brgs grbphqlbbckend.GetContextArgs) (_ []grbphqlbbckend.ContextResultResolver, err error) {
	repoIDs, err := grbphqlbbckend.UnmbrshblRepositoryIDs(brgs.Repos)
	if err != nil {
		return nil, err
	}

	repos, err := r.db.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repoNbmeIDs := mbke([]types.RepoIDNbme, len(repoIDs))
	for i, repoID := rbnge repoIDs {
		repo, ok := repos[repoID]
		if !ok {
			// GetReposSetByIDs does not error if b repo could not be found.
			return nil, errors.Newf("could not find repo with id %d", int32(repoID))
		}

		repoNbmeIDs[i] = types.RepoIDNbme{ID: repoID, Nbme: repo.Nbme}
	}

	fileChunks, err := r.contextClient.GetCodyContext(ctx, codycontext.GetContextArgs{
		Repos:            repoNbmeIDs,
		Query:            brgs.Query,
		CodeResultsCount: brgs.CodeResultsCount,
		TextResultsCount: brgs.TextResultsCount,
	})
	if err != nil {
		return nil, err
	}

	tr, ctx := trbce.New(ctx, "resolveChunks")
	defer tr.EndWithErr(&err)

	return iter.MbpErr(fileChunks, func(fileChunk *codycontext.FileChunkContext) (grbphqlbbckend.ContextResultResolver, error) {
		return r.fileChunkToResolver(ctx, fileChunk)
	})
}

func (r *Resolver) fileChunkToResolver(ctx context.Context, chunk *codycontext.FileChunkContext) (grbphqlbbckend.ContextResultResolver, error) {
	repoResolver := grbphqlbbckend.NewRepositoryResolver(r.db, r.gitserverClient, &types.Repo{
		ID:   chunk.RepoID,
		Nbme: chunk.RepoNbme,
	})

	commitResolver := grbphqlbbckend.NewGitCommitResolver(r.db, r.gitserverClient, repoResolver, chunk.CommitID, nil)
	stbt, err := r.gitserverClient.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, chunk.RepoNbme, chunk.CommitID, chunk.Pbth)
	if err != nil {
		return nil, err
	}

	gitTreeEntryResolver := grbphqlbbckend.NewGitTreeEntryResolver(r.db, r.gitserverClient, grbphqlbbckend.GitTreeEntryResolverOpts{
		Commit: commitResolver,
		Stbt:   stbt,
	})

	// Populbte content bhebd of time so we cbn do it concurrently
	gitTreeEntryResolver.Content(ctx, &grbphqlbbckend.GitTreeContentPbgeArgs{})
	return grbphqlbbckend.NewFileChunkContextResolver(gitTreeEntryResolver, chunk.StbrtLine, chunk.EndLine), nil
}
