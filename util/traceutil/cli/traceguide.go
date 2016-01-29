package cli

import (
	"os"

	"github.com/resonancelabs/go-pub/instrument"
	tg_client "github.com/resonancelabs/go-pub/instrument/client"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	if os.Getenv("SG_TRACEGUIDE_ACCESS_TOKEN") == "" {
		return
	}
	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		options := &tg_client.Options{
			AccessToken: os.Getenv("SG_TRACEGUIDE_ACCESS_TOKEN"),
		}
		if len(os.Getenv("SG_TRACEGUIDE_SERVICE_HOST")) > 0 {
			options.ServiceHost = os.Getenv("SG_TRACEGUIDE_SERVICE_HOST")
		}
		instrument.SetDefaultRuntime(tg_client.NewRuntime(options))
		instrument.Log("Initialized Traceguide runtime")
	})
}
