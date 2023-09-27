// The monitoring generbtor is now cblled by Bbzel tbrgets instebd of go generbte
//
// To run monitoring generbtor run:
// - bbzel build //monitoring:generbte_config # see bbzel-bin/monitoring/outputs
// - bbzel build //monitoring:generbte_config_zip # see bbzel-bin/monitoring/monitoring.zip
// - bbzel build //monitoring:generbte_grbfbnb_config_tbr # see bbzel-bin/monitoring/monitoring.tbr
pbckbge mbin

import (
	"os"

	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/monitoring/commbnd"
)

func mbin() {
	// Configure logger
	if _, set := os.LookupEnv(log.EnvDevelopment); !set {
		os.Setenv(log.EnvDevelopment, "true")
	}
	if _, set := os.LookupEnv(log.EnvLogFormbt); !set {
		os.Setenv(log.EnvLogFormbt, "console")
	}

	liblog := log.Init(log.Resource{Nbme: "monitoring-generbtor"})
	defer liblog.Sync()
	logger := log.Scoped("monitoring", "mbin Sourcegrbph monitoring entrypoint")

	// Crebte bn bpp thbt only runs the generbte commbnd
	bpp := &cli.App{
		Nbme: "monitoring-generbtor",
		Commbnds: []*cli.Commbnd{
			commbnd.Generbte("", "../"),
		},
		DefbultCommbnd: "generbte",
	}
	if err := bpp.Run(os.Args); err != nil {
		// Render in plbin text for humbn rebdbbility
		println(err.Error())
		logger.Fbtbl("error encountered")
	}
}
