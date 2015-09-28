package toolchain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"sourcegraph.com/sourcegraph/srclib"
)

// ToolInfo describes a tool in a toolchain.
type ToolInfo struct {
	// Subcmd is the subcommand name of this tool.
	//
	// By convention, this is the same as Op in toolchains that only have one
	// tool that performs this operation (e.g., a toolchain's "graph" subcommand
	// performs the "graph" operation).
	Subcmd string

	// Op is the operation that this tool performs (e.g., "scan", "graph",
	// "deplist", etc.).
	Op string

	// SourceUnitTypes is a list of source unit types (e.g., "GoPackage") that
	// this tool can operate on.
	//
	// If this tool doesn't operate on source units (for example, it operates on
	// directories or repositories, such as the "blame" tools), then this will
	// be empty.
	//
	// TODO(sqs): determine how repository- or directory-level tools will be
	// defined.
	SourceUnitTypes []string `json:",omitempty"`
}

// ListTools lists all tools in all available toolchains (returned by List). If
// op is non-empty, only tools that perform that operation are returned.
func ListTools(op string) ([]*srclib.ToolRef, error) {
	tcs, err := List()
	if err != nil {
		return nil, err
	}

	var tools []*srclib.ToolRef
	for _, tc := range tcs {
		c, err := tc.ReadConfig()
		if err != nil {
			return nil, err
		}

		for _, tool := range c.Tools {
			if op == "" || tool.Op == op {
				tools = append(tools, &srclib.ToolRef{Toolchain: tc.Path, Subcmd: tool.Subcmd})
			}
		}
	}
	return tools, nil
}

// OpenTool opens a tool in toolchain (which is a toolchain path) named subcmd.
// The mode parameter controls how the toolchain is opened.
func OpenTool(toolchain, subcmd string, mode Mode) (Tool, error) {
	tc, err := Open(toolchain, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open tool (%s %s): %s", toolchain, subcmd, err)
	}

	return &tool{tc, subcmd, log.New(os.Stderr, "", 0)}, nil
}

// A Tool is a subcommand of a Toolchain that performs an single operation, such
// as one type of analysis on a source unit.
type Tool interface {
	// Command returns an *exec.Cmd suitable for running this tool.
	Command() (*exec.Cmd, error)

	// Run executes this tool with args (sending the JSON-serialization of input
	// on stdin, if input is non-nil) and parses the JSON response into resp.
	Run(arg []string, input, resp interface{}) error

	// SetLogger sets the logger for Tool to l.
	SetLogger(l *log.Logger)
}

type tool struct {
	tc     Toolchain
	subcmd string
	log    *log.Logger
}

func (t *tool) SetLogger(l *log.Logger) {
	if l == nil {
		panic("SetLogger: logger cannot be nil")
	}
	t.log = l
}

func (t *tool) Command() (*exec.Cmd, error) {
	// make command
	cmd, err := t.tc.Command()
	if err != nil {
		return nil, err
	}
	cmd.Args = append(cmd.Args, t.subcmd)
	return cmd, nil
}

// TODO(sqs): is it possible for an early return to leave the subprocess running?
func (t *tool) Run(arg []string, input, resp interface{}) error {
	cmd, err := t.Command()
	if err != nil {
		return err
	}
	cmd.Args = append(cmd.Args, arg...)
	cmd.Stderr = os.Stderr

	t.log.Printf("Running: %v", cmd.Args)

	var stdin io.WriteCloser
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		t.log.Printf("  --> with input %s", data)

		stdin, err = cmd.StdinPipe()
		if err != nil {
			return err
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	if input != nil {
		w := bufio.NewWriter(stdin)
		if err := json.NewEncoder(w).Encode(input); err != nil {
			w.Flush()
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
		if err := stdin.Close(); err != nil {
			return err
		}
	}

	if err := json.NewDecoder(stdout).Decode(resp); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
