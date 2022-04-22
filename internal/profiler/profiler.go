package profiler

import (
	"os"

	"cloud.google.com/go/profiler"
	ddprofiler "gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// Init starts the Google Cloud Profiler when in sourcegraph.com mode in
// production.  https://cloud.google.com/profiler/docs/profiling-go
func Init() {
	if !envvar.SourcegraphDotComMode() {
		return
	}

	// SourcegraphDotComMode can be true in dev, so check we are in a k8s
	// cluster.
	if !deploy.IsDeployTypeKubernetes(deploy.Type()) {
		return
	}

	// https://docs.datadoghq.com/tracing/profiler/enabling/go/
	if os.Getenv("DD_ENV") != "" {
		profileTypes := []ddprofiler.ProfileType{ddprofiler.CPUProfile, ddprofiler.HeapProfile}
		if os.Getenv("DD_PROFILE_ALL") != "" {
			profileTypes = append(profileTypes, ddprofiler.MutexProfile, ddprofiler.BlockProfile)
		}
		err := ddprofiler.Start(
			ddprofiler.WithService(env.MyName),
			ddprofiler.WithVersion(version.Version()),
			ddprofiler.WithProfileTypes(profileTypes...,
			),
		)
		if err != nil {
			log15.Error("profiler.Init datadog", "error", err)
		}
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
