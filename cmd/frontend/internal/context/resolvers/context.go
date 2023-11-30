package resolvers

import (
	"context"

	"github.com/sourcegraph/conc/iter"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codycontext "github.com/sourcegraph/sourcegraph/internal/codycontext"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewResolver(db database.DB, gitserverClient gitserver.Client, contextClient *codycontext.CodyContextClient) graphqlbackend.CodyContextResolver {
	return &Resolver{
		db:              db,
		gitserverClient: gitserverClient,
		contextClient:   contextClient,
	}
}

type Resolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	contextClient   *codycontext.CodyContextClient
}

func (r *Resolver) GetCodyContext(ctx context.Context, args graphqlbackend.GetContextArgs) (_ []graphqlbackend.ContextResultResolver, err error) {
	repoIDs, err := graphqlbackend.UnmarshalRepositoryIDs(args.Repos)
	if err != nil {
		return nil, err
	}

	repos, err := r.db.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repoNameIDs := make([]types.RepoIDName, len(repoIDs))
	for i, repoID := range repoIDs {
		repo, ok := repos[repoID]
		if !ok {
			// GetReposSetByIDs does not error if a repo could not be found.
			return nil, errors.Newf("could not find repo with id %d", int32(repoID))
		}

		repoNameIDs[i] = types.RepoIDName{ID: repoID, Name: repo.Name}
	}

	fileChunks, err := r.contextClient.GetCodyContext(ctx, codycontext.GetContextArgs{
		Repos:            repoNameIDs,
		Query:            args.Query,
		CodeResultsCount: args.CodeResultsCount,
		TextResultsCount: args.TextResultsCount,
	})
	if err != nil {
		return nil, err
	}

	tr, ctx := trace.New(ctx, "resolveChunks")
	defer tr.EndWithErr(&err)

	return iter.MapErr(fileChunks, func(fileChunk *codycontext.FileChunkContext) (graphqlbackend.ContextResultResolver, error) {
		return r.fileChunkToResolver(ctx, fileChunk)
	})
}

func (r *Resolver) fileChunkToResolver(ctx context.Context, chunk *codycontext.FileChunkContext) (graphqlbackend.ContextResultResolver, error) {
	repoResolver := graphqlbackend.NewRepositoryResolver(r.db, r.gitserverClient, &types.Repo{
		ID:   chunk.RepoID,
		Name: chunk.RepoName,
	})

	commitResolver := graphqlbackend.NewGitCommitResolver(r.db, r.gitserverClient, repoResolver, chunk.CommitID, nil)
	stat, err := r.gitserverClient.Stat(ctx, chunk.RepoName, chunk.CommitID, chunk.Path)
	if err != nil {
		return nil, err
	}

	gitTreeEntryResolver := graphqlbackend.NewGitTreeEntryResolver(r.db, r.gitserverClient, graphqlbackend.GitTreeEntryResolverOpts{
		Commit: commitResolver,
		Stat:   stat,
	})

	// Populate content ahead of time so we can do it concurrently
	gitTreeEntryResolver.Content(ctx, &graphqlbackend.GitTreeContentPageArgs{})
	return graphqlbackend.NewFileChunkContextResolver(gitTreeEntryResolver, chunk.StartLine, chunk.EndLine), nil
}
