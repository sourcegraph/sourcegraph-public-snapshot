package docker

import (
	"strings"

	"github.com/grafana/regexp"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	illegalApkAddRegexp = regexp.MustCompile(`[^(\s)>]+[=|<][0-9]+[^\s]+`)
)

type commandCheck struct {
	name  string
	check func(c instructions.Command) error
}

// See .hadolint.yaml DL3018 docstring
var commandCheckApkAdd = commandCheck{
	name: "illegal apk add version requirement",
	check: func(c instructions.Command) error {
		if runCmd, okay := c.(*instructions.RunCommand); okay {
			run := runCmd.String()
			if strings.Contains(run, "apk add") {
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

// CheckCommand is a linter for Dockerfile commands.
func CheckCommand(c instructions.Command) error {
	var errs error
	for _, check := range []commandCheck{
		commandCheckApkAdd,
	} {
		if err := check.check(c); err != nil {
			errs = errors.Append(errs, errors.Wrap(err, check.name))
		}
	}
	return errs
}
