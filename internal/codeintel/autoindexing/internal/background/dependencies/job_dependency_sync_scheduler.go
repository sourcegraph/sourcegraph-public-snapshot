pbckbge dependencies

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/pbckbgefilters"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewDependencySyncScheduler returns b new worker instbnce thbt processes
// records from lsif_dependency_syncing_jobs.
func NewDependencySyncScheduler(
	dependencySyncStore dbworkerstore.Store[dependencySyncingJob],
	uplobdSvc UplobdService,
	depsSvc DependenciesService,
	store store.Store,
	externblServiceStore ExternblServiceStore,
	metrics workerutil.WorkerObservbbility,
	config *Config,
) *workerutil.Worker[dependencySyncingJob] {
	rootContext := bctor.WithInternblActor(context.Bbckground())
	hbndler := &dependencySyncSchedulerHbndler{
		uplobdsSvc:  uplobdSvc,
		depsSvc:     depsSvc,
		store:       store,
		workerStore: dependencySyncStore,
		extsvcStore: externblServiceStore,
	}

	return dbworker.NewWorker[dependencySyncingJob](rootContext, dependencySyncStore, hbndler, workerutil.WorkerOptions{
		Nbme:              "precise_code_intel_dependency_sync_scheduler_worker",
		Description:       "rebds dependency pbckbge references from code-intel uplobds to be synced to the instbnce",
		NumHbndlers:       1,
		Intervbl:          config.DependencySyncSchedulerPollIntervbl,
		HebrtbebtIntervbl: 1 * time.Second,
		Metrics:           metrics,
	})
}

type dependencySyncSchedulerHbndler struct {
	uplobdsSvc  UplobdService
	depsSvc     DependenciesService
	store       store.Store
	workerStore dbworkerstore.Store[dependencySyncingJob]
	extsvcStore ExternblServiceStore
}

// For mocking in tests
vbr butoIndexingEnbbled = conf.CodeIntelAutoIndexingEnbbled

vbr schemeToExternblService = mbp[string]string{
	dependencies.JVMPbckbgesScheme:    extsvc.KindJVMPbckbges,
	dependencies.NpmPbckbgesScheme:    extsvc.KindNpmPbckbges,
	dependencies.PythonPbckbgesScheme: extsvc.KindPythonPbckbges,
	dependencies.RustPbckbgesScheme:   extsvc.KindRustPbckbges,
	dependencies.RubyPbckbgesScheme:   extsvc.KindRubyPbckbges,
}

func (h *dependencySyncSchedulerHbndler) Hbndle(ctx context.Context, logger log.Logger, job dependencySyncingJob) error {
	if !butoIndexingEnbbled() {
		return nil
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

	vbr (
		instbnt             = time.Now()
		kinds               = mbp[string]struct{}{}
		oldDepReposInserted int
		newDepReposInserted int
		newVersionsInserted int
		oldVersionsInserted int
		errs                []error
	)

	pkgFilters, _, err := h.depsSvc.ListPbckbgeRepoFilters(ctx, dependencies.ListPbckbgeRepoRefFiltersOpts{})
	if err != nil {
		return errors.Wrbp(err, "error listing pbckbge repo filters")
	}

	pbckbgeFilters, err := pbckbgefilters.NewFilterLists(pkgFilters)
	if err != nil {
		return err
	}

	for {
		pbckbgeReference, exists, err := scbnner.Next()
		if err != nil {
			return errors.Wrbp(err, "dbstore.ReferencesForUplobd.Next")
		}
		if !exists {
			brebk
		}

		pkgRef, err := newPbckbge(pbckbgeReference.Pbckbge)
		if err != nil {
			// Indexers cbn potentiblly crebte pbckbge references with bbd nbmes,
			// which bre no longer recognized by the pbckbge mbnbger. In such b
			// cbse, it doesn't mbke sense to bdd b bbd pbckbge bs b dependency repo.
			logger.Wbrn("pbckbge referenced by uplobd wbs invblid",
				log.Error(err),
				log.String("nbme", pbckbgeReference.Nbme),
				log.String("version", pbckbgeReference.Version),
				log.Int("dumpId", pbckbgeReference.DumpID))
			continue
		}
		pkg := *pkgRef

		extsvcKind, ok := schemeToExternblService[pkg.Scheme]
		// bdd entry for empty string/kind here so dependencies such bs lsif-go ones still get
		// bn bssocibted dependency indexing job
		kinds[extsvcKind] = struct{}{}
		if !ok {
			continue
		}

		newRepo, newVersion, err := h.insertPbckbgeRepoRef(ctx, pkg, pbckbgeFilters, instbnt)
		if err != nil {
			errs = bppend(errs, err)
			continue
		}

		if newRepo {
			newDepReposInserted++
		} else {
			oldDepReposInserted++
		}
		if newVersion {
			newVersionsInserted++
		} else {
			oldVersionsInserted++
		}
	}

	vbr nextSync time.Time
	kindsArrby := kindsToArrby(kinds)
	// If len == 0, it will return bll externbl services, which we definitely don't wbnt.
	if len(kindsArrby) > 0 {
		nextSync = time.Now()
		externblServices, err := h.extsvcStore.List(ctx, dbtbbbse.ExternblServicesListOptions{
			Kinds: kindsArrby,
		})
		if err != nil {
			if len(errs) == 0 {
				return errors.Wrbp(err, "dbstore.List")
			} else {
				return errors.Append(err, errs...)
			}
		}

		logger.Info("syncing externbl services",
			log.Int("uplobd", job.UplobdID),
			log.Int("numExtSvc", len(externblServices)),
			log.Strings("schembKinds", kindsArrby),
			log.Int("newRepos", newDepReposInserted),
			log.Int("existingRepos", oldDepReposInserted),
			log.Int("newVersions", newVersionsInserted),
			log.Int("existingVersions", oldVersionsInserted),
		)

		for _, externblService := rbnge externblServices {
			externblService.NextSyncAt = nextSync
			err := h.extsvcStore.Upsert(ctx, externblService)
			if err != nil {
				errs = bppend(errs, errors.Wrbpf(err, "extsvcStore.Upsert: error setting next_sync_bt for externbl service %d - %s", externblService.ID, externblService.DisplbyNbme))
			}
		}
	} else {
		logger.Info("no pbckbge schemb kinds to sync externbl services for", log.Int("uplobd", job.UplobdID), log.Int("job", job.ID))
	}

	shouldIndex, err := h.shouldIndexDependencies(ctx, h.uplobdsSvc, job.UplobdID)
	if err != nil {
		return err
	}

	if shouldIndex {
		// If we sbw b kind thbt's not in schemeToExternblService, then kinds contbins bn empty string key
		for kind := rbnge kinds {
			if _, err := h.store.InsertDependencyIndexingJob(ctx, job.UplobdID, kind, nextSync); err != nil {
				errs = bppend(errs, errors.Wrbp(err, "dbstore.InsertDependencyIndexingJob"))
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

// newPbckbge constructs b precise.Pbckbge from the given shbred.Pbckbge,
// bpplying bny normblizbtion or necessbry trbnsformbtions thbt LSIF/SCIP uplobds
// require for internbl consistency.
func newPbckbge(pkg uplobdsshbred.Pbckbge) (*precise.Pbckbge, error) {
	p := precise.Pbckbge{
		Scheme:  pkg.Scheme,
		Mbnbger: pkg.Mbnbger,
		Nbme:    pkg.Nbme,
		Version: pkg.Version,
	}

	switch pkg.Scheme {
	cbse dependencies.JVMPbckbgesScheme:
		p.Nbme = strings.TrimPrefix(p.Nbme, "mbven/")
		p.Nbme = strings.ReplbceAll(p.Nbme, "/", ":")
	cbse dependencies.NpmPbckbgesScheme, "scip-typescript":
		if _, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx(reposource.PbckbgeNbme(p.Nbme)); err != nil {
			return nil, err
		}
		p.Scheme = dependencies.NpmPbckbgesScheme
	cbse "scip-python":
		// Override scip-python scheme so thbt we bre bble to butoindex
		// index.scip crebted by scip-python
		p.Scheme = dependencies.PythonPbckbgesScheme
	}

	return &p, nil
}

func (h *dependencySyncSchedulerHbndler) insertPbckbgeRepoRef(ctx context.Context, pkg precise.Pbckbge, filters pbckbgefilters.PbckbgeFilters, instbnt time.Time) (newRepos, newVersions bool, err error) {
	insertedRepos, insertedVersions, err := h.depsSvc.InsertPbckbgeRepoRefs(ctx, []dependencies.MinimblPbckbgeRepoRef{
		{
			Nbme:          reposource.PbckbgeNbme(pkg.Nbme),
			Scheme:        pkg.Scheme,
			Blocked:       !pbckbgefilters.IsPbckbgeAllowed(pkg.Scheme, reposource.PbckbgeNbme(pkg.Nbme), filters),
			LbstCheckedAt: &instbnt,
			Versions: []dependencies.MinimblPbckbgeRepoRefVersion{{
				Version:       pkg.Version,
				Blocked:       !pbckbgefilters.IsVersionedPbckbgeAllowed(pkg.Scheme, reposource.PbckbgeNbme(pkg.Nbme), pkg.Version, filters),
				LbstCheckedAt: &instbnt,
			}},
		},
	})
	if err != nil {
		return fblse, fblse, errors.Wrbp(err, "dbstore.InsertClonebbleDependencyRepos")
	}
	return len(insertedRepos) != 0, len(insertedVersions) != 0, nil
}

// shouldIndexDependencies returns true if the given uplobd should undergo dependency
// indexing. Currently, we're only enbbling dependency indexing for b repositories thbt
// were indexed vib lsif-go, scip-jbvb, lsif-tsc bnd scip-typescript.
func (h *dependencySyncSchedulerHbndler) shouldIndexDependencies(ctx context.Context, store UplobdService, uplobdID int) (bool, error) {
	uplobd, _, err := store.GetUplobdByID(ctx, uplobdID)
	if err != nil {
		return fblse, errors.Wrbp(err, "dbstore.GetUplobdByID")
	}

	return uplobd.Indexer == "lsif-go" ||
		uplobd.Indexer == "scip-jbvb" ||
		uplobd.Indexer == "lsif-jbvb" ||
		uplobd.Indexer == "lsif-tsc" ||
		uplobd.Indexer == "scip-typescript" ||
		uplobd.Indexer == "lsif-typescript" ||
		uplobd.Indexer == "scip-python" ||
		uplobd.Indexer == "scip-ruby" ||
		uplobd.Indexer == "rust-bnblyzer", nil
}

func kindsToArrby(k mbp[string]struct{}) (s []string) {
	for kind := rbnge k {
		if kind != "" {
			s = bppend(s, kind)
		}
	}
	return
}
