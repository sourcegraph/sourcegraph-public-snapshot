package proxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// handleWorkspaceExec handles exec requests adherent to the LSP exec extension (see
// language-server-protocol/extension-exec.md).
func (c *serverProxyConn) handleWorkspaceExec(ctx context.Context, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lspext.ExecParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if params.Command != "git" {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: fmt.Sprintf("workspace/exec: unsupported command %q (supported commands are: git)", params.Command)}
	}

	repo := gitserver.Repo{Name: c.id.rootURI.Repo()}
	stdout, stderr, exitCode, err := git.ExecSafe(ctx, repo, params.Arguments)
	if err != nil {
		return nil, err
	}
	return lspext.ExecResult{Stdout: string(stdout), Stderr: string(stderr), ExitCode: exitCode}, nil
}
