package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/servegit"
)

func init() {
	flagSet := flag.NewFlagSet("serve-git", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), `'src serve-git' serves your local git repositories over HTTP for Sourcegraph to pull.

USAGE
  src [-v] serve-git [-addr :3434] [path/to/dir]

By default 'src serve-git' will recursively serve your current directory on the address ':3434'.
`)
	}
	var (
		addrFlag = flagSet.String("addr", ":3434", "Address on which to serve (end with : for unused port)")
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
			return &usageError{errors.New("requires zero or one arguments")}
		}

		dbug := log.New(ioutil.Discard, "", log.LstdFlags)
		if *verbose {
			dbug = log.New(os.Stderr, "DBUG serve-git: ", log.LstdFlags)
		}

		s := &servegit.Serve{
			Addr:  *addrFlag,
			Root:  repoDir,
			Info:  log.New(os.Stderr, "serve-git: ", log.LstdFlags),
			Debug: dbug,
		}
		return s.Start()
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
