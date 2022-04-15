package main

import (
	"sort"
	"strings"

	"github.com/agext/levenshtein"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

// reconstructArgs reconstructs the argument string from the command context lineage.
func reconstructArgs(cmd *cli.Context) string {
	lineage := cmd.Lineage()
	root := lineage[len(lineage)-1]
	args := append([]string{cmd.App.Name}, root.Args().Slice()...)
	return strings.Join(args, " ")
}

// suggestSubcommandsAction is a cli.Action that calculates and suggests subcommands
// similar to the first argument.
func suggestSubcommandsAction(cmd *cli.Context) error {
	s := cmd.Args().First()
	if s == "" {
		// Use default if no args are provided
		return cli.ShowSubcommandHelp(cmd)
	}
	suggestCommands(cmd, s)
	return cli.Exit("", 1)
}

// suggestCommands is a cli.CommandNotFoundFunc that calculates and suggests subcommands
// similar arg.
func suggestCommands(cmd *cli.Context, arg string) {
	var cmds []*cli.Command
	if cmd.Command == nil || cmd.Command.Name == "" {
		cmds = cmd.App.Commands
	} else {
		cmds = cmd.Command.Subcommands
	}

	args := reconstructArgs(cmd)
	writeOrangeLinef("command '%s %s' not found", args, arg)

	suggestions := makeSuggestions(cmds, arg, 0.3, 3)
	if len(suggestions) == 0 {
		stdout.Out.Writef("try running '%s -h' for help", args)
		return
	}

	stdout.Out.Write("did you mean:")
	for _, s := range suggestions {
		writeFingerPointingLinef("  %s %s", args, s.name)
	}
	stdout.Out.Write("learn more about each command with the '-h' flag")
}

type commandSuggestion struct {
	name  string
	score float64
}

type commandSuggestions []commandSuggestion

// makeSuggestions returns the n most similar command names to arg, from most similar to
// least, where the levenshtein score is above the threshold.
func makeSuggestions(cmds []*cli.Command, arg string, threshold float64, n int) commandSuggestions {
	suggestions := commandSuggestions{}
	for _, c := range cmds {
		if c.Hidden || c.Name == "help" {
			continue
		}

		// Get the best suggestion for the names this command has, so as to make only one
		// suggestion per command
		closestName := commandSuggestion{}
		for _, n := range c.Names() {
			score := levenshtein.Match(n, arg, levenshtein.NewParams())
			if closestName.score < score {
				closestName.name = n
				closestName.score = score
			}
		}

		// Only suggest above our threshold
		if closestName.score >= threshold {
			suggestions = append(suggestions, closestName)
		}
	}

	sort.Sort(suggestions)
	if len(suggestions) < n {
		return suggestions
	}
	return suggestions[:n]
}

func (cs commandSuggestions) Len() int {
	return len(cs)
}

func (cs commandSuggestions) Less(i, j int) bool {
	// Higher score = better
	return cs[i].score > cs[j].score
}

func (cs commandSuggestions) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
