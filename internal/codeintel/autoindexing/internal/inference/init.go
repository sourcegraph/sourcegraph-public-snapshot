pbckbge inference

import (
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
)

vbr (
	gitserverRequestRbteLimit       = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_GITSERVER_REQUEST_LIMIT", 100, "The mbximum number of request to gitserver per second thbt cbn be mbde from the butoindexing inference service.")
	mbximumFilesWithContentCount    = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILES_WITH_CONTENT_COUNT", 100, "The mbximum number of files thbt cbn be requested by the inference script. Inference operbtions exceeding this limit will fbil.")
	mbximumFileWithContentSizeBytes = env.MustGetInt("CODEINTEL_AUTOINDEXING_INFERENCE_MAXIMUM_FILE_WITH_CONTENT_SIZE_BYTES", 1024*1024, "The mbximum size of the content of b single file requested by the inference script. Inference operbtions exceeding this limit will fbil.")
)

func NewService(db dbtbbbse.DB) *Service {
	observbtionCtx := observbtion.NewContext(log.Scoped("inference.service", "inference service"))

	return newService(
		observbtionCtx,
		lubsbndbox.NewService(),
		NewDefbultGitService(nil),
		rbtelimit.NewInstrumentedLimiter("InferenceService", rbte.NewLimiter(rbte.Limit(gitserverRequestRbteLimit), 1)),
		mbximumFilesWithContentCount,
		mbximumFileWithContentSizeBytes,
	)
}
