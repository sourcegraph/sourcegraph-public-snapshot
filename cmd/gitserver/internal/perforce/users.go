package perforce

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PerforceUserType string

const (
	PerforceUserTypeStandard PerforceUserType = "standard"
	PerforceUserTypeOperator PerforceUserType = "operator"
	PerforceUserTypeService  PerforceUserType = "service"
)

// PerforceUser is a definition of a user that matches the format returned from
// `p4 -Mj -ztag users`.
type PerforceUser struct {
	Email    string `json:"Email,omitempty"`
	User     string `json:"User,omitempty"`
	Password string `json:"Password,omitempty"`
	FullName string `json:"FullName,omitempty"`
	// Access is seconds since the Epoch, but p4 quotes it in the output, so it's a string
	Access string           `json:"Access,omitempty"`
	Update string           `json:"Update,omitempty"`
	Type   PerforceUserType `json:"type,omitempty"`
}

// P4Users returns all of the depots to which the user has access on the host
// and whose names match the given nameFilter, which can contain asterisks (*) for wildcards
// if nameFilter is blank, return all depots
func P4Users(ctx context.Context, port, username, password string) ([]PerforceUser, error) {
	cmd := exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "users")
	cmd.Env = append(os.Environ(),
		"P4PORT="+port,
		"P4USER="+username,
		"P4PASSWD="+password,
	)

	out, err := executil.RunCommandCombinedOutput(ctx, wrexec.Wrap(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 users context error")
		}
		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 users" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}
		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no users. Maybe the user doesn't have access to any users?
		return nil, nil
	}

	users := make([]PerforceUser, 0)
	lr := byteutils.NewLineReader(out)
	for lr.Scan() {
		line := lr.Line()
		// the output of `p4 -Mj -ztag users` is a series of JSON-formatted depot definitions, one per line
		var user PerforceUser
		err := json.Unmarshal(line, &user)
		if err != nil {
			return nil, errors.Wrap(err, "malformed output from p4 users")
		}
		users = append(users, user)
	}

	return users, nil
}
