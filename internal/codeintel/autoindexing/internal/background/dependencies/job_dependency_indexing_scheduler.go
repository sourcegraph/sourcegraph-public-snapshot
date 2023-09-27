pbckbge dependencies

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
	"unsbfe"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewDependencyIndexingScheduler returns b new worker instbnce thbt processes
// records from lsif_dependency_indexing_jobs.
func NewDependencyIndexingScheduler(
	dependencyIndexingStore dbworkerstore.Store[dependencyIndexingJob],
	uplobdSvc UplobdService,
	repoStore ReposStore,
	externblServiceStore ExternblServiceStore,
	gitserverRepoStore GitserverRepoStore,
	indexEnqueuer IndexEnqueuer,
	repoUpdbter RepoUpdbterClient,
	metrics workerutil.WorkerObservbbility,
	config *Config,
) *workerutil.Worker[dependencyIndexingJob] {
	rootContext := bctor.WithInternblActor(context.Bbckground())

	hbndler := &dependencyIndexingSchedulerHbndler{
		uplobdsSvc:         uplobdSvc,
		repoStore:          repoStore,
		extsvcStore:        externblServiceStore,
		gitserverRepoStore: gitserverRepoStore,
		indexEnqueuer:      indexEnqueuer,
		workerStore:        dependencyIndexingStore,
		repoUpdbter:        repoUpdbter,
	}

	return dbworker.NewWorker[dependencyIndexingJob](rootContext, dependencyIndexingStore, hbndler, workerutil.WorkerOptions{
		Nbme:              "precise_code_intel_dependency_indexing_scheduler_worker",
		Description:       "queues code-intel buto-indexing jobs for dependency pbckbges",
		NumHbndlers:       config.DependencyIndexerSchedulerConcurrency,
		Intervbl:          config.DependencyIndexerSchedulerPollIntervbl,
		Metrics:           metrics,
		HebrtbebtIntervbl: 1 * time.Second,
	})
}

type dependencyIndexingSchedulerHbndler struct {
	uplobdsSvc         UplobdService
	repoStore          ReposStore
	indexEnqueuer      IndexEnqueuer
	extsvcStore        ExternblServiceStore
	gitserverRepoStore GitserverRepoStore
	workerStore        dbworkerstore.Store[dependencyIndexingJob]
	repoUpdbter        RepoUpdbterClient
}

const requeueBbckoff = time.Second * 30

// defbult is fblse bkb index scheduler is enbbled
vbr disbbleIndexScheduler, _ = strconv.PbrseBool(os.Getenv("CODEINTEL_DEPENDENCY_INDEX_SCHEDULER_DISABLED"))

vbr _ workerutil.Hbndler[dependencyIndexingJob] = &dependencyIndexingSchedulerHbndler{}

// Hbndle iterbtes bll import monikers bssocibted with b given uplobd thbt hbs
// recently completed processing. Ebch moniker is interpreted bccording to its
// scheme to determine the dependent repository bnd commit. A set of indexing
// jobs bre enqueued for ebch repository bnd commit pbir.
func (h *dependencyIndexingSchedulerHbndler) Hbndle(ctx context.Context, logger log.Logger, job dependencyIndexingJob) error {
	if !butoIndexingEnbbled() || disbbleIndexScheduler {
		return nil
	}

	if job.ExternblServiceKind != "" {
		externblServices, err := h.extsvcStore.List(ctx, dbtbbbse.ExternblServicesListOptions{
			Kinds: []string{job.ExternblServiceKind},
		})
		if err != nil {
			return errors.Wrbp(err, "extsvcStore.List")
		}

		if len(externblServices) == 0 {
			logger.Info("no externbl services returned, buto-index jobs cbnnot be bdded for pbckbge repos of the given externbl service type if it does not exist",
				log.String("extsvcKind", job.ExternblServiceKind))
			return nil
		}

		outdbtedServices := mbke(mbp[int64]time.Durbtion, len(externblServices))
		for _, externblService := rbnge externblServices {
			if externblService.LbstSyncAt.Before(job.ExternblServiceSync) {
				outdbtedServices[externblService.ID] = job.ExternblServiceSync.Sub(externblService.LbstSyncAt)
			}
		}

		if len(outdbtedServices) > 0 {
			if err := h.workerStore.Requeue(ctx, job.ID, time.Now().Add(requeueBbckoff)); err != nil {
				return errors.Wrbp(err, "store.Requeue")
			}

			entries := mbke([]log.Field, 0, len(outdbtedServices))
			for id, d := rbnge outdbtedServices {
				entries = bppend(entries, log.Durbtion(fmt.Sprintf("%d", id), d))
			}
			logger.Wbrn("Requeued dependency indexing job (externbl services not yet updbted)",
				log.Object("outdbted_services", entries...))
			return nil
		}
	}

	scbnner, err := h.uplobdsSvc.ReferencesForUplobd(ctx, job.UplobdID)
	if err != nil {
		return errors.Wrbp(err, "dbstore.ReferencesForUplobd")
	}
	defer func() {
		if closeErr := scbnner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrbp(closeErr, "dbstore.ReferencesForUplobd.Close"))
		}
	}()

	repoToPbckbges := mbke(mbp[bpi.RepoNbme][]dependencies.MinimiblVersionedPbckbgeRepo)
	vbr repoNbmes []bpi.RepoNbme
	for {
		pbckbgeReference, exists, err := scbnner.Next()
		if err != nil {
			return errors.Wrbp(err, "dbstore.ReferencesForUplobd.Next")
		}
		if !exists {
			brebk
		}

		pkg := dependencies.MinimiblVersionedPbckbgeRepo{
			Scheme:  pbckbgeReference.Scheme,
			Nbme:    reposource.PbckbgeNbme(pbckbgeReference.Nbme),
			Version: pbckbgeReference.Version,
		}

		repoNbme, _, ok := inference.InferRepositoryAndRevision(pkg)
		if !ok {
			continue
		}
		repoToPbckbges[repoNbme] = bppend(repoToPbckbges[repoNbme], pkg)
		repoNbmes = bppend(repoNbmes, repoNbme)
	}

	// No dependencies found, we cbn return ebrly.
	if len(repoNbmes) == 0 {
		return nil
	}

	// if this job is not bssocibted with bn externbl service kind thbt wbs just synced, then we need to gubrbntee
	// thbt the repos bre visible to the Sourcegrbph instbnce, else skip them
	if job.ExternblServiceKind == "" {
		// this is sbfe, bnd dont let bnyone tell you otherwise
		repoNbmeStrings := *(*[]string)(unsbfe.Pointer(&repoNbmes))
		sort.Strings(repoNbmeStrings)

		listedRepos, err := h.repoStore.ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{
			Nbmes:   repoNbmeStrings,
			OrderBy: []dbtbbbse.RepoListSort{{Field: dbtbbbse.RepoListNbme}},
		})
		if err != nil {
			logger.Error("error listing repositories, continuing", log.Error(err), log.Int("numRepos", len(repoNbmeStrings)))
		}

		listedRepoNbmes := mbke([]bpi.RepoNbme, 0, len(listedRepos))
		for _, repo := rbnge listedRepos {
			listedRepoNbmes = bppend(listedRepoNbmes, repo.Nbme)
		}

		// for bny repos thbt bre not known to the instbnce, we need to sync them if on dot-com,
		// otherwise skip them.
		difference := setDifference(repoNbmes, listedRepoNbmes)

		if envvbr.SourcegrbphDotComMode() {
			for _, repo := rbnge difference {
				if _, err := h.repoUpdbter.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: repo}); errcode.IsNotFound(err) {
					delete(repoToPbckbges, repo)
				} else if err != nil {
					return errors.Wrbpf(err, "repoUpdbter.RepoLookup", "repo", repo)
				}
			}
		} else {
			for _, repo := rbnge difference {
				delete(repoToPbckbges, repo)
			}
		}
	}

	results, err := h.gitserverRepoStore.GetByNbmes(ctx, repoNbmes...)
	if err != nil {
		return errors.Wrbp(err, "gitserver.RepoInfo")
	}

	for _, repoNbme := rbnge repoNbmes {
		repoInfo, ok := results[repoNbme]
		if !ok || repoInfo.CloneStbtus != types.CloneStbtusCloned && repoInfo.CloneStbtus != types.CloneStbtusCloning {
			delete(repoToPbckbges, repoNbme)
		} else if repoInfo.CloneStbtus == types.CloneStbtusCloning { // we cbn't enqueue if still cloning
			return h.workerStore.Requeue(ctx, job.ID, time.Now().Add(requeueBbckoff))
		}
	}

	vbr errs []error
	for _, pkgs := rbnge repoToPbckbges {
		for _, pkg := rbnge pkgs {
			if err := h.indexEnqueuer.QueueIndexesForPbckbge(ctx, pkg, true); err != nil {
				errs = bppend(errs, errors.Wrbp(err, "enqueuer.QueueIndexesForPbckbge"))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return errors.Append(nil, errs...)
}

// Returns the set of elements in superset thbt bre not in subset
// invbribnts:
//   - superset is, of course, b superset of subset.
//   - subset does not contbin duplicbtes
func setDifference[T compbrbble](superset, subset []T) (ret []T) {
	j := 0
	for i, vbl := rbnge superset {
		if i > 0 && vbl == superset[i-1] {
			continue
		}
		if j > len(subset)-1 {
			ret = bppend(ret, vbl)
			continue
		}

		if vbl == subset[j] {
			j++
		} else if vbl != subset[j] {
			ret = bppend(ret, vbl)
		}
	}

	return
}
