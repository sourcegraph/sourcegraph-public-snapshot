package runner

import (
	"fmt"
	"path"
	"strings"

	"github.com/drone/drone-exec/parser"
	"github.com/drone/drone-plugin-go/plugin"
)

// isMatch is a helper function that returns true if
// all criteria is matched.
func isMatch(node *parser.FilterNode, s *State) (match bool) {

	var last string
	if s.BuildLast != nil {
		last = s.BuildLast.Status
	}

	switch {
	case !matchBranch(node.Branch, s.Build.Branch):
		return false
	case !matchMatrix(node.Matrix, s.Job.Environment):
		return false
	case !matchRepo(node.Repo, s.Repo.FullName):
		return false
	case !matchEvent(node.Event, s.Build.Event):
		return false
	}

	switch {
	case matchSuccess(node.Success, s.Job.Status):
		return true
	case matchFailure(node.Failure, s.Job.Status):
		return true
	case matchChange(node.Change, s.Job.Status, last):
		return true
	}

	return false
}

// matchBranch is a helper function that returns true
// if all_branches is true. Else it returns false if a
// branch condition is specified, and the branch does
// not match.
func matchBranch(want []string, got string) (match bool) {
	if len(want) == 0 {
		return true
	}
	if strings.HasPrefix(got, "refs/heads/") {
		got = got[11:]
	}
	for _, pattern := range want {
		if matchPath(pattern, got) {
			match = true
			break
		}
	}
	return
}

// matchRepo is a helper function that returns false
// if this task is only intended for a named repo,
// the current build does not match that repo.
//
// This is useful when you want to prevent forks from
// executing deployment, publish or notification steps.
func matchRepo(want, got string) bool {
	if len(want) == 0 {
		return true
	}
	return got == want
}

// matchEvent is a helper function that returns false
// if this task is only intended for a specific repository
// event not matched by the current build. For example,
// only executing a build for `tags` or `pull_requests`
func matchEvent(want []string, got string) (match bool) {
	if len(want) == 0 {
		return true
	}
	for _, want_ := range want {
		if got == want_ {
			match = true
			break
		}
	}
	return
}

// matchMatrix is a helper function that returns false
// to limit steps to only certain matrix axis.
func matchMatrix(want, got map[string]string) bool {
	if len(want) == 0 {
		return true
	}
	for k, v := range want {
		if got[k] != v {
			return false
		}
	}
	return true
}

func matchSuccess(toggle, status string) bool {
	ok, err := parseBool(toggle)
	if err != nil {
		return true
	}
	return ok && (status == plugin.StateSuccess || status == plugin.StateRunning)
}

func matchFailure(toggle, status string) bool {
	ok, err := parseBool(toggle)
	if err != nil {
		return true
	}
	return ok && status != plugin.StateSuccess && status != plugin.StateRunning
}

func matchChange(toggle, status, last string) bool {
	ok, err := parseBool(toggle)
	if err != nil {
		return true
	}
	switch status {
	case plugin.StateRunning:
		status = plugin.StateSuccess
	}
	return ok && status != last
}

func parseBool(str string) (value bool, err error) {
	switch str {
	case "true", "TRUE", "True", "On", "ON", "on":
		return true, nil
	case "false", "FALSE", "False", "Off", "off", "OFF":
		return false, nil
	}
	return false, fmt.Errorf("Error parsing boolean %s", str)
}

func matchPath(pattern, str string) bool {
	negate := strings.HasPrefix(pattern, "!")
	if negate {
		pattern = pattern[1:]
	}
	match, _ := path.Match(pattern, str)
	if negate {
		match = !match
	}
	return match
}
