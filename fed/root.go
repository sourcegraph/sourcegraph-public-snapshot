package fed

import (
	"log"

	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// Config is DEPRECATED.
//
// TODO(sqs): remove this after the next release when the deployment
// config is updated.
var Config struct {
	DeprecatedRootURLStr string `long:"fed.root-url" hidden:"yes"`
	DeprecatedIsRoot     bool   `long:"fed.is-root" hidden:"yes"`
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		if _, err := cli.Serve.AddGroup("Federation", "Federation", &Config); err != nil {
			log.Fatal(err)
		}
	})

	cli.ServeInit = append(cli.ServeInit, func() {
		if Config.DeprecatedIsRoot {
			log15.Warn("The --fed.is-root option is DEPRECATED. Remove it; it is a no-op.")
		}
		if Config.DeprecatedRootURLStr != "" {
			log15.Warn("The --fed.root-url option is DEPRECATED. Remove it; it is a no-op.")
		}
	})
}
