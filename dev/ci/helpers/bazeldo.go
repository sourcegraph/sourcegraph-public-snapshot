package helpers

import (
	"strings"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var allowedBazelFlags = map[string]struct{}{
	"--runs_per_test":        {},
	"--nobuild":              {},
	"--local_test_jobs":      {},
	"--test_arg":             {},
	"--nocache_test_results": {},
	"--test_tag_filters":     {},
	"--test_timeout":         {},
	"--config":               {},
	"--test_output":          {},
	"--verbose_failures":     {},
}

var bazelFlagsRe = regexp.MustCompile(`--\w+`)

// VerifyBazelDoCommand checks that the given command is allowed
// to run as a bazel-do runtype.
func VerifyBazelCommand(command string) error {
	// check for shell escape mechanisms.
	bannedChars := []string{"`", "$", "(", ")", ";", "&", "|", "<", ">"}
	for _, c := range bannedChars {
		if strings.Contains(command, c) {
			return errors.Newf("unauthorized input for bazel command: %q", c)
		}
	}

	// check for command and targets
	strs := strings.Split(command, " ")
	if len(strs) < 2 {
		return errors.New("invalid command")
	}

	// command must be either build or test.
	switch strs[0] {
	case "build":
	case "test":
	default:
		return errors.Newf("disallowed bazel command: %q", strs[0])
	}

	// need at least one target.
	if !strings.HasPrefix(strs[1], "//") {
		return errors.New("misconstructed command, need at least one target")
	}

	// ensure flags are in the allow-list.
	matches := bazelFlagsRe.FindAllString(command, -1)
	for _, m := range matches {
		if _, ok := allowedBazelFlags[m]; !ok {
			return errors.Newf("disallowed bazel flag: %q", m)
		}
	}
	return nil
}
