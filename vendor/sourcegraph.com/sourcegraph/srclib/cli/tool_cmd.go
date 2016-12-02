package cli

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("tool",
			"run a tool",
			"Run a srclib tool with the specified arguments.",
			&toolCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
		c.ArgsRequired = true
	})
}

type ToolCmd struct {
	Args struct {
		Toolchain ToolchainPath `name:"TOOLCHAIN" description:"toolchain path of the toolchain to run"`
		Tool      ToolName      `name:"TOOL" description:"tool subcommand name to run (in TOOLCHAIN)"`
		ToolArgs  []string      `name:"ARGS" description:"args to pass to TOOL"`
	} `positional-args:"yes" required:"yes"`
}

var toolCmd ToolCmd

func (c *ToolCmd) Execute(args []string) error {
	cmdName, err := toolchain.Command(string(c.Args.Toolchain))
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(cmdName)
	if c.Args.Tool != "" {
		cmd.Args = append(cmd.Args, string(c.Args.Tool))
		cmd.Args = append(cmd.Args, c.Args.ToolArgs...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if GlobalOpt.Verbose {
		log.Printf("Running tool: %v", cmd.Args)
	}
	return cmd.Run()
}

type ToolName string

func (t ToolName) Complete(match string) []flags.Completion {
	// Assume toolchain is the last arg.
	toolchainPath := os.Args[len(os.Args)-2]
	tc, err := toolchain.Lookup(toolchainPath)
	if err != nil {
		log.Println(err)
		return nil
	}
	c, err := tc.ReadConfig()
	if err != nil {
		log.Println(err)
		return nil
	}
	var comps []flags.Completion
	for _, tt := range c.Tools {
		if strings.HasPrefix(tt.Subcmd, match) {
			comps = append(comps, flags.Completion{Item: tt.Subcmd})
		}
	}
	return comps
}
