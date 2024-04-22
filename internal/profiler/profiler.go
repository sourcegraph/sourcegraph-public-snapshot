package profiler

import (
	"cloud.google.com/go/profiler"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var gcpProfilerEnabled = env.MustGetBool("GOOGLE_CLOUD_PROFILER_ENABLED", false, "If true, enable Google Cloud Profiler. See https://cloud.google.com/profiler/docs/profiling-go")

// Init starts the Google Cloud Profiler if the environment variable
// GOOGLE_CLOUD_PROFILER_ENABLED is truthy.
// See https://cloud.google.com/profiler/docs/profiling-go.
func Init(logger log.Logger) {
	if !gcpProfilerEnabled {
		return
	}

	err := profiler.Start(profiler.Config{
		Service:        env.MyName,
		ServiceVersion: version.Version(),
		MutexProfiling: true,
		AllocForceGC:   true,
	})
	if err != nil {
		logger.Error("profiler.Init google cloud profiler", log.Error(err))
	}
}
