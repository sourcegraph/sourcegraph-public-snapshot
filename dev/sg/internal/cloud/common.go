package cloud

import (
	"context"
	"strings"

	"github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var ErrUserCancelled = errors.New("user cancelled")
var ErrWrongBranch = errors.New("wrong current branch")
var ErrBranchOutOfSync = errors.New("branch is out of sync with remote")
var ErrNotEphemeralInstance = errors.New("instance is not ephemeral")
var ErrExpiredInstance = errors.New("instance has already expired")

const CloudEmoji = "☁️"

func sanitizeInstanceName(name string) string {
	name = strings.ToLower(name)
	return strings.ReplaceAll(name, "/", "-")
}

func inferInstanceNameFromBranch(ctx context.Context) (string, error) {
	currentBranch, err := repo.GetCurrentBranch(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to determine current branch")
	}
	return sanitizeInstanceName(currentBranch), nil
}

func wipAction(actionFn cli.ActionFunc) cli.ActionFunc {
	if actionFn == nil {
		return nil
	}
	return func(ctx *cli.Context) error {
		if err := printWIPNotice(ctx); err != nil {
			return err
		}

		return actionFn(ctx)
	}
}

func printWIPNotice(ctx *cli.Context) error {
	if ctx.Bool("skip-wip-notice") {
		return nil
	}
	notice := output.Line("🧪", output.StyleBold, "EXPERIMENTAL COMMAND - Do you want to continue? (yes/no)")

	var answer string
	if _, err := std.FancyPromptAndScan(std.Out, notice, &answer); err != nil {
		return err
	}

	if oneOfEquals(answer, "yes", "y") {
		return nil
	}

	return ErrUserCancelled
}

func oneOfEquals(value string, i ...string) bool {
	for _, item := range i {
		if value == item {
			return true
		}
	}
	return false
}

func getGCloudAccount(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "gcloud", "config", "get", "account").Run().String()
}
