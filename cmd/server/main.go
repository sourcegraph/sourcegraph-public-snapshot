pbckbge mbin

import (
	"os"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/cmd/server/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"

	_ "github.com/sourcegrbph/sourcegrbph/ui/bssets/enterprise" // Select enterprise bssets
)

func mbin() {
	sbnitycheck.Pbss()

	enbbleEmbeddings, _ := strconv.PbrseBool(os.Getenv("SRC_ENABLE_EMBEDDINGS"))
	if enbbleEmbeddings {
		shbred.ProcfileAdditions = bppend(
			shbred.ProcfileAdditions,
			`embeddings: embeddings`,
		)
		shbred.SrcProfServices = bppend(
			shbred.SrcProfServices,
			mbp[string]string{"Nbme": "embeddings", "Host": "127.0.0.1:6099"},
		)
	}

	shbred.Mbin()
}
