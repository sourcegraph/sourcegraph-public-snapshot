// Package sysreq implements checking for Sourcegraph system requirements.
package sysreq

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Status describes the status of a system requirement.
type Status struct {
	Name    string // the required component
	Problem string // if non-empty, a description of the problem
	Fix     string // if non-empty, how to fix the problem
	Err     error  // if non-nil, the error encountered
	Skipped bool   // if true, indicates this check was skipped
}

// Equals returns true if other has the same fields as the receiver.
// Used for testing as we don't want to DeepEqual or cmp.Diff structs
// holding error values.
func (s Status) Equals(other Status) bool {
	return s.Name == other.Name && s.Problem == other.Problem && s.Fix == other.Fix && errors.Is(s.Err, other.Err) && s.Skipped == other.Skipped
}

// OK is whether the component is present, has no errors, and was not
// skipped.
func (s *Status) OK() bool {
	return s.Problem == "" && s.Fix == "" && s.Err == nil && !s.Skipped
}

func (s *Status) Failed() bool { return s.Problem != "" || s.Err != nil }

// Check checks for the presence of system requirements, such as
// Docker and Git. The skip list contains case-insensitive names of
// requirement checks (such as "Docker" and "Git") that should be
// skipped.
func Check(ctx context.Context, skip []string) []Status {
	shouldSkip := func(name string) bool {
		for _, v := range skip {
			if strings.EqualFold(name, v) {
				return true
			}
		}
		return false
	}

	statuses := make([]Status, len(checks))
	for i, c := range checks {
		statuses[i].Name = c.Name

		if shouldSkip(c.Name) {
			statuses[i].Skipped = true
			continue
		}

		problem, fix, err := c.Check(ctx)
		if err != nil {
			statuses[i].Err = err
		}
		statuses[i].Problem = problem
		statuses[i].Fix = fix
	}

	return statuses
}

type check struct {
	Name  string
	Check CheckFunc
}

// CheckFunc is a function that checks for a system requirement. If
// any of problem, fix, or err are non-zero, then the system
// requirement check is deemed to have failed.
type CheckFunc func(context.Context) (problem, fix string, err error)

// AddCheck adds a new check that will be run when this package's
// Check func is called. It is used by other packages to specify
// system requirements.
func AddCheck(name string, fn CheckFunc) {
	checks = append(checks, check{name, fn})
}

var checks = []check{
	{
		Name:  "Rlimit",
		Check: rlimitCheck,
	},
}
