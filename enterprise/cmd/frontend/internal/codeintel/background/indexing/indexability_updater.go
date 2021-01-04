package indexing

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/inference"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

const MaxGitserverRequestsPerSecond = 20

type IndexabilityUpdater struct {
	dbStore             DBStore
	gitserverClient     GitserverClient
	operations          *operations
	minimumSearchCount  int
	minimumSearchRatio  float64
	minimumPreciseCount int
	enableIndexingCNCF  bool
	limiter             *rate.Limiter
}

var _ goroutine.Handler = &IndexabilityUpdater{}

func NewIndexabilityUpdater(
	dbStore DBStore,
	gitserverClient GitserverClient,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	interval time.Duration,
	operations *operations,
) goroutine.BackgroundRoutine {
	updater := &IndexabilityUpdater{
		dbStore:             dbStore,
		gitserverClient:     gitserverClient,
		operations:          operations,
		minimumSearchCount:  minimumSearchCount,
		minimumSearchRatio:  minimumSearchRatio,
		minimumPreciseCount: minimumPreciseCount,
		enableIndexingCNCF:  os.Getenv("DISABLE_CNCF") == "",
		limiter:             rate.NewLimiter(MaxGitserverRequestsPerSecond, 1),
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(context.Background(), interval, updater, operations.handleIndexabilityUpdater)
}

func (u *IndexabilityUpdater) Handle(ctx context.Context) error {
	start := time.Now().UTC()

	stats, err := u.dbStore.RepoUsageStatistics(ctx)
	if err != nil {
		return errors.Wrap(err, "store.RepoUsageStatistics")
	}

	if u.enableIndexingCNCF {
		stats = append(stats, u.cncfStats()...)
	}

	var queueErr error
	for _, stat := range stats {
		if err := u.queueRepository(ctx, stat); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			if queueErr != nil {
				queueErr = err
			} else {
				queueErr = multierror.Append(queueErr, err)
			}
		}
	}
	if queueErr != nil {
		return queueErr
	}

	// Anything we didn't update hasn't had any activity and didn't come back
	// from RepoUsageStatitsics. Ensure we don't retain the last usage count
	// for these repositories indefinitely.
	if err := u.dbStore.ResetIndexableRepositories(ctx, start); err != nil {
		return errors.Wrap(err, "store.ResetIndexableRepositories")
	}

	return nil
}

func (u *IndexabilityUpdater) HandleError(err error) {
	log15.Error("Failed to update index queue", "err", err)
}

func (u *IndexabilityUpdater) queueRepository(ctx context.Context, repoUsageStatistics store.RepoUsageStatistics) (err error) {
	// Enable tracing on the context and trace the operation
	ctx = ot.WithShouldTrace(ctx, true)

	ctx, traceLog, endObservation := u.operations.queueRepository.WithAndLogger(ctx, &err, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", repoUsageStatistics.RepositoryID),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := u.limiter.Wait(ctx); err != nil {
		return err
	}

	commit, err := u.gitserverClient.Head(ctx, repoUsageStatistics.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}
	traceLog(log.String("commit", commit))

	paths, err := u.gitserverClient.ListFiles(ctx, repoUsageStatistics.RepositoryID, commit, inference.Patterns)
	if err != nil {
		return errors.Wrap(err, "gitserver.ListFiles")
	}
	traceLog(log.Int("numPaths", len(paths)))

	matched := false
	for name, handler := range inference.Recognizers {
		matched = handler.CanIndex(paths)
		traceLog(log.Bool(fmt.Sprintf("%s.CanIndex", name), matched))

		if matched {
			break
		}
	}
	if !matched {
		return nil
	}

	// TODO(efritz) - also check repo size

	indexableRepository := store.UpdateableIndexableRepository{
		RepositoryID: repoUsageStatistics.RepositoryID,
		SearchCount:  &repoUsageStatistics.SearchCount,
		PreciseCount: &repoUsageStatistics.PreciseCount,
	}
	if err := u.dbStore.UpdateIndexableRepository(ctx, indexableRepository, time.Now().UTC()); err != nil {
		return errors.Wrap(err, "store.UpdateIndexableRepository")
	}

	log15.Debug("Updated indexable repository metadata", "repository_id", repoUsageStatistics.RepositoryID)
	return nil
}

// Below we enable indexing CNCF repositories automatically as a work around for
// not having finished implementation of RFC 201 before having an opportunity where
// it would be the perfect solution T_T
//
// Each of these repositories will be the first in the list to try to be indexed. We
// artificially create a repo stats object for each one that will trick the scheduler
// into thinking the repo is "hot" enough to index, regardless of its actual use.
//
// Below is a list of repository IDs marked with the name of the repo:
// see https://github.com/sourcegraph/deploy-sourcegraph-dot-com/blob/5d7dce1a56e799c6b8167ee58c2c68ac25c67ee1/base/frontend/sourcegraph-frontend.ConfigMap.yaml#L4957
//
// Follow-up issue: https://github.com/sourcegraph/sourcegraph/issues/14343

var cncfRepositoryIDs = []int{
	480,      // github.com/prometheus/prometheus
	30214,    // github.com/helm/helm
	45657,    // github.com/kubernetes/kubernetes
	45756,    // github.com/nats-io/nats-server
	50798,    // github.com/opentracing/opentracing-go
	50912,    // github.com/grpc/grpc
	54472,    // github.com/containernetworking/cni
	60368,    // github.com/fluent/fluentd
	82186,    // github.com/theupdateframework/tuf
	87511,    // github.com/fluxcd/flux
	204428,   // github.com/openebs/openebs
	459734,   // github.com/open-policy-agent/opa
	490140,   // github.com/rook/rook
	615667,   // github.com/kubevirt/kubevirt
	749288,   // github.com/coredns/coredns
	1107281,  // github.com/containerd/containerd
	1244892,  // github.com/OpenObservability/OpenMetrics
	1252983,  // github.com/in-toto/in-toto
	1382850,  // github.com/argoproj/argo
	1452554,  // github.com/envoyproxy/envoy
	1481540,  // github.com/jaegertracing/jaeger
	1513627,  // github.com/spiffe/spire
	35581042, // github.com/brigadecore/brigade
	35595017, // github.com/theupdateframework/notary
	35613504, // github.com/projectcontour/contour
	35654543, // github.com/thanos-io/thanos
	35683453, // github.com/dragonflyoss/Dragonfly
	35733223, // github.com/linkerd/linkerd2
	35736704, // github.com/virtual-kubelet/virtual-kubelet
	35965026, // github.com/vitessio/vitess
	36096026, // github.com/litmuschaos/litmus
	36168644, // github.com/operator-framework/operator-sdk
	36239375, // github.com/telepresenceio/telepresence
	36305039, // github.com/strimzi/strimzi-kafka-operator
	36583472, // github.com/tikv/tikv
	36589859, // github.com/goharbor/harbor
	36645706, // github.com/buildpacks/pack
	36664934, // github.com/etcd-io/etcd
	36683122, // github.com/dexidp/dex
	36708822, // github.com/cri-o/cri-o
	36764924, // github.com/cortexproject/cortex
	36827876, // github.com/falcosecurity/falco
	37069094, // github.com/kubeedge/kubeedge
	37249424, // github.com/cloud-custodian/cloud-custodian
	37252548, // github.com/crossplane/crossplane
	37519923, // github.com/networkservicemesh/networkservicemesh
	37612302, // github.com/kudobuilder/kudo
	37650592, // github.com/cni-genie/CNI-Genie
	37700634, // github.com/rancher/k3s
	37764298, // github.com/chubaofs/chubaofs
	37779257, // github.com/longhorn/longhorn
	38195917, // github.com/kedacore/keda
	38697647, // github.com/servicemeshinterface/smi-spec
	38766483, // github.com/volcano-sh/volcano
	39017834, // github.com/bfenetworks/bfe
	39194379, // github.com/kumahq/kuma
	39299029, // github.com/spiffe/spiffe
	39322847, // github.com/parallaxsecond/parsec
	39738895, // github.com/chaos-mesh/chaos-mesh
	39957286, // github.com/cloudevents/spec
	39966375, // github.com/open-telemetry/opentelemetry-specification
	39969558, // github.com/keptn/keptn
	40014856, // github.com/tremor-rs/tremor-runtime
	40243107, // github.com/artifacthub/hub
	40313715, // github.com/spotify/backstage
	41224626, // github.com/serverlessworkflow/specification
	41696876, // github.com/openservicemesh/osm
}

func (u *IndexabilityUpdater) cncfStats() (stats []store.RepoUsageStatistics) {
	max := u.minimumSearchCount
	if max < u.minimumPreciseCount {
		max = u.minimumPreciseCount
	}

	for _, repositoryID := range cncfRepositoryIDs {
		stats = append(stats, store.RepoUsageStatistics{
			RepositoryID: repositoryID,
			SearchCount:  max,
			PreciseCount: max,
		})
	}

	return stats
}
