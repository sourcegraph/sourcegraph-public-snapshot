package docker

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	illegalApkAddRegexp = regexp.MustCompile(`[^(\s)>]+[=|<][0-9]+[^\s]+`)
	illegalAlpineRegexp = regexp.MustCompile(`^alpine:[^@]*`)
)

type commandCheck struct {
	name  string
	check func(instructions.Command) error
}

// See .hadolint.yaml DL3018 docstring
var commandCheckApkAdd = commandCheck{
	name: "illegal apk add version requirement",
	check: func(c instructions.Command) error {
		if runCmd, okay := c.(*instructions.RunCommand); okay {
			run := runCmd.String()
			if strings.Contains(run, "apk") && strings.Contains(run, "add") {
				matches := illegalApkAddRegexp.FindAllString(run, -1)
				if len(matches) == 0 {
					return nil
				}
				return errors.Newf("%s (use '>=' or no version requirement instead)", strings.Join(matches, ", "))
			}
		}
		return nil
	},
}

type stageCheck struct {
	name  string
	check func(instructions.Stage) error
}

// Using 'alpine' is forbidden. Use 'sourcegraph/alpine' instead which provides:
//
// - Fixes DNS resolution in some deployment environments.
// - A non-root 'sourcegraph' user.
// - Static UID and GIDs that are consistent across all containers.
// - Base packages like 'tini' and 'curl' that we expect in all containers.
//
// You should use 'sourcegraph/alpine' even in build stages for consistency sake.
// Use explicit 'USER root' and 'USER sourcegraph' sections when adding packages, etc.
//
// If the linter is incorrect, add the comment "CHECK:ALPINE_OK" as a docstring on the
// named stage where the import is used.
var stageCheckNoAlpine = stageCheck{
	name: "forbidden 'alpine' usage",
	check: func(s instructions.Stage) error {
		matches := illegalAlpineRegexp.FindAllString(s.BaseName, -1)
		if len(matches) == 0 {
			return nil
		}
		const okayFlag = "CHECK:ALPINE_OK"
		if strings.Contains(s.Comment, okayFlag) {
			return nil
		}
		if s.Name != "" {
			return errors.Newf("%s (use 'sourcegraph/alpine' instead, or add the comment '%s %s')",
				strings.Join(matches, ", "), s.Name, okayFlag)
		}
		return errors.Newf("%s (use 'sourcegraph/alpine' instead, or add 'AS alpine_base' to the stage and comment 'alpine_base %s')",
			strings.Join(matches, ", "), okayFlag)
	},
}

// LintDockerfile is a linter for Dockerfile directives.
func LintDockerfile(dockerfile string) func(is []instructions.Stage) error {
	return func(is []instructions.Stage) error {
		var errs error
		for _, s := range is {
			for _, sc := range []stageCheck{
				stageCheckNoAlpine,
			} {
				if err := sc.check(s); err != nil {
					label := fmt.Sprintf("%s:%d: %s", dockerfile, s.Location[0].Start.Line, sc.name)
					errs = errors.Append(errs, errors.Wrap(err, label))
				}
				for _, c := range s.Commands {
					for _, cc := range []commandCheck{
						commandCheckApkAdd,
					} {
						if err := cc.check(c); err != nil {
							label := fmt.Sprintf("%s:%d: %s", dockerfile, c.Location()[0].Start.Line, cc.name)
							errs = errors.Append(errs, errors.Wrap(err, label))
						}
					}
				}
			}
		}
		return errs
	}
}
