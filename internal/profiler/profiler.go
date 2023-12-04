package profiler

import (
	"cloud.google.com/go/profiler"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var gcpProfilerEnabled = env.MustGetBool("GOOGLE_CLOUD_PROFILER_ENABLED", false, "If true, enable Google Cloud Profiler. See https://cloud.google.com/profiler/docs/profiling-go")

// Init starts the Google Cloud Profiler if configured.
// Will enable when in sourcegraph.com mode in production, or when
// GOOGLE_CLOUD_PROFILER_ENABLED is truthy.
// See https://cloud.google.com/profiler/docs/profiling-go.
func Init() {
	if !shouldEnableProfiler() {
		return
	}

	err := profiler.Start(profiler.Config{
		Service:        env.MyName,
		ServiceVersion: version.Version(),
		MutexProfiling: true,
		AllocForceGC:   true,
	})
	if err != nil {
		log15.Error("profiler.Init google cloud profiler", "error", err)
	}
}

func shouldEnableProfiler() bool {
	// Force overwrite.
	if gcpProfilerEnabled {
		return true
	}
	if envvar.SourcegraphDotComMode() {
		// SourcegraphDotComMode can be true in dev, so check we are in a k8s
		// cluster.
		if deploy.IsDeployTypeKubernetes(deploy.Type()) {
			return true
		}
	}
	return false
}
