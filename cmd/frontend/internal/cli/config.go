package cli

import (
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/conf"

	multierror "github.com/hashicorp/go-multierror"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

func printConfigValidation() {
	var errs *multierror.Error
	basicMessages, err := conf.ValidateBasic(globals.ConfigurationServerFrontendOnly.RawBasic())
	if err != nil {
		errs = multierror.Append(errs, errs)
	}

	coreMessages, err := conf.ValidateBasic(globals.ConfigurationServerFrontendOnly.RawCore())
	if err != nil {
		errs = multierror.Append(errs, errs)
	}

	if errs.ErrorOrNil() != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", errs.ErrorOrNil())
		return
	}

	var messages []string
	messages = append(messages, basicMessages...)
	messages = append(messages, coreMessages...)

	if len(messages) > 0 {
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log15.Warn("⚠️ Warnings related to the Sourcegraph site configuration:")
		for _, verr := range messages {
			log15.Warn(verr)
		}
		log15.Warn("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
