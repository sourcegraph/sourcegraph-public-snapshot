package perforce

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	v1 "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4ProtectsForUser returns all protect definitions that apply to the given username.
func P4ProtectsForUser(ctx context.Context, p4port, p4username, p4password, username string) ([]*v1.PerforceProtect, error) {
	cmd := exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "protects", "-u", username)
	cmd.Env = append(os.Environ(),
		"P4PORT="+p4port,
		"P4USER="+p4username,
		"P4PASSWD="+p4password,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 protects context error")
		}

		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 protects" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}

		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no protects.
		return nil, nil
	}

	return parseP4Protects(out)
}

// P4ProtectsForUser returns all protect definitions that apply to the given depot.
func P4ProtectsForDepot(ctx context.Context, p4port, p4username, p4password, depot string) ([]*v1.PerforceProtect, error) {
	cmd := exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "protects", "-a", depot)
	cmd.Env = append(os.Environ(),
		"P4PORT="+p4port,
		"P4USER="+p4username,
		"P4PASSWD="+p4password,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 protects context error")
		}

		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 protects" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}

		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no protects.
		return nil, nil
	}

	return parseP4Protects(out)
}

type perforceJSONProtect struct {
	DepotFile string  `json:"depotFile"`
	Host      string  `json:"host"`
	Line      string  `json:"line"`
	Perm      string  `json:"perm"`
	IsGroup   *string `json:"isgroup,omitempty"`
	Unmap     *string `json:"unmap,omitempty"`
	User      string  `json:"user"`
}

func parseP4Protects(out []byte) ([]*v1.PerforceProtect, error) {
	protects := make([]*v1.PerforceProtect, 0)

	lr := byteutils.NewLineReader(out)
	for lr.Scan() {
		line := lr.Line()

		// Trim whitespace
		line = bytes.TrimSpace(line)

		var parsedLine perforceJSONProtect
		if err := json.Unmarshal(line, &parsedLine); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal protect line")
		}

		entityType := "user"
		if parsedLine.IsGroup != nil {
			entityType = "group"
		}

		// TODO: Reconcile names?
		protects = append(protects, &v1.PerforceProtect{
			Host:        parsedLine.Host,
			EntityType:  entityType,
			EntityName:  parsedLine.User,
			Match:       parsedLine.DepotFile,
			IsExclusion: parsedLine.Unmap != nil,
			Level:       parsedLine.Perm,
		})
	}

	return protects, nil
}
