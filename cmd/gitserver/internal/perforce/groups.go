package perforce

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4Groups returns all usernames that are members of the given group.
func P4GroupMembers(ctx context.Context, port, username, password, group string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "p4", "-Mj", "-ztag", "group", "-o", group)
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
			err = errors.Wrapf(err, `failed to run command "p4 group" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd))
		}

		return nil, err
	}

	if len(out) == 0 {
		// no error, but also no members. Maybe the group doesn't have any members?
		return nil, nil
	}

	return parseP4GroupMembers(out)
}

func parseP4GroupMembers(out []byte) ([]string, error) {
	var jsonGroup map[string]any
	err := json.Unmarshal(out, &jsonGroup)
	if err != nil {
		return nil, errors.Wrap(err, "malformed output from p4 group")
	}

	users := make([]string, 0)
	currentUserIdx := 0
	for {
		user, ok := jsonGroup[fmt.Sprintf("Users%d", currentUserIdx)]
		if !ok {
			break
		}
		username, ok := user.(string)
		if !ok {
			continue
		}
		users = append(users, username)
		currentUserIdx++
	}

	return users, nil
}
