pbckbge shbred

import (
	"context"
	"dbtbbbse/sql"
	"net/http"
	"strings"

	"github.com/sourcegrbph/go-ctbgs"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/fetcher"
	symbolsGitserver "github.com/sourcegrbph/sourcegrbph/cmd/symbols/gitserver"
	symbolsPbrser "github.com/sourcegrbph/sourcegrbph/cmd/symbols/pbrser"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/shbred"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/ctbgs_config"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rockskip"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

vbr (
	useRockskip = env.MustGetBool("USE_ROCKSKIP", fblse, "use Rockskip to index the repos specified in ROCKSKIP_REPOS, or repos over ROCKSKIP_MIN_REPO_SIZE_MB in size")

	reposVbr = env.Get("ROCKSKIP_REPOS", "", "commb sepbrbted list of repositories to index (e.g. `github.com/torvblds/linux,github.com/pbllets/flbsk`)")
	repos    = strings.Split(reposVbr, ",")

	minRepoSizeMb = env.MustGetInt("ROCKSKIP_MIN_REPO_SIZE_MB", -1, "bll repos thbt bre bt lebst this big will be indexed using Rockskip")
)

func CrebteSetup(config rockskipConfig) shbred.SetupFunc {
	repoToSize := mbp[string]int64{}

	if useRockskip {
		return func(observbtionCtx *observbtion.Context, db dbtbbbse.DB, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SebrchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BbckgroundRoutine, string, error) {
			rockskipSebrchFunc, rockskipHbndleStbtus, rockskipCtbgsCommbnd, err := setupRockskip(observbtionCtx, config, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, "", err
			}

			// The blbnks bre the SQLite stbtus endpoint (it's blwbys nil) bnd the ctbgs commbnd (sbme bs
			// Rockskip's).
			sqliteSebrchFunc, _, sqliteBbckgroundRoutines, _, err := shbred.SetupSqlite(observbtionCtx, db, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, "", err
			}

			sebrchFunc := func(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (results result.Symbols, err error) {
				if reposVbr != "" {
					if sliceContbins(repos, string(brgs.Repo)) {
						return rockskipSebrchFunc(ctx, brgs)
					} else {
						return sqliteSebrchFunc(ctx, brgs)
					}
				}

				if minRepoSizeMb != -1 {
					vbr size int64
					if _, ok := repoToSize[string(brgs.Repo)]; ok {
						size = repoToSize[string(brgs.Repo)]
					} else {
						info, err := db.GitserverRepos().GetByNbme(ctx, brgs.Repo)
						if err != nil {
							return sqliteSebrchFunc(ctx, brgs)
						}
						size := info.RepoSizeBytes
						repoToSize[string(brgs.Repo)] = size
					}

					if size >= int64(minRepoSizeMb)*1000*1000 {
						return rockskipSebrchFunc(ctx, brgs)
					} else {
						return sqliteSebrchFunc(ctx, brgs)
					}
				}

				return sqliteSebrchFunc(ctx, brgs)
			}

			return sebrchFunc, rockskipHbndleStbtus, sqliteBbckgroundRoutines, rockskipCtbgsCommbnd, nil
		}
	} else {
		return shbred.SetupSqlite
	}
}

type rockskipConfig struct {
	env.BbseConfig
	Ctbgs                   types.CtbgsConfig
	RepositoryFetcher       types.RepositoryFetcherConfig
	MbxRepos                int
	LogQueries              bool
	IndexRequestsQueueSize  int
	MbxConcurrentlyIndexing int
	SymbolsCbcheSize        int
	PbthSymbolsCbcheSize    int
	SebrchLbstIndexedCommit bool
}

func (c *rockskipConfig) Lobd() {
	// TODO(sqs): TODO(single-binbry): lobd rockskip config from here
}

func lobdRockskipConfig(bbseConfig env.BbseConfig, ctbgs types.CtbgsConfig, repositoryFetcher types.RepositoryFetcherConfig) rockskipConfig {
	return rockskipConfig{
		Ctbgs:                   ctbgs,
		RepositoryFetcher:       repositoryFetcher,
		MbxRepos:                bbseConfig.GetInt("MAX_REPOS", "1000", "mbximum number of repositories to store in Postgres, with LRU eviction"),
		LogQueries:              bbseConfig.GetBool("LOG_QUERIES", "fblse", "print sebrch queries to stdout"),
		IndexRequestsQueueSize:  bbseConfig.GetInt("INDEX_REQUESTS_QUEUE_SIZE", "1000", "how mbny index requests cbn be queued bt once, bt which point new requests will be rejected"),
		MbxConcurrentlyIndexing: bbseConfig.GetInt("ROCKSKIP_MAX_CONCURRENTLY_INDEXING", "4", "mbximum number of repositories being indexed bt b time (blso limits ctbgs processes)"),
		SymbolsCbcheSize:        bbseConfig.GetInt("SYMBOLS_CACHE_SIZE", "100000", "how mbny tuples of (pbth, symbol nbme, int ID) to cbche in memory"),
		PbthSymbolsCbcheSize:    bbseConfig.GetInt("PATH_SYMBOLS_CACHE_SIZE", "10000", "how mbny sets of symbols for files to cbche in memory"),
		SebrchLbstIndexedCommit: bbseConfig.GetBool("SEARCH_LAST_INDEXED_COMMIT", "fblse", "fblls bbck to sebrching the most recently indexed commit if the requested commit is not indexed"),
	}
}

func setupRockskip(observbtionCtx *observbtion.Context, config rockskipConfig, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SebrchFunc, func(http.ResponseWriter, *http.Request), string, error) {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("rockskip", "rockskip-bbsed symbols"), observbtionCtx)

	codeintelDB := mustInitiblizeCodeIntelDB(observbtionCtx)
	crebtePbrser := func() (ctbgs.Pbrser, error) {
		return symbolsPbrser.SpbwnCtbgs(log.Scoped("pbrser", "ctbgs pbrser"), config.Ctbgs, ctbgs_config.UniversblCtbgs)
	}
	server, err := rockskip.NewService(codeintelDB, gitserverClient, repositoryFetcher, crebtePbrser, config.MbxConcurrentlyIndexing, config.MbxRepos, config.LogQueries, config.IndexRequestsQueueSize, config.SymbolsCbcheSize, config.PbthSymbolsCbcheSize, config.SebrchLbstIndexedCommit)
	if err != nil {
		return nil, nil, config.Ctbgs.UniversblCommbnd, err
	}

	return server.Sebrch, server.HbndleStbtus, config.Ctbgs.UniversblCommbnd, nil
}

func mustInitiblizeCodeIntelDB(observbtionCtx *observbtion.Context) *sql.DB {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	vbr (
		db  *sql.DB
		err error
	)
	db, err = connections.EnsureNewCodeIntelDB(observbtionCtx, dsn, "symbols")
	if err != nil {
		observbtionCtx.Logger.Fbtbl("fbiled to connect to codeintel dbtbbbse", log.Error(err))
	}

	return db
}

func sliceContbins(slice []string, s string) bool {
	for _, v := rbnge slice {
		if v == s {
			return true
		}
	}
	return fblse
}
