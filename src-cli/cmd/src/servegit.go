package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sourcegraph/src-cli/internal/cmderrors"
	"github.com/sourcegraph/src-cli/internal/servegit"
)

func init() {
	flagSet := flag.NewFlagSet("serve-git", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), `'src serve-git' serves your local git repositories over HTTP for Sourcegraph to pull.

USAGE
  src [-v] serve-git [-list] [-addr :3434] [path/to/dir]

By default 'src serve-git' will recursively serve your current directory on the address ':3434'.

'src serve-git -list' will not start up the server. Instead it will write to stdout a list of
repository names it would serve.

Documentation at https://docs.sourcegraph.com/admin/external_service/src_serve_git
`)
	}
	var (
		addrFlag = flagSet.String("addr", ":3434", "Address on which to serve (end with : for unused port)")
		listFlag = flagSet.Bool("list", false, "list found repository names")
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		var repoDir string
		switch args := flagSet.Args(); len(args) {
		case 0:
			repoDir, err = os.Getwd()
			if err != nil {
				return err
			}

		case 1:
			repoDir = args[0]

		default:
			return cmderrors.Usage("requires zero or one arguments")
		}

		dbug := log.New(io.Discard, "", log.LstdFlags)
		if *verbose {
			dbug = log.New(os.Stderr, "DBUG serve-git: ", log.LstdFlags)
		}

		s := &servegit.Serve{
			Addr:  *addrFlag,
			Root:  repoDir,
			Info:  log.New(os.Stderr, "serve-git: ", log.LstdFlags),
			Debug: dbug,
		}

		if *listFlag {
			repos, err := s.Repos()
			if err != nil {
				return err
			}
			for _, r := range repos {
				fmt.Println(r.Name)
			}
			return nil
		}

		return s.Start()
	}

	// Register the command.
	commands = append(commands, &command{
		aliases:   []string{"servegit"},
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
