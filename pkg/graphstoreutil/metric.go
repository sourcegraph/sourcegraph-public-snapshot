package graphstoreutil

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

// withMetrics wraps s such that each method is instrumented with Prometheus
// duration metrics
func withMetrics(name string, s store.MultiRepoStoreImporterIndexer) store.MultiRepoStoreImporterIndexer {
	return &instrumentedGraphStore{name, s}
}

var opDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "graphstore",
	Name:      "op_duration_seconds",
	Help:      "GraphStore operation latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, []string{"name", "method", "repo", "error"})

func init() {
	prometheus.MustRegister(opDuration)
}

type instrumentedGraphStore struct {
	name string
	s    store.MultiRepoStoreImporterIndexer
}

var _ store.MultiRepoStoreImporterIndexer = (*instrumentedGraphStore)(nil)

func (s *instrumentedGraphStore) Defs(df ...store.DefFilter) (d []*graph.Def, err error) {
	start := time.Now()
	defer func() { s.opObserve("Defs", "", start, err) }()
	return s.s.Defs(df...)
}

func (s *instrumentedGraphStore) Refs(rf ...store.RefFilter) (r []*graph.Ref, err error) {
	start := time.Now()
	defer func() { s.opObserve("Refs", "", start, err) }()
	return s.s.Refs(rf...)
}

func (s *instrumentedGraphStore) Units(uf ...store.UnitFilter) (u []*unit.SourceUnit, err error) {
	start := time.Now()
	defer func() { s.opObserve("Units", "", start, err) }()
	return s.s.Units(uf...)
}

func (s *instrumentedGraphStore) Versions(vf ...store.VersionFilter) (v []*store.Version, err error) {
	start := time.Now()
	defer func() { s.opObserve("Versions", "", start, err) }()
	return s.s.Versions(vf...)
}

func (s *instrumentedGraphStore) Repos(rf ...store.RepoFilter) (x []string, err error) {
	start := time.Now()
	defer func() { s.opObserve("Repos", "", start, err) }()
	return s.s.Repos(rf...)
}

func (s *instrumentedGraphStore) Import(repo, commitID string, unit *unit.SourceUnit, data graph.Output) (err error) {
	start := time.Now()
	defer func() { s.opObserve("Import", repo, start, err) }()
	return s.s.Import(repo, commitID, unit, data)
}

func (s *instrumentedGraphStore) CreateVersion(repo, commitID string) (err error) {
	start := time.Now()
	defer func() { s.opObserve("CreateVersion", repo, start, err) }()
	return s.s.CreateVersion(repo, commitID)
}

func (s *instrumentedGraphStore) Index(repo, commitID string) (err error) {
	start := time.Now()
	defer func() { s.opObserve("Index", repo, start, err) }()
	return s.s.Index(repo, commitID)
}

func (s *instrumentedGraphStore) opObserve(method, repo string, start time.Time, err error) {
	isError := "false"
	if err != nil {
		isError = "true"
	}
	if repo != "" {
		repo = repotrackutil.GetTrackedRepo(repo)
	}
	opDuration.WithLabelValues(s.name, method, repo, isError).Observe(time.Since(start).Seconds())
}
