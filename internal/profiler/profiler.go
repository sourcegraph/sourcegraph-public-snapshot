package profiler

import (
	"cloud.google.com/go/profiler"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// Init starts the Google Cloud Profiler when in sourcegraph.com mode in
// production.  https://cloud.google.com/profiler/docs/profiling-go
func Init() error {
	if !envvar.SourcegraphDotComMode() {
		return nil
	}

	// SourcegraphDotComMode can be true in dev, so check we are in a k8s
	// cluster.
	if !deploy.IsDeployTypeKubernetes(deploy.Type()) {
		return nil
	}

	return profiler.Start(profiler.Config{
		Service:        env.MyName,
		ServiceVersion: version.Version(),
		MutexProfiling: true,
		AllocForceGC:   true,
	})
}
