package perforce

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"

	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetChangeListByIDArguments are the arguments for GetChangelistByID.
type GetChangeListByIDArguments struct {
	// ReposDir is the directory where the repositories are stored.
	ReposDir string
	// P4Home is the path to the directory that 'p4' will use as $HOME
	// and where it will store cache data.
	P4Home string

	// P4PORT is the address of the Perforce server.
	P4Port string
	// P4User is the Perforce username to authenticate with.
	P4User string
	// P4Passwd is the Perforce password to authenticate with.
	P4Passwd string

	// ChangelistID is the ID of the changelist to get.
	ChangelistID string
}

func GetChangelistByID(ctx context.Context, args GetChangeListByIDArguments) (*p4types.Changelist, error) {
	options := []P4OptionFunc{
		WithAuthentication(args.P4User, args.P4Passwd),
		WithHost(args.P4Port),
	}

	options = append(options, WithArguments(
		"-Mj",
		"-z", "tag",
		"changes",
		"-r",      // list in reverse order, which means that the given changelist id will be the first one listed
		"-m", "1", // limit output to one record, so that the given changelist is the only one listed
		"-l",                    // use a long listing, which includes the whole commit message
		"-e", args.ChangelistID, // start from this changelist and go up
	))

	scratchDir, err := gitserverfs.TempDir(args.ReposDir, "p4-changelist-")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir to invoke 'p4 changes'")
	}
	defer os.Remove(scratchDir)

	cmd := NewBaseCommand(ctx, args.P4Home, scratchDir, options...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 changes context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 changes" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd.Unwrap()))
		}
		return nil, err
	}

	output := bytes.TrimSpace(out)

	if len(output) == 0 {
		return nil, errors.New("invalid changelist " + args.ChangelistID)
	}

	pcl, err := parseChangelistOutput(output)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse change output")
	}

	return pcl, nil
}

// GetChangeListByClientArguments are the arguments for GetChangelistByClient.
type GetChangeListByClientArguments struct {
	// P4Home is the path to the directory that 'p4' will use as $HOME
	// and where it will store cache data.
	P4Home string

	// P4PORT is the address of the Perforce server.
	P4Port string
	// P4User is the Perforce username to authenticate with.
	P4User string
	// P4Passwd is the Perforce password to authenticate with.
	P4Passwd string

	// WorkDir is the working directory of the command.
	WorkDir string

	// Client is the client name to use to get the changelist.
	Client string
}

func GetChangelistByClient(ctx context.Context, args GetChangeListByClientArguments) (*p4types.Changelist, error) {
	options := []P4OptionFunc{
		WithAuthentication(args.P4User, args.P4Passwd),
		WithHost(args.P4Port),
		WithClient(args.Client),
	}

	options = append(options, WithArguments(
		"-Mj",
		"-z", "tag",
		"changes",
		"-r",      // list in reverse order, which means that the given changelist id will be the first one listed
		"-m", "1", // limit output to one record, so that the given changelist is the only one listed
		"-l", // use a long listing, which includes the whole commit message
		"-c", args.Client,
	))

	cmd := NewBaseCommand(ctx, args.P4Home, args.WorkDir, options...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 changes context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 changes" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd.Unwrap()))
		}
		return nil, err
	}

	output := bytes.TrimSpace(out)

	if len(output) == 0 {
		return nil, errors.New("no changelist found for client " + args.Client)
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
