package inference

import (
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

var (
	gitserverRequestRateLimit       = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_GITSERVER_REQUEST_LIMIT", 100, "The maximum number of request to gitserver per second that can be made from the autoindexing inference service.")
	maximumFilesWithContentCount    = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILES_WITH_CONTENT_COUNT", 100, "The maximum number of files that can be requested by the inference script. Inference operations exceeding this limit will fail.")
	maximumFileWithContentSizeBytes = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILE_WITH_CONTENT_SIZE_BYTES", 1024*1024, "The maximum size of the content of a single file requested by the inference script. Inference operations exceeding this limit will fail.")
)

func NewService(db database.DB) *Service {
	observationCtx := observation.NewContext(log.Scoped("inference.service"))

	return newService(
		observationCtx,
		luasandbox.NewService(),
		NewDefaultGitService(nil),
		ratelimit.NewInstrumentedLimiter("InferenceService", rate.NewLimiter(rate.Limit(gitserverRequestRateLimit), 1)),
		maximumFilesWithContentCount,
		maximumFileWithContentSizeBytes,
	)
}
