package inference

import (
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	svc     *Service
	svcOnce sync.Once
)

var (
	gitserverRequestRateLimit       = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_GITSERVER_REQUEST_LIMIT", 100, "The maximum number of request to gitserver per second that can be made from the autoindexing inference service.")
	maximumFilesWithContentCount    = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILES_WITH_CONTENT_COUNT", 100, "The maximum number of files that can be requested by the inference script. Inference operations exceeding this limit will fail.")
	maximumFileWithContentSizeBytes = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILE_WITH_CONTENT_SIZE_BYTES", 1024*1024, "The maximum size of the content of a single file requested by the inference script. Inference operations exceeding this limit will fail.")
)

func GetService(db database.DB) *Service {
	svcOnce.Do(func() {
		observationContext := &observation.Context{
			Logger:     log.Scoped("inference.service", "inference service"),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}

		svc = newService(
			luasandbox.GetService(),
			NewDefaultGitService(nil, db),
			rate.NewLimiter(rate.Limit(gitserverRequestRateLimit), 1),
			maximumFilesWithContentCount,
			maximumFileWithContentSizeBytes,
			observationContext,
		)
	})

	return svc
}
