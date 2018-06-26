package cxp

import (
	"context"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// ExecCmdFunc returns a git.CmdFunc that can be used by CXP servers to execute Git commands on the
// client. It requires that the client supports the LSP exec extension.
func ExecCmdFunc(command string, conn *jsonrpc2.Conn) git.CmdFunc {
	return func(args []string) git.Cmd {
		return &execCmd{conn: conn, params: lspext.ExecParams{Command: command, Arguments: args}}
	}
}

type execCmd struct {
	conn   *jsonrpc2.Conn
	params lspext.ExecParams
}

func (c *execCmd) Output(ctx context.Context) ([]byte, error) {
	var result lspext.ExecResult
	if err := c.conn.Call(ctx, "workspace/exec", c.params, &result); err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		stderr := result.Stderr
		if len(stderr) > 100 {
			stderr = stderr[:100] + "... (truncated)"
		}
		return nil, fmt.Errorf("nonzero exit code %d (stderr: %q)", result.ExitCode, stderr)
	}
	return []byte(result.Stdout), nil
}

func (c execCmd) String() string { return fmt.Sprintf("%+v", c.params) }
