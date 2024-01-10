package perforce

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetChangelistByID(ctx context.Context, p4home, p4port, p4user, p4passwd, changelistID string) (*p4types.Changelist, error) {
	cmd := exec.CommandContext(
		ctx,
		"p4",
		"-Mj",
		"-z", "tag",
		"changes",
		"-r",      // list in reverse order, which means that the given changelist id will be the first one listed
		"-m", "1", // limit output to one record, so that the given changelist is the only one listed
		"-l",               // use a long listing, which includes the whole commit message
		"-e", changelistID, // start from this changelist and go up
	)
	cmd.Env = append(os.Environ(),
		"P4PORT="+p4port,
		"P4USER="+p4user,
		"P4PASSWD="+p4passwd,
		"HOME="+p4home,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 changes context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 changes" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}
		return nil, err
	}

	output := bytes.TrimSpace(out)

	if len(output) == 0 {
		return nil, errors.New("invalid changelist " + changelistID)
	}

	pcl, err := parseChangelistOutput(output)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse change output")
	}

	return pcl, nil
}

func GetChangelistByClient(ctx context.Context, p4port, p4user, p4passwd, workDir, client string) (*p4types.Changelist, error) {
	cmd := exec.CommandContext(
		ctx,
		"p4",
		"-Mj",
		"-z", "tag",
		"changes",
		"-r",      // list in reverse order, which means that the given changelist id will be the first one listed
		"-m", "1", // limit output to one record, so that the given changelist is the only one listed
		"-l", // use a long listing, which includes the whole commit message
		"-c", client,
	)
	cmd.Env = append(os.Environ(),
		"P4PORT="+p4port,
		"P4USER="+p4user,
		"P4PASSWD="+p4passwd,
		"P4CLIENT="+client,
	)
	cmd.Dir = workDir

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 changes context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 changes" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}
		return nil, err
	}

	output := bytes.TrimSpace(out)

	if len(output) == 0 {
		return nil, errors.New("no changelist found for client " + client)
	}

	pcl, err := parseChangelistOutput(output)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse change output")
	}

	return pcl, nil
}

type changelistJson struct {
	// Change is the number of the changelist.
	Change     string `json:"change"`
	ChangeType string `json:"changeType"`
	Client     string `json:"client"`
	Desc       string `json:"desc"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Time       string `json:"time"`
	User       string `json:"user"`
}

// parseChangelistOutput parses one JSON line of p4 changes output.
// output should be whitespace-trimmed and not empty.
func parseChangelistOutput(output []byte) (*p4types.Changelist, error) {
	var cidj changelistJson
	err := json.Unmarshal(output, &cidj)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal change output")
	}

	state, err := parseChangelistState(cidj.Status)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse changelist state")
	}

	intTime, err := strconv.Atoi(cidj.Time)
	if err != nil {
		return nil, errors.Wrap(err, "invalid time: "+cidj.Time)
	}

	creationDate := time.Unix(int64(intTime), 0)

	return &p4types.Changelist{
		ID:           cidj.Change,
		State:        state,
		Author:       cidj.User,
		CreationDate: creationDate,
		Title:        cidj.Client,
		Message:      strings.TrimSpace(cidj.Desc),
	}, nil
}

func parseChangelistState(state string) (p4types.ChangelistState, error) {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "submitted":
		return p4types.ChangelistStateSubmitted, nil
	case "pending":
		return p4types.ChangelistStatePending, nil
	case "shelved":
		return p4types.ChangelistStateShelved, nil
	case "closed":
		return p4types.ChangelistStateClosed, nil
	default:
		return "", errors.Newf("invalid Perforce changelist state: %s", state)
	}
}
