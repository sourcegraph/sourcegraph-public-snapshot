package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

type RecentChangesOpts struct {
	Count int
	Next  bool
	Color bool
}

// RecentSGChanges provides a summary and list of changes to sg-related code (inferred).
func RecentSGChanges(build string, opts RecentChangesOpts) (string, []string, error) {
	var (
		title   string
		logArgs = []string{
			// Format nicely
			"log", "--pretty=%C(reset)%s %C(dim)%h by %an, %ar",
			// Filter out stuff we don't want
			"--no-merges",
			// Limit entries
			fmt.Sprintf("--max-count=%d", opts.Count),
		}
	)
	if opts.Color {
		logArgs = append(logArgs, "--color=always")
	} else {
		logArgs = append(logArgs, "--color=never")
	}
	if build != "dev" {
		current := strings.TrimPrefix(build, "dev-")
		if opts.Next {
			logArgs = append(logArgs, current+"..origin/main")
			title = fmt.Sprintf("Changes since sg release %s", build)
		} else {
			logArgs = append(logArgs, current)
			title = fmt.Sprintf("Changes in sg release %s", build)
		}
	} else {
		std.Out.WriteWarningf("Dev version detected - just showing recent changes.")
		title = "Recent sg changes"
	}

	gitLog := exec.Command("git", append(logArgs, "--", "./dev/sg")...)
	gitLog.Env = os.Environ()
	out, err := run.InRoot(gitLog)
	if err != nil {
		return title, nil, err
	}
	return title, strings.Split(strings.TrimSpace(out), "\n"), nil
}
