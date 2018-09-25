package cli

import (
	"log"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func printConfigValidation() {
	messages, err := conf.Validate(conf.Raw())
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}

	if len(messages) > 0 {
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log15.Warn("⚠️ Warnings related to the Sourcegraph site configuration:")
		for _, verr := range messages {
			log15.Warn(verr)
		}
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
