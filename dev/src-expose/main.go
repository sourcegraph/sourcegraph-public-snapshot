// Command "src-expose" serves directories as git repositories over HTTP.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type usageError struct {
	Message string
	FlagSet *flag.FlagSet
}

func (e *usageError) Usage() {
	e.FlagSet.Usage()
}

func (e *usageError) Error() string {
	return e.Message
}

func dockerAddr(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		port = "3434"
	}
	return "host.docker.internal:" + port
}

func main() {
	log.SetPrefix("")

	var (
		globalFlags    = flag.NewFlagSet("src-expose", flag.ExitOnError)
		globalVerbose  = globalFlags.Bool("verbose", false, "")
		globalBefore   = globalFlags.String("before", "", "A command to run before sync. It is run from the current working directory.")
		globalReposDir = globalFlags.String("repos-dir", "", "src-expose's git directories. src-expose creates a git repo per directory synced. The git repo is then served to Sourcegraph. The repositories are stored and served relative to this directory. Default: ~/.sourcegraph/src-expose-repos")
		globalConfig   = globalFlags.String("config", "", "If set will be used instead of command line arguments to specify configuration.")

		syncFlags = flag.NewFlagSet("sync", flag.ExitOnError)

		serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)
		serveAddr  = serveFlags.String("addr", "127.0.0.1:3434", "address on which to serve (end with : for unused port)")
	)

	parseSnapshotter := func(flagSet *flag.FlagSet, args []string) (*Snapshotter, error) {
		var s Snapshotter
		if *globalConfig != "" {
			if len(args) != 0 {
				return nil, &usageError{
					Message: "does not take arguments if -config is specified",
					FlagSet: flagSet,
				}
			}
			b, err := ioutil.ReadFile(*globalConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "could read configuration at %s", *globalConfig)
			}
			if err := yaml.Unmarshal(b, &s); err != nil {
				return nil, errors.Wrapf(err, "could not parse configuration at %s", *globalConfig)
			}
		} else {
			if len(args) == 0 {
				return nil, &usageError{
					Message: "requires atleast 1 argument, or -config to be specified.",
					FlagSet: flagSet,
				}
			}
			for _, dir := range args {
				s.Dirs = append(s.Dirs, &SyncDir{Dir: dir})
			}
		}
		if s.Destination == "" {
			s.Destination = *globalReposDir
		}
		if *globalBefore != "" {
			s.Before = *globalBefore
		}

		if err := s.SetDefaults(); err != nil {
			return nil, err
		}

		return &s, nil
	}

	serve := &ffcli.Command{
		Name:      "serve",
		Usage:     "src-expose [flags] serve [flags] [path/to/dir/containing/git/dirs]",
		ShortHelp: "Serve git repos for Sourcegraph to list and clone.",
		LongHelp: `src-expose serve will serve the git repositories over HTTP. These can be git
cloned, and they can be discovered by Sourcegraph.

src-expose will default to serving ~/.sourcegraph/src-expose-repos`,
		FlagSet: serveFlags,
		Exec: func(args []string) error {
			var repoDir string
			switch len(args) {
			case 0:
				repoDir = *globalReposDir

			case 1:
				repoDir = args[0]

			default:
				return &usageError{
					Message: "too many arguments",
					FlagSet: serveFlags,
				}
			}

			return serveRepos(*serveAddr, repoDir)
		},
	}

	sync := &ffcli.Command{
		Name:      "sync",
		Usage:     "src-expose [flags] sync [flags] <src1> [<src2> ...]",
		ShortHelp: "Do a one-shot sync of directories",
		FlagSet:   syncFlags,
		Exec: func(args []string) error {
			s, err := parseSnapshotter(syncFlags, args)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}

	root := &ffcli.Command{
		Name:  "src-expose",
		Usage: "src-expose [flags] <src1> [<src2> ...]",
		LongHelp: `Periodically sync directories src1, src2, ... and serve them.

For more advanced uses specify -config pointing to a yaml file.
See https://github.com/sourcegraph/sourcegraph/tree/master/dev/src-expose/examples`,
		Subcommands: []*ffcli.Command{serve, sync},
		FlagSet:     globalFlags,
		Exec: func(args []string) error {
			s, err := parseSnapshotter(globalFlags, args)
			if err != nil {
				return err
			}

			if *globalVerbose {
				b, _ := yaml.Marshal(s)
				_, _ = os.Stdout.Write(b)
				fmt.Println()
			}

			fmt.Printf(`Periodically syncing directories as git repositories to %s.
- %s
Serving the repositories at http://%s.
Paste the following configuration as an Other External Service in Sourcegraph:

  {
    "url": "http://%s", // Use http://%s if Sourcegraph is running in Docker
    "repos": ["hack-ignore-me"],
    "experimental.srcExpose": true
  }

`, *globalReposDir, strings.Join(args[1:], "\n- "), *serveAddr, *serveAddr, dockerAddr(*serveAddr))

			go func() {
				if err := serveRepos(*serveAddr, *globalReposDir); err != nil {
					log.Fatal(err)
				}
			}()

			for {
				if err := s.Run(); err != nil {
					return err
				}
				time.Sleep(s.Duration)
			}
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		if u, ok := err.(interface{ Usage() }); ok {
			u.Usage()
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
