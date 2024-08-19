package ffcli

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff"
)

// Command combines a main function with a flag.FlagSet, and zero or more
// sub-commands. A commandline program can be represented as a declarative tree
// of commands.
type Command struct {
	// Name of the command. Used for sub-command matching, and as a replacement
	// for Usage, if no Usage string is provided. Required for sub-commands,
	// optional for the root command.
	Name string

	// Usage for this command. Printed at the top of the help output.
	// Recommended but not required. Should be one line of the form
	//
	//     cmd [flags] subcmd [flags] <required> [<optional> ...]
	//
	Usage string

	// ShortHelp is printed next to the command name when it appears as a
	// sub-command, in the help output of its parent command. Recommended.
	ShortHelp string

	// LongHelp is printed in the help output, after usage and before flags.
	// Typically a paragraph or more of prose-like text, providing more explicit
	// context and guidance than what is implied by flags and arguments.
	// Optional.
	LongHelp string

	// UsageFunc generates a complete usage output, displayed to the user when
	// the -h flag is passed. The function is invoked with its corresponding
	// command, and its output should reflect the command's short and long help
	// strings, subcommands, and available flags. Optional; if not provided, a
	// suitable, compact default is used.
	UsageFunc func(*Command) string

	// FlagSet associated with this command. Optional, but if none is provided,
	// an empty FlagSet will be defined and attached during Run, so that the -h
	// flag works as expected.
	FlagSet *flag.FlagSet

	// Options provided to ff.Parse when parsing arguments for this command.
	// Optional.
	Options []ff.Option

	// Subcommands accessible underneath (i.e. after) this command. Optional.
	Subcommands []*Command

	// Exec is invoked if this command has been determined to be the terminal
	// command selected by the arguments provided to Run. The args passed to
	// Exec are the args left over after flags parsing. Optional.
	//
	// If Exec returns flag.ErrHelp, Run will behave as if -h were passed and
	// emit the complete usage output.
	Exec func(args []string) error
}

// Run parses the commandline arguments for this command and all sub-commands
// recursively, and invokes the Exec function for the terminal command, if it's
// defined.
func (c *Command) Run(args []string) error {
	if c.FlagSet == nil {
		c.FlagSet = flag.NewFlagSet(c.Name, flag.ExitOnError)
	}

	if c.UsageFunc == nil {
		c.UsageFunc = DefaultUsageFunc
	}

	c.FlagSet.Usage = func() {
		fmt.Fprintln(c.FlagSet.Output(), c.UsageFunc(c))
	}

	if err := ff.Parse(c.FlagSet, args, c.Options...); err != nil {
		return err
	}

	if len(c.FlagSet.Args()) > 0 {
		for _, subcommand := range c.Subcommands {
			if strings.EqualFold(c.FlagSet.Args()[0], subcommand.Name) {
				return subcommand.Run(c.FlagSet.Args()[1:])
			}
		}
	}

	if c.Exec != nil {
		err := c.Exec(c.FlagSet.Args())
		if errors.Is(err, flag.ErrHelp) {
			c.FlagSet.Usage()
		}
		return err
	}
	return nil
}

// DefaultUsageFunc is the default UsageFunc used for all commands
// if no custom UsageFunc is provided.
func DefaultUsageFunc(c *Command) string {
	var b strings.Builder

	fmt.Fprintf(&b, "USAGE\n")
	if c.Usage != "" {
		fmt.Fprintf(&b, "  %s\n", c.Usage)
	} else {
		fmt.Fprintf(&b, "  %s\n", c.Name)
	}
	fmt.Fprintf(&b, "\n")

	if c.LongHelp != "" {
		fmt.Fprintf(&b, "%s\n\n", c.LongHelp)
	}

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, subcommand := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", subcommand.Name, subcommand.ShortHelp)
		}
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	if countFlags(c.FlagSet) > 0 {
		fmt.Fprintf(&b, "FLAGS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			def := f.DefValue
			if def == "" {
				def = "..."
			}
			fmt.Fprintf(tw, "  -%s %s\t%s\n", f.Name, def, f.Usage)
		})
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	return strings.TrimSpace(b.String())
}

func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })
	return n
}
