package runtime

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

// passSanityCheck exits with a code zero if the environment variable SANITY_CHECK equals
// to "true". See internal/sanitycheck.
func passSanityCheck(svc contract.ServiceMetadataProvider) {
	if os.Getenv("SANITY_CHECK") == "true" {
		// dump metadata to stdout
		if err := json.NewEncoder(os.Stdout).Encode(map[string]string{
			"name":    svc.Name(),
			"version": svc.Version(),
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Dump metadata: %s", err)
			os.Exit(1)
		}
		// report success in stderr
		fmt.Fprint(os.Stderr, "Sanity check passed, exiting without error")
		os.Exit(0)
	}
}
