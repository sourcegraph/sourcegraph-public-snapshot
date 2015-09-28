package sgx

import (
	"log"
	"os"

	"bytes"
	"strings"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
)

// globalOpt contains global options.
var globalOpt struct {
	Config     func(s string) error `long:"config" description:"INI config file" no-ini:"true"`
	Verbose    bool                 `short:"v" description:"show verbose output (same as --log-level=dbug)"`
	VerbosePkg string               `long:"verbose-pkg" description:"if set, only log output from specified package"`
	LogLevel   string               `long:"log-level" description:"upper log level to restrict log output to (dbug, info, warn, error, crit)" default:"info"`
}

func SetVerbose(v bool) {
	globalOpt.Verbose = v
	srclib.GlobalOpt.Verbose = v
}

func init() {
	cli.CLI.LongDescription = "src runs and manages a Sourcegraph instance."
	cli.CLI.AddGroup("Global options", "", &globalOpt)
}

func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")

	// Add our own custom help group to each command.
	if err := addHelpGroups(cli.CLI.Command); err != nil {
		return err
	}

	for _, f := range cli.PostInit {
		f()
	}

	globalOpt.Config = func(s string) error {
		return flags.NewIniParser(cli.CLI).ParseFile(s)
	}

	_, err := cli.CLI.Parse()
	printErrorHelp(err)
	return err
}

func init() {
	// Hack: treat command-line args after "sgx env exec" and "sgx
	// runas" as args to pass through to the program, not as flags
	// that we should parse. Use a heuristic to tell whether one of
	// these subexecing commands is probably being invoked.
	var isEnv, seenFirstCmd bool
	for i, a := range os.Args {
		if i == 0 {
			continue
		}
		if !seenFirstCmd && a == "env" {
			isEnv = true
		}
		if !strings.HasPrefix(a, "-") {
			seenFirstCmd = true
		}
		if ((isEnv && a == "exec") || a == "runas") && i+1 < len(os.Args) {
			// Insert a "--" after "exec" or "runas" in the
			// command-line args to stop flag parsing.
			os.Args = append(os.Args[:i+1], append([]string{"--"}, os.Args[i+1:]...)...)
			break
		}
	}
}

// addHelpGroups adds help groups to the given command and all of it's sub
// commands, recursively.
func addHelpGroups(cmd *flags.Command) error {
	// Determine whether or not we should register the default help group.
	register := true
	for _, name := range cli.CustomHelpCmds {
		if cmd.Name == name {
			register = false
			break
		}
	}

	if register {
		// Build the group.
		var help struct {
			ShowHelp func() error `short:"h" long:"help" description:"Show this help message"`
		}
		help.ShowHelp = func() error {
			var b bytes.Buffer
			cli.CLI.WriteHelp(&b)
			return &flags.Error{
				Type:    flags.ErrHelp,
				Message: b.String(),
			}
		}

		// Add the group to the command.
		_, err := cmd.AddGroup("Help Options", "", &help)
		if err != nil {
			return err
		}
	}

	// Do the same for each sub command, recursively.
	for _, subCmd := range cmd.Commands() {
		if err := addHelpGroups(subCmd); err != nil {
			return err
		}
	}
	return nil
}
