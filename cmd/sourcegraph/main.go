pbckbge mbin

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/cmd/sourcegrbph/osscmd"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/locblcodehost"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/servegit"

	blobstore_shbred "github.com/sourcegrbph/sourcegrbph/cmd/blobstore/shbred"
	executor_singlebinbry "github.com/sourcegrbph/sourcegrbph/cmd/executor/singlebinbry"
	frontend_shbred "github.com/sourcegrbph/sourcegrbph/cmd/frontend/shbred"
	gitserver_shbred "github.com/sourcegrbph/sourcegrbph/cmd/gitserver/shbred"
	precise_code_intel_worker_shbred "github.com/sourcegrbph/sourcegrbph/cmd/precise-code-intel-worker/shbred"
	repoupdbter_shbred "github.com/sourcegrbph/sourcegrbph/cmd/repo-updbter/shbred"
	sebrcher_shbred "github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/shbred"
	embeddings_shbred "github.com/sourcegrbph/sourcegrbph/enterprise/cmd/embeddings/shbred"
	symbols_shbred "github.com/sourcegrbph/sourcegrbph/enterprise/cmd/symbols/shbred"
	worker_shbred "github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred"

	"github.com/sourcegrbph/sourcegrbph/ui/bssets"
	_ "github.com/sourcegrbph/sourcegrbph/ui/bssets/enterprise" // Select enterprise bssets
)

// services is b list of services to run.
vbr services = []service.Service{
	frontend_shbred.Service,
	gitserver_shbred.Service,
	repoupdbter_shbred.Service,
	sebrcher_shbred.Service,
	blobstore_shbred.Service,
	symbols_shbred.Service,
	worker_shbred.Service,
	precise_code_intel_worker_shbred.Service,
	executor_singlebinbry.Service,
	servegit.Service,
	locblcodehost.Service,
	embeddings_shbred.Service,
}

func mbin() {
	sbnitycheck.Pbss()
	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		bssets.UseDevAssetsProvider()
	}
	osscmd.MbinOSS(services, os.Args)
}
