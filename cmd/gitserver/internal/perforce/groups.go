package perforce

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/executil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// P4GroupMembers returns all usernames that are members of the given group.
func P4GroupMembers(ctx context.Context, p4home, p4port, p4user, p4passwd, group string) ([]string, error) {
	options := []P4OptionFunc{
		WithAuthentication(p4user, p4passwd),
		WithHost(p4port),
		WithAlternateHomeDir(p4home),
	}

	options = append(options, WithArguments("-Mj", "-ztag", "group", "-o", group))
	options = append(options, WithEnvironment(os.Environ()...))

	scratchDir, err := os.MkdirTemp(p4home, "p4-group-")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir to invoke 'p4 group'")
	}
	defer os.Remove(scratchDir)

	cmd := NewBaseCommand(ctx, scratchDir, options...)
	out, err := executil.RunCommandCombinedOutput(ctx, cmd)
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrap(ctxerr, "p4 group context error")
		}

		if len(out) > 0 {
			err = errors.Wrapf(err, `failed to run command "p4 group" (output follows)\n\n%s`, specifyCommandInErrorMessage(string(out), cmd.Unwrap()))
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
		currentUserIdx++
		if !ok {
			break
		}
		username, ok := user.(string)
		if !ok {
			continue
		}
		users = append(users, username)
	}

	return users, nil
}
