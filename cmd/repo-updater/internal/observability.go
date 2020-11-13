package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

// StoreMetrics encapsulates the Prometheus metrics of a Store.
type StoreMetrics struct {
	Transact              *metrics.OperationMetrics
	Done                  *metrics.OperationMetrics
	UpsertRepos           *metrics.OperationMetrics
	UpsertSources         *metrics.OperationMetrics
	ListExternalRepoSpecs *metrics.OperationMetrics
	SetClonedRepos        *metrics.OperationMetrics
	CountNotClonedRepos   *metrics.OperationMetrics
	CountUserAddedRepos   *metrics.OperationMetrics
	EnqueueSyncJobs       *metrics.OperationMetrics
}

// MustRegister registers all metrics in StoreMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (sm StoreMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := range []*metrics.OperationMetrics{
		sm.Transact,
		sm.Done,
		sm.ListExternalRepoSpecs,
		sm.UpsertRepos,
		sm.UpsertSources,
		sm.SetClonedRepos,
	} {
		r.MustRegister(om.Count)
		r.MustRegister(om.Duration)
		r.MustRegister(om.Errors)
	}
}

// NewStoreMetrics returns StoreMetrics that need to be registered
// in a Prometheus registry.
func NewStoreMetrics() StoreMetrics {
	return StoreMetrics{
		Transact: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_transact_duration_seconds",
				Help: "Time spent opening a transaction",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_transact_total",
				Help: "Total number of opened transactions",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_transact_errors_total",
				Help: "Total number of errors when opening a transaction",
			}, []string{}),
		},
		Done: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_done_duration_seconds",
				Help: "Time spent closing a transaction",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_done_total",
				Help: "Total number of closed transactions",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_done_errors_total",
				Help: "Total number of errors when closing a transaction",
			}, []string{}),
		},
		UpsertRepos: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_upsert_repos_duration_seconds",
				Help: "Time spent upserting repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_upsert_repos_total",
				Help: "Total number of upserted repositories",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_upsert_repos_errors_total",
				Help: "Total number of errors when upserting repos",
			}, []string{}),
		},
		UpsertSources: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_upsert_sources_duration_seconds",
				Help: "Time spent upserting sources",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_upsert_sources_total",
				Help: "Total number of upserted sources",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_upsert_sources_errors_total",
				Help: "Total number of errors when upserting sources",
			}, []string{}),
		},
		ListExternalRepoSpecs: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_list_external_repo_specs_duration_seconds",
				Help: "Time spent listing external repo specs",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_list_external_repo_specs_total",
				Help: "Total number of listed external repo specs",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_list_external_repo_specs_errors_total",
				Help: "Total number of errors when listing external repo specs",
			}, []string{}),
		},
		SetClonedRepos: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_set_cloned_repos_duration_seconds",
				Help: "Time spent setting cloned repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_set_cloned_repos_total",
				Help: "Total number of set cloned repos calls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_set_cloned_repos_errors_total",
				Help: "Total number of errors when setting cloned repos",
			}, []string{}),
		},
		CountNotClonedRepos: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_count_not_cloned_repos_duration_seconds",
				Help: "Time spent counting not-cloned repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_count_not_cloned_repos_total",
				Help: "Total number of count not-cloned repos calls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_count_not_cloned_repos_errors_total",
				Help: "Total number of errors when counting not-cloned repos",
			}, []string{}),
		},
		CountUserAddedRepos: &metrics.OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: "src_repoupdater_store_count_user_added_repos",
				Help: "Time spent counting the number of user added repos",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_count_user_added_repos_total",
				Help: "Total number of count user added repo calls",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "src_repoupdater_store_count_user_added_repos_errors_total",
				Help: "Total number of errors when counting user added repos",
			}, []string{}),
		},
	}
}
