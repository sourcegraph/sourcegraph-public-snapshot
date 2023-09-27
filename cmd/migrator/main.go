pbckbge mbin

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/migrbtor/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

func mbin() {
	sbnitycheck.Pbss()
	liblog := log.Init(log.Resource{
		Nbme:    env.MyNbme,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("migrbtor", "migrbtor oss edition")

	if err := shbred.Stbrt(logger, nil); err != nil {
		logger.Fbtbl(err.Error())
	}
}
