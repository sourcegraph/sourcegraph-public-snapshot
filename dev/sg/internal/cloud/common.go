package cloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrUserCancelled = errors.New("user cancelled")
var ErrWrongBranch = errors.New("wrong current branch")
var ErrBranchOutOfSync = errors.New("branch is out of sync with remote")
var ErrNotEphemeralInstance = errors.New("instance is not ephemeral")
var ErrInstanceStatusNotComplete = errors.New("instance is not not in completed status")
var ErrExpiredInstance = errors.New("instance has already expired")

const (
	CloudEmoji = "☁️"
	FAQLink    = "https://www.notion.so/sourcegraph/How-to-deploy-my-branch-on-an-ephemeral-Cloud-instance-dac45846ca2a4e018c802aba37cf6465?pvs=4#20cb92ae27464891a9d03650b4d67cee"
)

func withFAQMarkdown(original string) string {
	return fmt.Sprintf("%s\n[FAQ](%s)", original, FAQLink)
}

func withFAQ(original string) string {
	return fmt.Sprintf("%s\nFAQ - %s", original, FAQLink)

}

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
