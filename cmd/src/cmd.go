package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// command is a subcommand handler and its flag set.
type command struct {
	// flagSet is the flag set for the command.
	flagSet *flag.FlagSet

	// aliases for the command.
	aliases []string

	// handler is the function that is invoked to handle this command.
	handler func(args []string) error

	// flagSet.Usage function to invoke on e.g. -h flag. If nil, a default one
	// one is used.
	usageFunc func()
}

// matches tells if the given name matches this command or one of its aliases.
func (c *command) matches(name string) bool {
	if name == c.flagSet.Name() {
		return true
	}
	for _, alias := range c.aliases {
		if name == alias {
			return true
		}
	}
	return false
}

// commander represents a top-level command with subcommands.
type commander []*command

// run runs the command.
func (c commander) run(flagSet *flag.FlagSet, cmdName, usageText string, args []string) {
	// Parse flags.
	flagSet.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usageText)
	}
	if !flagSet.Parsed() {
		_ = flagSet.Parse(args)
	}

	// Print usage if the command is "help".
	if flagSet.Arg(0) == "help" || flagSet.NArg() == 0 {
		flagSet.Usage()
		os.Exit(0)
	}

	// Configure default usage funcs for commands.
	for _, cmd := range c {
		cmd := cmd
		if cmd.usageFunc != nil {
			cmd.flagSet.Usage = cmd.usageFunc
			continue
		}
		cmd.flagSet.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of '%s %s':\n", cmdName, cmd.flagSet.Name())
			cmd.flagSet.PrintDefaults()
		}
	}

	// Find the subcommand to execute.
	name := flagSet.Arg(0)
	for _, cmd := range c {
		if !cmd.matches(name) {
			continue
		}

		// Read global configuration now.
		var err error
		cfg, err = readConfig()
		if err != nil {
			log.Fatal("reading config: ", err)
		}

		// Parse subcommand flags.
		args := flagSet.Args()[1:]
		if err := cmd.flagSet.Parse(args); err != nil {
			panic(fmt.Sprintf("all registered commands should use flag.ExitOnError: error: %s", err))
		}

		// Execute the subcommand.
		if err := cmd.handler(flagSet.Args()[1:]); err != nil {
			if _, ok := err.(*usageError); ok {
				log.Printf("error: %s\n\n", err)
				cmd.flagSet.Usage()
				os.Exit(2)
			}
			if e, ok := err.(*exitCodeError); ok {
				if e.error != nil {
					log.Println(e.error)
				}
				os.Exit(e.exitCode)
			}
			log.Fatal(err)
		}
		os.Exit(0)
	}
	log.Printf("%s: unknown subcommand %q", cmdName, name)
	log.Fatalf("Run '%s help' for usage.", cmdName)
}

// usageError is an error type that subcommands can return in order to signal
// that a usage error has occurred.
type usageError struct {
	error
}

// exitCodeError is an error type that subcommands can return in order to
// specify the exact exit code.
type exitCodeError struct {
	error
	exitCode int
}

func (e *exitCodeError) Error() string {
	if e.error != nil {
		return fmt.Sprintf("%s (exit code: %d)", e.error, e.exitCode)
	}
	return fmt.Sprintf("exit code: %d", e.exitCode)
}

const (
	graphqlErrorsExitCode = 2
)

func didYouMeanOtherCommand(actual string, suggested []string) *command {
	fullSuggestions := make([]string, len(suggested))
	for i, s := range suggested {
		fullSuggestions[i] = "src " + s
	}
	msg := fmt.Sprintf("src: unknown subcommand %q\n\nDid you mean:\n\n\t%s", actual, strings.Join(fullSuggestions, "\n\t"))
	return &command{
		flagSet:   flag.NewFlagSet(actual, flag.ExitOnError),
		handler:   func(args []string) error { return errors.New(msg) },
		usageFunc: func() { log.Println(msg) },
	}
}
