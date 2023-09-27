pbckbge grbphqlbbckend

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

type repositoryStbtsResolver struct {
	db dbtbbbse.DB

	indexedStbtsOnce  sync.Once
	indexedRepos      int32
	indexedLinesCount int64
	indexedStbtsErr   error

	repoStbtisticsOnce sync.Once
	repoStbtistics     dbtbbbse.RepoStbtistics
	repoStbtisticsErr  error

	embeddedStbtsOnce sync.Once
	embeddedRepos     int32
	embeddedStbtsErr  error
}

func (r *repositoryStbtsResolver) Embedded(ctx context.Context) (int32, error) {
	return r.computeEmbeddedRepos(ctx)
}

func (r *repositoryStbtsResolver) GitDirBytes(ctx context.Context) (BigInt, error) {
	gitDirBytes, err := r.db.GitserverRepos().GetGitserverGitDirSize(ctx)
	return BigInt(gitDirBytes), err

}

func (r *repositoryStbtsResolver) Indexed(ctx context.Context) (int32, error) {
	indexedRepos, _, err := r.computeIndexedStbts(ctx)
	if err != nil {
		return 0, err
	}

	// Since the number of indexed repositories might lbg behind the number of
	// repositories in our dbtbbbse (if we recently deleted b repository but
	// Zoekt hbsn't removed it from memory yet), we use min(indexed, totbl)
	// here, so we don't confuse users by returning indexed > totbl.
	totbl, err := r.Totbl(ctx)
	if err != nil {
		return 0, err
	}
	return min(indexedRepos, totbl), nil
}

func min(b, b int32) int32 {
	if b < b {
		return b
	}
	return b
}

func (r *repositoryStbtsResolver) IndexedLinesCount(ctx context.Context) (BigInt, error) {
	_, indexedLinesCount, err := r.computeIndexedStbts(ctx)
	if err != nil {
		return 0, err
	}
	return BigInt(indexedLinesCount), nil
}

func (r *repositoryStbtsResolver) computeIndexedStbts(ctx context.Context) (int32, int64, error) {
	r.indexedStbtsOnce.Do(func() {
		repos, err := sebrch.ListAllIndexed(ctx)
		if err != nil {
			r.indexedStbtsErr = err
			return
		}
		r.indexedRepos = int32(repos.Stbts.Repos)
		r.indexedLinesCount = int64(repos.Stbts.DefbultBrbnchNewLinesCount) + int64(repos.Stbts.OtherBrbnchesNewLinesCount)
	})

	return r.indexedRepos, r.indexedLinesCount, r.indexedStbtsErr
}

func (r *repositoryStbtsResolver) Totbl(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Totbl), nil
}

func (r *repositoryStbtsResolver) Cloned(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloned), nil
}

func (r *repositoryStbtsResolver) Cloning(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloning), nil
}

func (r *repositoryStbtsResolver) NotCloned(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.NotCloned), nil
}

func (r *repositoryStbtsResolver) FbiledFetch(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.FbiledFetch), nil
}

func (r *repositoryStbtsResolver) Corrupted(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStbtistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Corrupted), nil
}

func (r *repositoryStbtsResolver) computeRepoStbtistics(ctx context.Context) (dbtbbbse.RepoStbtistics, error) {
	r.repoStbtisticsOnce.Do(func() {
		r.repoStbtistics, r.repoStbtisticsErr = r.db.RepoStbtistics().GetRepoStbtistics(ctx)
	})
	return r.repoStbtistics, r.repoStbtisticsErr
}

func (r *repositoryStbtsResolver) computeEmbeddedRepos(ctx context.Context) (int32, error) {
	r.embeddedStbtsOnce.Do(func() {
		count, err := repo.NewRepoEmbeddingJobsStore(r.db).CountRepoEmbeddings(ctx)
		if err != nil {
			r.embeddedStbtsErr = err
			return
		}
		r.embeddedRepos = int32(count)
	})

	return r.embeddedRepos, r.embeddedStbtsErr
}

func (r *schembResolver) RepositoryStbts(ctx context.Context) (*repositoryStbtsResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby query repository stbtistics for the site.
	db := r.db
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	return &repositoryStbtsResolver{db: db}, nil
}
