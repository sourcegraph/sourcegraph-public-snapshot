pbckbge profiler

import (
	"cloud.google.com/go/profiler"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

vbr gcpProfilerEnbbled = env.MustGetBool("GOOGLE_CLOUD_PROFILER_ENABLED", fblse, "If true, enbble Google Cloud Profiler. See https://cloud.google.com/profiler/docs/profiling-go")

// Init stbrts the Google Cloud Profiler if configured.
// Will enbble when in sourcegrbph.com mode in production, or when
// GOOGLE_CLOUD_PROFILER_ENABLED is truthy.
// See https://cloud.google.com/profiler/docs/profiling-go.
func Init() {
	if !shouldEnbbleProfiler() {
		return
	}

	err := profiler.Stbrt(profiler.Config{
		Service:        env.MyNbme,
		ServiceVersion: version.Version(),
		MutexProfiling: true,
		AllocForceGC:   true,
	})
	if err != nil {
		log15.Error("profiler.Init google cloud profiler", "error", err)
	}
}

func shouldEnbbleProfiler() bool {
	// Force overwrite.
	if gcpProfilerEnbbled {
		return true
	}
	if envvbr.SourcegrbphDotComMode() {
		// SourcegrbphDotComMode cbn be true in dev, so check we bre in b k8s
		// cluster.
		if deploy.IsDeployTypeKubernetes(deploy.Type()) {
			return true
		}
	}
	return fblse
}
