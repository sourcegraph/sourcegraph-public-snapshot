package codehost_testing

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/codehost_testing/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type action struct {
	name     string
	apply    func(context.Context) error
	teardown func(context.Context) error
}

// Scenario is an interface for executing a sequence of actions in a test scenario. Actions can be added
// by the relevant struct implementing the interface.
//
// The methods of the interface have the following intentions:
// Apply: should apply the actions that are part of the interface
// Teardown: should remove or undo the actions applied by Apply
// Plan: should return a human-readable string describing the actions that will be performed
type Scenario interface {
	append(a ...*action)
	Plan() string
	Apply(ctx context.Context) error
	Teardown(ctx context.Context) error
}

// GitHubScenario implements the Scenario interface for testing GitHub functionality. At its base GitHubScenario
// provides two top level methods to create GitHub resources namely:
// * create GitHub Organization, which returns a codehost_scenario Org
// * create a GitHub User, which returns a codehost_scenario User
//
// Further resources can be created by calling methods on the returned Org or User. For instance, since a repository
// is tied to an organization, one can call org.CreateRepo, which will add an action for a repo to be created in the
// organization.
//
// Calling any action creating method does not immediately create the resource in GitHub. Instead a action is added
// the list of actions contained in this scenario. Only once Apply() has been called on the Scenario itself will
// the resources be created on GitHub.
//
// Once Apply() is called, all the corresponding resources should be realized on GitHub. To fetch the corresponding
// GitHub resources once can call Get() on the resources.
type GitHubScenario struct {
	id               string
	t                *testing.T
	client           *GitHubClient
	actions          []*action
	reporter         Reporter
	appliedActionIdx int
}

var _ Scenario = (*GitHubScenario)(nil)

// NewGitHubScenario creates a new GitHubScenario instance. A base64 ID will be generated to identify this scenario.
// This ID will also be used to uniquely identify any resources created as part of this scenario.
//
// By default a GitHubScenario is created with a NoopReporter. To have more verbose output, call SetVerbose() on the scenario,
// and to reduce the output, call SetQuiet().
func NewGitHubScenario(t *testing.T, cfg config.Config) (*GitHubScenario, error) {
	client, err := NewGitHubClient(t, cfg.GitHub)
	if err != nil {
		return nil, err
	}
	uid := []byte(uuid.NewString())
	id := base64.RawStdEncoding.EncodeToString(uid[:])[:10]
	return &GitHubScenario{
		id:       id,
		t:        t,
		client:   client,
		actions:  make([]*action, 0),
		reporter: NoopReporter{},
	}, nil
}

// Verbose sets the reporter to ConsoleReporter to enable verbose output
func (s *GitHubScenario) SetVerbose() {
	s.reporter = &ConsoleReporter{}
}

// Quiet sets the reporter to a no-op reporter to reduce output
func (s *GitHubScenario) SetQuiet() {
	s.reporter = NoopReporter{}
}

func (s *GitHubScenario) append(actions ...*action) {
	s.actions = append(s.actions, actions...)
}

// Plan returns a string describing the actions that will be performed
func (s *GitHubScenario) Plan() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Scenario %q\n", s.id)
	sb.WriteString("== Setup ==\n")
	for _, action := range s.actions {
		fmt.Fprintf(sb, "- %s\n", action.name)
	}
	sb.WriteString("== Teardown ==\n")
	for _, action := range reverse(s.actions) {
		if action.teardown == nil {
			continue
		}
		fmt.Fprintf(sb, "- %s\n", action.name)
	}
	return sb.String()
}

// IsApplied returns whether Apply has already been called on this scenario. If more actions
// have been added since the last Apply(), it will return false.
func (s *GitHubScenario) IsApplied() bool {
	return s.appliedActionIdx >= len(s.actions)
}

// Apply performs all the actions that have been added to this scenario sequentially in the order they were added.
// Furthemore cleanup function is registered so Teardown is called even if Apply fails to make sure we cleanup any
// left over resources due to a half applied scenario.
//
// Note that calling Apply more than once and with no new actions added, will result in an error be returned. This
// is done since duplicate resources cannot be created.
//
// Finally, if any action fails, no further actions will be executed and this method will return with the error
func (s *GitHubScenario) Apply(ctx context.Context) error {
	s.t.Helper()
	s.t.Cleanup(func() { s.Teardown(ctx) })
	var errs errors.MultiError
	setup := s.actions
	failFast := true

	if s.appliedActionIdx >= len(s.actions) {
		return errors.New("all actions already applied")
	}

	start := time.Now()
	for i, action := range setup {
		now := time.Now().UTC()

		var err error
		if s.appliedActionIdx <= i {
			s.reporter.Writef("(Setup) Applying [%-50s] ", action.name)
			err = action.apply(ctx)
			s.appliedActionIdx++
		} else {
			s.reporter.Writef("(Setup) Skipping [%-50s]\n", action.name)
			continue
		}

		duration := time.Now().UTC().Sub(now)
		if err != nil {
			s.reporter.Writef("FAILED (%s)\n", duration.String())
			if failFast {
				return err
			}
			errs = errors.Append(errs, err)
		} else {
			s.reporter.Writef("SUCCESS (%s)\n", duration.String())
		}
	}

	s.reporter.Writef("Setup complete in %s\n\n", time.Now().UTC().Sub(start))
	return errs
}

// Teardown cleans up any resources created by Apply. This method is automatically registerd with *testing.Cleanup to
// cleanup resources, so generally it would not have to be called explicitly.
//
// Teardown iterates through the scenario actions in reverse order, calling teardown on each action. If a action
// has a nil teardown function it will be skipped. Teardown does not stop iterating when an action returns with an error,
// instead the error is accumulated and the next teardown action is executed.
//
// Note that Teardown is not idempotent. Multiple calls will result in failures.
func (s *GitHubScenario) Teardown(ctx context.Context) error {
	s.t.Helper()
	var errs errors.MultiError
	teardown := reverse(s.actions)
	failFast := false

	start := time.Now()
	for _, action := range teardown {
		s.appliedActionIdx--
		if action.teardown == nil {
			continue
		}
		now := time.Now().UTC()

		s.reporter.Writef("(Teardown) Applying [%-50s] ", action.name)
		err := action.teardown(ctx)
		duration := time.Now().UTC().Sub(now)

		if err != nil {
			s.reporter.Writef("FAILED (%s)\n", duration.String())
			if failFast {
				return err
			}
			errs = errors.Append(errs, err)
		} else {
			s.reporter.Writef("SUCCESS (%s)\n", duration.String())
		}
	}
	if s.appliedActionIdx < 0 {
		s.t.Logf("scenario applied action Idx went negative. This is almost certainly a bug")
		s.appliedActionIdx = 0
	}
	s.reporter.Writef("Teardown complete in %s\n", time.Now().UTC().Sub(start))
	return errs
}
