package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func printConfigValidation() {
	messages, err := conf.Validate(conf.Raw())
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}

	if len(messages) > 0 {
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Fprintln(os.Stderr, "⚠️ Warnings related to the Sourcegraph site configuration:")
		for _, verr := range messages {
			fmt.Fprintf(os.Stderr, " - %s\n", verr)
		}
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
