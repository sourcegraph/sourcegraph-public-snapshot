package runtime

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

// ConfigLoader is implemented by custom service configuration.
type ConfigLoader[ConfigT any] interface {
	*ConfigT

	// Load should populate ConfigT with values from env. Errors should be added
	// to env using env.AddError().
	Load(env *Env)
}

func renderHelp(service contract.ServiceMetadataProvider, env *Env) {
	fmt.Printf("SERVICE: %s\nVERSION: %s\n",
		service.Name(), service.Version())
	fmt.Printf("CONFIGURATION OPTIONS:\n")
	for _, v := range env.GetRequestedEnvVars() {
		fmt.Printf("- '%s': %s", v.Name, v.Description)
		if v.DefaultValue != "" {
			fmt.Printf(" (default: %q)", v.DefaultValue)
		} else {
			fmt.Printf(" (required)")
		}
		fmt.Println()
	}
}
