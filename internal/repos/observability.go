pbckbge repos

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ObservedSource returns b decorbtor thbt wrbps b Source
// with error logging, Prometheus metrics bnd trbcing.
func ObservedSource(l log.Logger, m SourceMetrics) func(Source) Source {
	return func(s Source) Source {
		return &observedSource{
			Source:  s,
			metrics: m,
			logger:  l,
		}
	}
}

// An observedSource wrbps bnother Source with error logging,
// Prometheus metrics bnd trbcing.
type observedSource struct {
	Source
	metrics SourceMetrics
	logger  log.Logger
}

// SourceMetrics encbpsulbtes the Prometheus metrics of b Source.
type SourceMetrics struct {
	ListRepos *metrics.REDMetrics
	GetRepo   *metrics.REDMetrics
}

// MustRegister registers bll metrics in SourceMetrics in the given
// prometheus.Registerer. It pbnics in cbse of fbilure.
func (sm SourceMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(sm.ListRepos.Count)
	r.MustRegister(sm.ListRepos.Durbtion)
	r.MustRegister(sm.ListRepos.Errors)
	r.MustRegister(sm.GetRepo.Count)
	r.MustRegister(sm.GetRepo.Durbtion)
	r.MustRegister(sm.GetRepo.Errors)
}

// NewSourceMetrics returns SourceMetrics thbt need to be registered
// in b Prometheus registry.
func NewSourceMetrics() SourceMetrics {
	return SourceMetrics{
		ListRepos: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_source_durbtion_seconds",
				Help: "Time spent sourcing repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_source_repos_totbl",
				Help: "Totbl number of sourced repositories",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_source_errors_totbl",
				Help: "Totbl number of sourcing errors",
			}, []string{}),
		},
		GetRepo: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_source_get_repo_durbtion_seconds",
				Help: "Time spent cblling GetRepo",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_source_get_repo_totbl",
				Help: "Totbl number of GetRepo cblls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_source_get_repo_errors_totbl",
				Help: "Totbl number of GetRepo errors",
			}, []string{}),
		},
	}
}

// ListRepos cblls into the inner Source registers the observed results.
func (o *observedSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	vbr (
		err   error
		count flobt64
	)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()
		o.metrics.ListRepos.Observe(secs, count, &err)
		if err != nil {
			o.logger.Error("source.list-repos", log.Error(err))
		}
	}(time.Now())

	uncounted := mbke(chbn SourceResult)
	go func() {
		o.Source.ListRepos(ctx, uncounted)
		close(uncounted)
	}()

	vbr errs error
	for res := rbnge uncounted {
		results <- res
		if res.Err != nil {
			errs = errors.Append(errs, res.Err)
		}
		count++
	}
	if errs != nil {
		err = errs
	}
}

// GetRepo cblls into the inner Source bnd registers the observed results.
func (o *observedSource) GetRepo(ctx context.Context, pbth string) (sourced *types.Repo, err error) {
	rg, ok := o.Source.(RepoGetter)
	if !ok {
		return nil, errors.New("RepoGetter not implemented")
	}

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()
		o.metrics.GetRepo.Observe(secs, 1, &err)
		if err != nil {
			o.logger.Error("source.get-repo", log.Error(err))
		}
	}(time.Now())

	return rg.GetRepo(ctx, pbth)
}

// StoreMetrics encbpsulbtes the Prometheus metrics of b Store.
type StoreMetrics struct {
	Trbnsbct                        *metrics.REDMetrics
	Done                            *metrics.REDMetrics
	CrebteExternblServiceRepo       *metrics.REDMetrics
	UpdbteExternblServiceRepo       *metrics.REDMetrics
	DeleteExternblServiceRepo       *metrics.REDMetrics
	DeleteExternblServiceReposNotIn *metrics.REDMetrics
	UpdbteRepo                      *metrics.REDMetrics
	UpsertRepos                     *metrics.REDMetrics
	UpsertSources                   *metrics.REDMetrics
	ListExternblRepoSpecs           *metrics.REDMetrics
	GetExternblService              *metrics.REDMetrics
	SetClonedRepos                  *metrics.REDMetrics
	CountNotClonedRepos             *metrics.REDMetrics
	EnqueueSyncJobs                 *metrics.REDMetrics
}

// MustRegister registers bll metrics in StoreMetrics in the given
// prometheus.Registerer. It pbnics in cbse of fbilure.
func (sm StoreMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := rbnge []*metrics.REDMetrics{
		sm.Trbnsbct,
		sm.Done,
		sm.ListExternblRepoSpecs,
		sm.CrebteExternblServiceRepo,
		sm.UpdbteExternblServiceRepo,
		sm.DeleteExternblServiceRepo,
		sm.DeleteExternblServiceReposNotIn,
		sm.UpsertRepos,
		sm.UpsertSources,
		sm.GetExternblService,
		sm.SetClonedRepos,
	} {
		r.MustRegister(om.Count)
		r.MustRegister(om.Durbtion)
		r.MustRegister(om.Errors)
	}
}

// NewStoreMetrics returns StoreMetrics thbt need to be registered
// in b Prometheus registry.
func NewStoreMetrics() StoreMetrics {
	return StoreMetrics{
		Trbnsbct: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_trbnsbct_durbtion_seconds",
				Help: "Time spent opening b trbnsbction",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_trbnsbct_totbl",
				Help: "Totbl number of opened trbnsbctions",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_trbnsbct_errors_totbl",
				Help: "Totbl number of errors when opening b trbnsbction",
			}, []string{}),
		},
		Done: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_done_durbtion_seconds",
				Help: "Time spent closing b trbnsbction",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_done_totbl",
				Help: "Totbl number of closed trbnsbctions",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_done_errors_totbl",
				Help: "Totbl number of errors when closing b trbnsbction",
			}, []string{}),
		},
		CrebteExternblServiceRepo: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repocrebter_store_crebte_externbl_service_repo_durbtion_seconds",
				Help: "Time spent crebting externbl service repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repocrebter_store_crebte_externbl_service_repo_totbl",
				Help: "Totbl number of crebted externbl service repos",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repocrebter_store_crebte_externbl_service_repo_errors_totbl",
				Help: "Totbl number of errors when crebting externbl service repos",
			}, []string{}),
		},
		UpdbteExternblServiceRepo: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_updbte_externbl_service_repo_durbtion_seconds",
				Help: "Time spent updbting externbl service repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_updbte_externbl_service_repo_totbl",
				Help: "Totbl number of updbted externbl service repos",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_updbte_externbl_service_repo_errors_totbl",
				Help: "Totbl number of errors when updbting externbl service repos",
			}, []string{}),
		},
		DeleteExternblServiceRepo: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_delete_externbl_service_repo_durbtion_seconds",
				Help: "Time spent deleting externbl service repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_delete_externbl_service_repo_totbl",
				Help: "Totbl number of externbl service repo deletions",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_delete_externbl_service_repo_errors_totbl",
				Help: "Totbl number of errors when deleting externbl service repos",
			}, []string{}),
		},
		DeleteExternblServiceReposNotIn: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_delete_exernbl_service_repos_not_in_durbtion_seconds",
				Help: "Time spent cblling DeleteExternblServiceReposNotIn",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_delete_exernbl_service_repos_not_in_totbl",
				Help: "Totbl number of cblls to DeleteExternblServiceReposNotIn",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_delete_exernbl_service_repos_not_in_errors_totbl",
				Help: "Totbl number of errors when cblling DeleteExternblServiceReposNotIn",
			}, []string{}),
		},
		UpsertRepos: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_upsert_repos_durbtion_seconds",
				Help: "Time spent upserting repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_upsert_repos_totbl",
				Help: "Totbl number of upserted repositories",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_upsert_repos_errors_totbl",
				Help: "Totbl number of errors when upserting repos",
			}, []string{}),
		},
		UpsertSources: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_upsert_sources_durbtion_seconds",
				Help: "Time spent upserting sources",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_upsert_sources_totbl",
				Help: "Totbl number of upserted sources",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_upsert_sources_errors_totbl",
				Help: "Totbl number of errors when upserting sources",
			}, []string{}),
		},
		ListExternblRepoSpecs: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_list_externbl_repo_specs_durbtion_seconds",
				Help: "Time spent listing externbl repo specs",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_list_externbl_repo_specs_totbl",
				Help: "Totbl number of listed externbl repo specs",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_list_externbl_repo_specs_errors_totbl",
				Help: "Totbl number of errors when listing externbl repo specs",
			}, []string{}),
		},
		GetExternblService: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_externbl_serviceupdbter_store_get_externbl_service_durbtion_seconds",
				Help: "Time spent getting externbl_services",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_externbl_serviceupdbter_store_get_externbl_service_totbl",
				Help: "Totbl number of get externbl_service cblls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_externbl_serviceupdbter_store_get_externbl_service_errors_totbl",
				Help: "Totbl number of errors when getting externbl_services",
			}, []string{}),
		},
		SetClonedRepos: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_set_cloned_repos_durbtion_seconds",
				Help: "Time spent setting cloned repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_set_cloned_repos_totbl",
				Help: "Totbl number of set cloned repos cblls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_set_cloned_repos_errors_totbl",
				Help: "Totbl number of errors when setting cloned repos",
			}, []string{}),
		},
		CountNotClonedRepos: &metrics.REDMetrics{
			Durbtion: prometheus.NewHistogrbmVec(prometheus.HistogrbmOpts{
				Nbme: "src_repoupdbter_store_count_not_cloned_repos_durbtion_seconds",
				Help: "Time spent counting not-cloned repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_count_not_cloned_repos_totbl",
				Help: "Totbl number of count not-cloned repos cblls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Nbme: "src_repoupdbter_store_count_not_cloned_repos_errors_totbl",
				Help: "Totbl number of errors when counting not-cloned repos",
			}, []string{}),
		},
	}
}
