// Command fakehub serves git repositories within some directory over HTTP,
// along with a pastable config for easier manual testing of sourcegraph.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/ffcli"
	"github.com/pkg/errors"
)

func main() {
	log.SetPrefix("")

	var (
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

fakehub will default to serving ~/.sourcegraph/snapshots
`,
		FlagSet: serveFlags,
		Exec: func(args []string) error {
			return serve(*serveN, *serveAddr, args)
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
			var s Snapshotter
			for _, dir := range args {
				s.Snapshots = append(s.Snapshots, Snapshot{Dir: dir})
			}
			return s.Run()
		},
	}

	root := &ffcli.Command{
		Name:        "fakehub",
		Subcommands: []*ffcli.Command{serve, snapshot},
		Exec: func(args []string) error {
			return errors.New("specify a subcommand")
		},
	}

	if err := root.Run(os.Args[1:]); err != nil {
		log.Fatalf("error: %v", err)
	}
}
