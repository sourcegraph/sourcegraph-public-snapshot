package tst

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
)

// stub types
type GitHubScenarioUser struct{}
type GitHubScenarioTeam struct{}
type GitHubScenarioRepo struct{}
type GitHubClient struct{}

type GitHubScenarioBuilder struct {
	test     *testing.T
	client   *GitHubClient
	store    *ScenarioStore
	actions  *actionRunner
	reporter Reporter
}

type GitHubScenario struct {
	client *GitHubClient
	users  []*github.User
	teams  []*github.Team
	repos  []*github.Repository
	org    *github.Organization
}

func NewGitHubScenario(ctx context.Context, cfg *config.Config, t *testing.T) (*GitHubScenarioBuilder, error) {
	return &GitHubScenarioBuilder{
		test:     t,
		client:   nil,
		store:    NewStore(t),
		actions:  NewActionManager(t),
		reporter: NoopReporter{},
	}, nil
}

func (sb *GitHubScenarioBuilder) T(t *testing.T) *GitHubScenarioBuilder {
	sb.test = t
	sb.actions.T = t
	return sb
}

func (sb *GitHubScenarioBuilder) Verbose() {
	sb.reporter = ConsoleReporter{}
	sb.actions.Reporter = sb.reporter
}

func (sb *GitHubScenarioBuilder) Quiet() {
	sb.reporter = NoopReporter{}
	sb.actions.Reporter = sb.reporter
}

func (sb *GitHubScenarioBuilder) Org(name string) *GitHubScenarioBuilder {
	sb.test.Helper()
	// stub
	return sb
}

func (sb *GitHubScenarioBuilder) Users(users ...GitHubScenarioUser) *GitHubScenarioBuilder {
	sb.test.Helper()
	// stub
	return sb
}

func Team(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	// stub
	return nil
}

func (sb *GitHubScenarioBuilder) Teams(teams ...*GitHubScenarioTeam) *GitHubScenarioBuilder {
	sb.test.Helper()
	// stub
	return sb
}

func (sb *GitHubScenarioBuilder) Repos(repos ...*GitHubScenarioRepo) *GitHubScenarioBuilder {
	sb.test.Helper()

	// stub

	return sb
}

func PublicRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	// stub
	return nil
}

func PrivateRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	// stub
	return nil
}

func (sb *GitHubScenarioBuilder) Setup(ctx context.Context) (GitHubScenario, func(context.Context) error, error) {
	sb.test.Helper()
	sb.reporter.Writeln("-- Setup --")
	start := time.Now().UTC()
	err := sb.actions.Apply(ctx, sb.store, sb.actions.setup, false)
	sb.reporter.Writef("Run complete: %s\n", time.Now().UTC().Sub(start))
	return GitHubScenario{}, sb.TearDown, err
}

func (sb *GitHubScenarioBuilder) TearDown(ctx context.Context) error {
	sb.test.Helper()
	sb.reporter.Writeln("-- Teardown --")
	start := time.Now().UTC()
	err := sb.actions.Apply(ctx, sb.store, reverse(sb.actions.teardown), false)
	sb.reporter.Writef("Run complete: %s\n", time.Now().UTC().Sub(start))
	return err
}

func (sb *GitHubScenarioBuilder) String() string {
	return sb.actions.String()
}
