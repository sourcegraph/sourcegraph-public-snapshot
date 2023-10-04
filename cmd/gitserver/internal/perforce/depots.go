package perforce

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PerforceDepotType string

const (
	Local   PerforceDepotType = "local"
	Remote  PerforceDepotType = "remote"
	Stream  PerforceDepotType = "stream"
	Spec    PerforceDepotType = "spec"
	Unload  PerforceDepotType = "unload"
	Archive PerforceDepotType = "archive"
	Tangent PerforceDepotType = "tangent"
	Graph   PerforceDepotType = "graph"
)

// PerforceDepot is a definiton of a depot that matches the format
// returned from `p4 -Mj -ztag depots`
type PerforceDepot struct {
	Desc string `json:"desc,omitempty"`
	Map  string `json:"map,omitempty"`
	Name string `json:"name,omitempty"`
	// Time is seconds since the Epoch, but p4 quotes it in the output, so it's a string
	Time string `json:"time,omitempty"`
	// Type is local, remote, stream, spec, unload, archive, tangent, graph
	Type PerforceDepotType `json:"type,omitempty"`
}

// P4Depots returns all of the depots to which the user has access on the host
// and whose names match the given nameFilter, which can contain asterisks (*) for wildcards
// if nameFilter is blank, return all depots
func P4Depots(ctx context.Context, host, username, password, nameFilter string) ([]PerforceDepot, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if nameFilter == "" {
		cmd = exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "depots")
	} else {
		cmd = exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "depots", "-e", nameFilter)
	}
	cmd.Env = append(os.Environ(),
		"P4PORT="+host,
		"P4USER="+username,
		"P4PASSWD="+password,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 depots context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 depots" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}
		return nil, err
	}
	depots := make([]PerforceDepot, 0)
	if len(out) > 0 {
		// the output of `p4 -Mj -ztag depots` is a series of JSON-formatted depot definitions, one per line
		buf := bufio.NewScanner(bytes.NewBuffer(out))
		for buf.Scan() {
			depot := PerforceDepot{}
			err := json.Unmarshal(buf.Bytes(), &depot)
			if err != nil {
				return nil, errors.Wrap(err, "malformed output from p4 depots")
			}
			depots = append(depots, depot)
		}
		if err := buf.Err(); err != nil {
			return nil, errors.Wrap(err, "malformed output from p4 depots")
		}
		return depots, nil
	}

	// no error, but also no depots. Maybe the user doesn't have access to any depots?
	return depots, nil
}
