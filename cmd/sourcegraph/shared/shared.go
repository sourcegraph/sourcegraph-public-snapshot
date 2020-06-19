package shared

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	replacer_shared "github.com/sourcegraph/sourcegraph/cmd/replacer/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	symbols_shared "github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for name, program := range programs {
		// TODO(sqs): parallelize
		cmd := exec.CommandContext(ctx, "sourcegraph", name)
		cmd.Env = append(os.Environ(), program.env...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pdeathsig: syscall.SIGTERM,
		}
		go func(name string) {
			err := cmd.Run()
			log.Printf("%s: exited (%s)", name, err)
			cancel()
		}(name)
	}

	<-ctx.Done()
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
			"SRC_GIT_SERVERS=127.0.0.1:3178",
			"LOGO=t",
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
	"gitserver": {
		env: []string{
			"SRC_FRONTEND_INTERNAL=localhost:7078",
			"SRC_REPOS_DIR=/tmp/sourcegraph-repos",
		},
		main: func() error {
			gitserver_shared.Main()
			return nil
		},
	},
	"symbols": {
		env: []string{
			"SRC_FRONTEND_INTERNAL=localhost:7078",
			"LIBSQLITE3_PCRE=/usr/lib/sqlite3/pcre.so", // TODO(sqs)
			// TODO(sqs): also requires universal-ctags in $PATH
		},
		main: func() error {
			symbols_shared.Main()
			return nil
		},
	},
	"searcher": {
		env: []string{
			"SRC_FRONTEND_INTERNAL=localhost:7078",
		},
		main: func() error {
			searcher_shared.Main()
			return nil
		},
	},
	"replacer": {
		env: []string{
			"SRC_FRONTEND_INTERNAL=localhost:7078",
		},
		main: func() error {
			replacer_shared.Main()
			return nil
		},
	},
	"repo-updater": {
		env: []string{
			"SRC_FRONTEND_INTERNAL=localhost:7078",
		},
		main: func() error {
			repoupdater_shared.Main(nil) // TODO(sqs): missing enterprise
			return nil
		},
	},
	// TODO(sqs): no zoekt, query-runner
}
