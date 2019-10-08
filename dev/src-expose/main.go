// Command "src-expose" serves directories as git repositories over HTTP.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
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

	var defaultSnapshotDir string
	if h, err := os.UserHomeDir(); err != nil {
		log.Fatal(err)
	} else {
		defaultSnapshotDir = filepath.Join(h, ".sourcegraph", "snapshots")
	}

	var (
		globalFlags          = flag.NewFlagSet("src-expose", flag.ExitOnError)
		globalVerbose        = globalFlags.Bool("verbose", false, "")
		globalSnapshotDir    = globalFlags.String("snapshot-dir", defaultSnapshotDir, "Git snapshot directory. Snapshots are stored relative to this directory. The snapshots are served from this directory.")
		globalSnapshotConfig = globalFlags.String("snapshot-config", "", "If set will be used instead of command line arguments to specify snapshot configuration.")

		snapshotFlags = flag.NewFlagSet("snapshot", flag.ExitOnError)

		serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)
		serveAddr  = serveFlags.String("addr", "127.0.0.1:3434", "address on which to serve (end with : for unused port)")
	)

	parseSnapshotter := func(flagSet *flag.FlagSet, args []string) (*Snapshotter, error) {
		var s Snapshotter
		if *globalSnapshotConfig != "" {
			if len(args) != 0 {
				return nil, &usageError{
					Message: "does not take arguments if -snapshot-config is specified",
					FlagSet: flagSet,
				}
			}
			b, err := ioutil.ReadFile(*globalSnapshotConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "could read configuration at %s", *globalSnapshotConfig)
			}
			if err := yaml.Unmarshal(b, &s); err != nil {
				return nil, errors.Wrapf(err, "could not parse configuration at %s", *globalSnapshotConfig)
			}
		} else {
			if len(args) == 0 {
				return nil, &usageError{
					Message: "requires atleast 1 argument, or -snapshot-config to be specified.",
					FlagSet: flagSet,
				}
			}
			for _, dir := range args {
				s.Snapshots = append(s.Snapshots, &Snapshot{Dir: dir})
			}
		}
		if s.Destination == "" {
			s.Destination = *globalSnapshotDir
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

src-expose will default to serving ~/.sourcegraph/snapshots`,
		FlagSet: serveFlags,
		Exec: func(args []string) error {
			var repoDir string
			switch len(args) {
			case 0:
				repoDir = *globalSnapshotDir

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

	snapshot := &ffcli.Command{
		Name:      "snapshot",
		Usage:     "src-expose [flags] snapshot [flags] <src1> [<src2> ...]",
		ShortHelp: "Create a Git snapshot of directories",
		FlagSet:   snapshotFlags,
		Exec: func(args []string) error {
			s, err := parseSnapshotter(snapshotFlags, args)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}

	root := &ffcli.Command{
		Name:  "src-expose",
		Usage: "src-expose [flags] <precommand> <src1> [<src2> ...]",
		LongHelp: `Periodically create snapshots of directories src1, src2, ... and serve them.

For more advanced uses specify -snapshot-config pointing to a yaml file.

EXAMPLE CONFIGURATION

` + MustAssetString("example.yaml"),
		Subcommands: []*ffcli.Command{serve, snapshot},
		FlagSet:     globalFlags,
		Exec: func(args []string) error {
			var err error
			var s *Snapshotter
			if len(args) == 0 {
				s, err = parseSnapshotter(globalFlags, args)
				if err != nil {
					return err
				}
			} else if len(args) < 2 {
				return &usageError{
					Message: "requires atleast 2 argument",
					FlagSet: globalFlags,
				}
			} else {
				preCommand := args[0]
				s, err = parseSnapshotter(globalFlags, args[1:])
				if err != nil {
					return err
				}
				s.PreCommand = preCommand
			}

			if *globalVerbose {
				b, _ := yaml.Marshal(s)
				_, _ = os.Stdout.Write(b)
				fmt.Println()
			}

			fmt.Printf(`Periodically snapshotting directories as git repositories to %s.
- %s
Serving the repositories at http://%s.
Paste the following configuration as an Other External Service in Sourcegraph:

  {
    "url": "http://%s", // Use http://%s if Sourcegraph is running in Docker
    "repos": ["src-expose"], // This may change in versions later than 3.9
  }

`, *globalSnapshotDir, strings.Join(args[1:], "\n- "), *serveAddr, *serveAddr, dockerAddr(*serveAddr))

			go func() {
				if err := serveRepos(*serveAddr, *globalSnapshotDir); err != nil {
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
