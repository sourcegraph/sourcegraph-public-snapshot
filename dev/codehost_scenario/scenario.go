package codehost_scenario

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/codehost_scenario/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type action struct {
	name     string
	apply    func(context.Context) error
	teardown func(context.Context) error
}

type Scenario interface {
	append(a ...*action)
	Plan() string
	Apply(ctx context.Context) error
	Teardown(ctx context.Context) error
}

type GithubScenario struct {
	id               string
	t                *testing.T
	client           *GitHubClient
	actions          []*action
	reporter         Reporter
	appliedActionIdx int
}

var _ Scenario = (*GithubScenario)(nil)

func NewGithubScenario(ctx context.Context, t *testing.T, cfg config.Config) (*GithubScenario, error) {
	client, err := NewGitHubClient(ctx, cfg.GitHub)
	if err != nil {
		return nil, err
	}
	uid := []byte(uuid.NewString())
	id := base64.RawStdEncoding.EncodeToString(uid[:])[:10]
	return &GithubScenario{
		id:       id,
		t:        t,
		client:   client,
		actions:  make([]*action, 0),
		reporter: NoopReporter{},
	}, nil
}

func (s *GithubScenario) Verbose() {
	s.reporter = &ConsoleReporter{}
}

func (s *GithubScenario) Quiet() {
	s.reporter = NoopReporter{}
}

func (s *GithubScenario) append(actions ...*action) {
	s.actions = append(s.actions, actions...)
}

func (s *GithubScenario) Plan() string {
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

func (s *GithubScenario) IsApplied() bool {
	return s.appliedActionIdx >= len(s.actions)
}

func (s *GithubScenario) Apply(ctx context.Context) error {
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

func (s *GithubScenario) Teardown(ctx context.Context) error {
	s.t.Helper()
	var errs errors.MultiError
	teardown := reverse(s.actions)
	failFast := false

	start := time.Now()
	for _, action := range teardown {
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
	s.reporter.Writef("Teardown complete in %s\n", time.Now().UTC().Sub(start))
	return errs
}
