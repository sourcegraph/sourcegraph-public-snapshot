package shared

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
)

// Main is the single-program command function, which is shared between Sourcegraph's open-source
// and enterprise variant.
func Main() {
	flag.Parse()
	log.SetFlags(0)

	if flag.NArg() == 0 {
		log.Fatal(RunAll(context.Background()))
	} else {
		log.Fatal(RunOne(flag.Arg(0)))
	}
}

func RunAll(ctx context.Context) error {
	for name, program := range programs {
		// TODO(sqs): parallelize
		cmd := exec.CommandContext(ctx, "sourcegraph", name)
		cmd.Env = append(os.Environ(), program.env...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

func RunOne(name string) error {
	program, ok := programs[name]
	if !ok {
		return fmt.Errorf("no program: %s", name)
	}
	return program.main()
}

type program struct {
	env  []string
	main func() error
}

var programs = map[string]program{
	"frontend": {
		env: []string{
			"CONFIGURATION_MODE=server",
			"SRC_HTTP_ADDR=:7077",
			"SRC_HTTP_ADDR_INTERNAL=:7078",
			"SRC_FRONTEND_INTERNAL=localhost:7078",
			"SRC_LOG_LEVEL=info",
			"REDIS_ENDPOINT=127.0.0.1:6379",
		},
		main: func() error {
			// Set dummy authz provider to unblock channel for checking permissions in GraphQL APIs.
			// See https://github.com/sourcegraph/sourcegraph/issues/3847 for details.
			authz.SetProviders(true, []authz.Provider{})

			frontend_shared.Main(func() enterprise.Services {
				return enterprise.DefaultServices()
			})
			return nil
		},
	},
}
