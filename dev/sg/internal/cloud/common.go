package cloud

import (
	"context"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrUserCancelled = errors.New("user cancelled")
var ErrWrongBranch = errors.New("wrong current branch")
var ErrBranchOutOfSync = errors.New("branch is out of sync with remote")
var ErrNotEphemeralInstance = errors.New("instance is not ephemeral")
var ErrExpiredInstance = errors.New("instance has already expired")

const CloudEmoji = "☁️"

func sanitizeInstanceName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func inferInstanceNameFromBranch(ctx context.Context) (string, error) {
	currentBranch, err := repo.GetCurrentBranch(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to determine current branch")
	}
	return sanitizeInstanceName(currentBranch), nil
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
