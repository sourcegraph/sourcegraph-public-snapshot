package cli

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/toolchain"
)

func init() {
	c, err := CLI.AddCommand("tool",
		"run a tool",
		"Run a srclib tool with the specified arguments.",
		&toolCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
	c.ArgsRequired = true
}

type ToolCmd struct {
	ToolchainExecOpt

	Args struct {
		Toolchain ToolchainPath `name:"TOOLCHAIN" description:"toolchain path of the toolchain to run"`
		Tool      ToolName      `name:"TOOL" description:"tool subcommand name to run (in TOOLCHAIN)"`
		ToolArgs  []string      `name:"ARGS" description:"args to pass to TOOL"`
	} `positional-args:"yes" required:"yes"`
}

var toolCmd ToolCmd

func (c *ToolCmd) Execute(args []string) error {
	tc, err := toolchain.Open(string(c.Args.Toolchain), c.ToolchainMode())
	if err != nil {
		log.Fatal(err)
	}

	var cmder interface {
		Command() (*exec.Cmd, error)
	}
	if c.Args.Tool != "" {
		cmder, err = toolchain.OpenTool(string(c.Args.Toolchain), string(c.Args.Tool), c.ToolchainMode())
	} else {
		cmder = tc
	}

	// HACK: Buffer stdout to work around
	// https://github.com/docker/docker/issues/3631. Otherwise, lots
	// of builds fail. Also, if a lot of data is printed, the return
	// code is 0, but JSON parsing fails, retry it up to 4 times until
	// JSON parsing succeeds.
	for tries := 4; ; tries-- {
		cmd, err := cmder.Command()
		if err != nil {
			log.Fatal(err)
		}
		cmd.Args = append(cmd.Args, c.Args.ToolArgs...)
		cmd.Stderr = os.Stderr
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stdin = os.Stdin
		if GlobalOpt.Verbose {
			log.Printf("Running tool: %v", cmd.Args)
		}
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		b := out.Bytes()

		if tries > 0 {
			// Try parsing (see HACK) above.
			if len(b) > 2000 {
				var o interface{}
				if err := json.Unmarshal(b, &o); err != nil {
					log.Printf("Suspect JSON output by %v (%d bytes, parse error %q); retrying %d more times. (This is a workaround for https://github.com/docker/docker/issues/3631.)", cmd.Args, len(b), err, tries)
					continue // retry
				}
			}
		}

		os.Stdout.Write(b)
		return nil
	}
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
