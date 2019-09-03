// Command fakehub serves git repositories within some directory over HTTP,
// along with a pastable config for easier manual testing of sourcegraph.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/pkg/errors"
)

func main() {
	log.SetPrefix("")

	var defaultSnapshotDir string
	if h, err := os.UserHomeDir(); err != nil {
		log.Fatal(err)
	} else {
		defaultSnapshotDir = filepath.Join(h, ".sourcegraph", "snapshots")
	}

	var (
		globalFlags       = flag.NewFlagSet("fakehub", flag.ExitOnError)
		globalSnapshotDir = globalFlags.String("snapshot-dir", defaultSnapshotDir, "Git snapshot directory. Snapshots are stored relative to this directory. The snapshots are served from this directory.")

		serveFlags = flag.NewFlagSet("serve", flag.ExitOnError)
		serveN     = serveFlags.Int("n", 1, "number of instances of each repo to make")
		serveAddr  = serveFlags.String("addr", "127.0.0.1:3434", "address on which to serve (end with : for unused port)")
	)

	serve := &ffcli.Command{
		Name:      "serve",
		Usage:     "fakehub [flags] serve [flags] [path/to/dir/containing/git/dirs]",
		ShortHelp: "Serve git repos for Sourcegraph to list and clone.",
		LongHelp: `fakehub will serve any number (controlled with -n) of copies of the repo over
HTTP at /repo/1/.git, /repo/2/.git etc. These can be git cloned, and they can
be used as test data for sourcegraph. The easiest way to get them into
sourcegraph is to visit the URL printed out on startup and paste the contents
into the text box for adding single repos in sourcegraph Site Admin.

fakehub will default to serving ~/.sourcegraph/snapshots`,
		FlagSet: serveFlags,
		Exec: func(args []string) error {
			var repoDir string
			switch len(args) {
			case 0:
				repoDir = *globalSnapshotDir

			case 1:
				repoDir = args[0]

			default:
				return errors.New("too many arguments")
			}

			return serveRepos(*serveN, *serveAddr, repoDir)
		},
	}

	snapshot := &ffcli.Command{
		Name:      "snapshot",
		Usage:     "fakehub [flags] snapshot [flags] <src1> [<src2> ...]",
		ShortHelp: "Create a git snapshot of directories",
		Exec: func(args []string) error {
			if len(args) == 0 {
				return errors.New("requires atleast 1 argument")
			}
			s := Snapshotter{
				Destination: *globalSnapshotDir,
			}
			for _, dir := range args {
				s.Snapshots = append(s.Snapshots, Snapshot{Dir: dir})
			}
			return s.Run()
		},
	}

	root := &ffcli.Command{
		Name:        "fakehub",
		Usage:       "fakehub [flags] <precommand> <src1> [<src2> ...]",
		ShortHelp:   "Periodically create snapshots of directories src1, src2, ... and serve them.",
		Subcommands: []*ffcli.Command{serve, snapshot},
		Exec: func(args []string) error {
			if len(args) < 2 {
				return errors.New("requires atleast 2 argument")
			}
			s := Snapshotter{
				Destination: *globalSnapshotDir,
				PreCommand:  args[0],
			}
			for _, dir := range args[1:] {
				s.Snapshots = append(s.Snapshots, Snapshot{Dir: dir})
			}

			fmt.Printf(`Periodically snapshotting directories as git repositories to %s.
- %s
Serving the repositories at http://%s.
Paste the following configuration as an Other External Service in Sourcegraph:

  {
    "url": "http://%s",
    "repos": ["hack-ignore-me"],
    "experimental.fakehub": true
  }

`, *globalSnapshotDir, strings.Join(args[1:], "\n- "), *serveAddr, *serveAddr)

			go func() {
				if err := serveRepos(*serveN, *serveAddr, *globalSnapshotDir); err != nil {
					log.Fatal(err)
				}
			}()

			for {
				if err := s.Run(); err != nil {
					log.Fatalf("error: %v", err)
				}
				time.Sleep(10 * time.Second)
			}
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}
