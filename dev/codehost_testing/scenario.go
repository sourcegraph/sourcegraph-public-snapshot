package codehost_testing

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/codehost_testing/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ActionFn func(context.Context) error

// Action represents a task that is performed that cosists of task which creates or alters some resource with Apply
// and a corresponding Teardown task that destroys the resource created with Apply or undos any alterations. Effectively,
// Teardown should do the inverse of Apply.
//
// A Nil Teardown means there is no teardown to be performed, whereas Apply should never be nil
type Action struct {
	Name     string
	Apply    ActionFn
	Teardown ActionFn
}

// Scenario is an interface for executing a sequence of actions in a test scenario. Actions can be added
// by the relevant struct implementing the interface.
//
// The methods of the interface have the following intentions:
// Apply: should apply the actions that are part of the interface
// Teardown: should remove or undo the actions applied by Apply
// Plan: should return a human-readable string describing the actions that will be performed
type Scenario interface {
	Append(a ...*Action)
	Plan() string
	Apply(ctx context.Context) error
	Teardown(ctx context.Context) error
}

// GitHubScenario implements the Scenario interface for testing GitHub functionality. At its base GitHubScenario
// provides two top level methods to create GitHub resources namely:
// * create GitHub Organization, which returns a codehost_testing Org
// * create a GitHub User, which returns a codehost_testing User
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
	id            string
	t             *testing.T
	client        *GitHubClient
	actions       []*Action
	reporter      Reporter
	nextActionIdx int
	adminUser     *User
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
	scenario := &GitHubScenario{
		id:       id,
		t:        t,
		client:   client,
		actions:  make([]*Action, 0),
		reporter: &NoopReporter{},
	}
	scenario.adminUser = &User{
		s:    scenario,
		name: cfg.GitHub.AdminUser,
	}

	return scenario, err
}

// Verbose sets the reporter to ConsoleReporter to enable verbose output
func (s *GitHubScenario) SetVerbose() {
	s.reporter = &ConsoleReporter{}
}

// Quiet sets the reporter to a no-op reporter to reduce output
func (s *GitHubScenario) SetQuiet() {
	s.reporter = &NoopReporter{}
}

func (s *GitHubScenario) Append(actions ...*Action) {
	s.actions = append(s.actions, actions...)
}

// Plan returns a string describing the actions that will be performed
func (s *GitHubScenario) Plan() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Scenario %q\n", s.id)
	sb.WriteString("== Setup ==\n")
	for _, action := range s.actions {
		fmt.Fprintf(sb, "- %s\n", action.Name)
	}
	sb.WriteString("== Teardown ==\n")
	for _, action := range reverse(s.actions) {
		if action.Teardown == nil {
			continue
		}
		fmt.Fprintf(sb, "- %s\n", action.Name)
	}
	return sb.String()
}

// IsApplied returns whether Apply has already been called on this scenario. If more actions
// have been added since the last Apply(), it will return false.
func (s *GitHubScenario) IsApplied() bool {
	return s.nextActionIdx >= len(s.actions)
}

// Apply performs all the actions that have been added to this scenario sequentially in the order they were added.
// Furthemore cleanup function is registered so Teardown is called even if Apply fails to make sure we cleanup any
// left over resources due to a half applied scenario.
//
// Note that calling Apply more than once and with no new actions added, will result in an error be returned. This
// is done since duplicate resources cannot be created.
//
// If a scenario is applied and fails midway and Apply is called again, it will continue where it left off. The only
// way to reset this behaviour is to call teardown.
//
// Finally, if any action fails, no further actions will be executed and this method will return with the error
func (s *GitHubScenario) Apply(ctx context.Context) error {
	s.t.Helper()
	s.t.Cleanup(func() { s.Teardown(ctx) })
	var errs errors.MultiError
	setup := s.actions
	failFast := true

	if s.nextActionIdx >= len(s.actions) {
		return errors.New("all actions already applied")
	}

	start := time.Now()
	for currActionIdx, action := range setup {
		now := time.Now().UTC()

		if s.nextActionIdx > currActionIdx {
			s.reporter.Writef("(Setup) Skipping [%-50s]\n", action.Name)
			continue
		}

		if action.Apply == nil {
			return errors.Newf("action %q has nil Apply", action.Name)
		}

		s.reporter.Writef("(Setup) Applying [%-50s] ", action.Name)
		err := action.Apply(ctx)
		duration := time.Now().UTC().Sub(now)

		if err != nil {
			errs = errors.Append(errs, err)
			s.reporter.Writef("FAILED (%s)\n", duration.String())
			if failFast {
				break
			}
		} else {
			s.nextActionIdx++
			s.reporter.Writef("SUCCESS (%s)\n", duration.String())
		}
	}

	s.reporter.Writef("Setup complete in %s\n\n", time.Now().UTC().Sub(start))
	return errs
}

// Teardown cleans up any resources created by Apply. This method is automatically registered with *testing.Cleanup to
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
		// Nil means this action has no means of being torn down
		if action.Teardown == nil {
			continue
		}
		now := time.Now().UTC()

		s.reporter.Writef("(Teardown) Applying [%-50s] ", action.Name)
		err := action.Teardown(ctx)
		duration := time.Now().UTC().Sub(now)

		if err != nil {
			s.reporter.Writef("FAILED (%s)\n", duration.String())
			if failFast {
				break
			}
			errs = errors.Append(errs, err)
		} else {
			s.reporter.Writef("SUCCESS (%s)\n", duration.String())
		}
	}
	// Actions create new resources, therefore we can safely set the nextActionIdx to 0 here
	// so that new resources be created
	s.nextActionIdx = 0
	s.reporter.Writef("Teardown complete in %s\n", time.Now().UTC().Sub(start))
	return errs
}

func (s *GitHubScenario) CreateOrg(name string) *Org {
	baseOrg := &Org{
		s:    s,
		name: name,
	}

	createOrg := &Action{
		Name: "org:create:" + name,
		Apply: func(ctx context.Context) error {
			orgName := fmt.Sprintf("org-%s-%s", name, s.id)
			org, err := s.client.CreateOrg(ctx, orgName)
			if err != nil {
				return err
			}
			baseOrg.name = org.GetLogin()
			return nil
		},
		Teardown: func(context.Context) error {
			host := baseOrg.s.client.cfg.URL
			deleteURL := fmt.Sprintf("%s/organizations/%s/settings/profile", host, baseOrg.name)
			fmt.Printf("Visit %q to delete the org\n", deleteURL)
			return nil
		},
	}

	s.Append(createOrg)
	return baseOrg
}

// CreateUser adds an action to the scenario that will create a GitHub user with the given name. The username of the
// user will have the following format `user-{name}-{scenario id}` and email `test-user-e2e@sourcegraph.com`.
func (s *GitHubScenario) CreateUser(name string) *User {
	baseUser := &User{
		s:    s,
		name: name,
	}

	createUser := &Action{
		Name: "user:create:" + name,
		Apply: func(ctx context.Context) error {
			name := fmt.Sprintf("user-%s-%s", name, s.id)
			emailID := md5.Sum([]byte(s.id + time.Now().String()))
			email := fmt.Sprintf("test-user-e2e-%s@sourcegraph.com", hex.EncodeToString(emailID[:]))
			user, err := s.client.CreateUser(ctx, name, email)
			if err != nil {
				return err
			}

			baseUser.name = user.GetLogin()
			return nil
		},
		Teardown: func(ctx context.Context) error {
			return s.client.DeleteUser(ctx, baseUser.name)
		},
	}

	s.Append(createUser)
	return baseUser
}

// GetAdmin returns a User representing the GitHub admin user configured in the client.
//
// NOTE: this method does not actually add an explicit action to the scenario, but will still
// require that the scenario has been applied before the admin user can be retrieved - even though
// it is not strictly required as the Admin already exists.
func (s *GitHubScenario) GetAdmin() *User {
	return s.adminUser
}
