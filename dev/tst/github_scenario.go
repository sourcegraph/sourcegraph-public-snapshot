package tst

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubScenarioBuilder struct {
	test     *testing.T
	client   *GitHubClient
	store    *ScenarioStore
	actions  *actionRunner
	reporter Reporter
	vars     *githubScenarioVars
}

type githubScenarioVars struct {
	org   *GitHubScenarioOrg
	users map[string]*GitHubScenarioUser
	teams map[string]*GitHubScenarioTeam
	repos map[string]*GitHubScenarioRepo
}

type GitHubScenario struct {
	client *GitHubClient
	org    *github.Organization
	users  map[string]*github.User
	teams  map[string]*github.Team
	repos  map[string]*github.Repository
}

func NewGitHubScenario(ctx context.Context, cfg *config.Config, t *testing.T) (*GitHubScenarioBuilder, error) {
	client, err := NewGitHubClient(ctx, cfg.GitHub)
	if err != nil {
		return nil, err
	}
	return &GitHubScenarioBuilder{
		test:     t,
		client:   client,
		store:    NewStore(t),
		actions:  NewActionManager(t),
		reporter: &NoopReporter{},
		vars:     newGitHubScenarioVars(t),
	}, nil
}

func (sb *GitHubScenarioBuilder) T(t *testing.T) *GitHubScenarioBuilder {
	sb.test = t
	sb.actions.T = t
	return sb
}

func (sb *GitHubScenarioBuilder) Verbose() *GitHubScenarioBuilder {
	sb.reporter = &ConsoleReporter{}
	sb.actions.Reporter = sb.reporter
	return sb
}

func (sb *GitHubScenarioBuilder) Quiet() *GitHubScenarioBuilder {
	sb.reporter = &NoopReporter{}
	sb.actions.Reporter = sb.reporter
	return sb
}

func (sb *GitHubScenarioBuilder) Org(name string) *GitHubScenarioBuilder {
	sb.test.Helper()
	org := NewGitHubScenarioOrg(name)
	// we need to keep track of the org for later usage in constructing the scenario
	sb.vars.org = org

	sb.actions.AddSetup(org.CreateOrgAction(sb.client), org.UpdateOrgPermissionsAction(sb.client))
	sb.actions.AddTeardown(org.DeleteOrgAction(sb.client))
	return sb
}

func (sb *GitHubScenarioBuilder) Users(users ...GitHubScenarioUser) *GitHubScenarioBuilder {
	sb.test.Helper()
	// keep track of the provided users for later use in constructing the scenario
	sb.vars.AddUsers(users...)
	for _, u := range users {
		if u == Admin {
			println("🚨", u.Name(), Admin.Name(), u == Admin)
			sb.actions.AddSetup(u.GetUserAction(sb.client))
		} else {
			sb.actions.AddSetup(u.CreateUserAction(sb.client))
			sb.actions.AddTeardown(u.DeleteUserAction(sb.client))
		}
	}
	return sb
}

func Team(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	return NewGitHubScenarioTeam(name, u...)
}

func (sb *GitHubScenarioBuilder) Teams(teams ...*GitHubScenarioTeam) *GitHubScenarioBuilder {
	sb.test.Helper()
	sb.vars.AddTeams(teams...)
	for _, t := range teams {
		sb.actions.AddSetup(t.CreateTeamAction(sb.client), t.AssignTeamAction(sb.client))
		sb.actions.AddTeardown(t.DeleteTeamAction(sb.client))
	}

	return sb
}

func (sb *GitHubScenarioBuilder) Repos(repos ...*GitHubScenarioRepo) *GitHubScenarioBuilder {
	sb.test.Helper()
	sb.vars.AddRepos(repos...)
	for _, r := range repos {
		if r.fork {
			sb.actions.AddSetup(r.ForkRepoAction(sb.client), r.GetRepoAction(sb.client))
			// Seems like you can't change permissions for a repo fork
			//sb.setupActions = append(sb.setupActions, r.SetPermissionsAction(sb.client))
			sb.actions.AddTeardown(r.DeleteRepoAction(sb.client))
		} else {
			sb.actions.AddSetup(r.NewRepoAction(sb.client),
				r.GetRepoAction(sb.client),
				r.InitLocalRepoAction(sb.client),
				r.SetPermissionsAction(sb.client),
			)

			sb.actions.AddTeardown(r.DeleteRepoAction(sb.client))
		}
		sb.actions.AddSetup(r.AssignTeamAction(sb.client))
	}

	return sb
}

func PublicRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, false)
}

func PrivateRepo(name string, team string, fork bool) *GitHubScenarioRepo {
	return NewGitHubScenarioRepo(name, team, fork, true)
}

func (sb *GitHubScenarioBuilder) Setup(ctx context.Context) (*GitHubScenario, func(context.Context) error, error) {
	sb.test.Helper()
	sb.reporter.Writeln("-- Setup --")
	start := time.Now().UTC()
	err := sb.actions.Apply(ctx, sb.store, sb.actions.setup, false)
	sb.reporter.Writef("Run complete: %s\n", time.Now().UTC().Sub(start))

	var scenario *GitHubScenario
	if err == nil {
		scenario, err = sb.buildScenario()
		return scenario, sb.TearDown, err
	}
	return scenario, sb.TearDown, err
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

func (sb *GitHubScenarioBuilder) buildScenario() (*GitHubScenario, error) {
	sb.test.Helper()
	scenario := createScenario()

	scenario.client = sb.client

	org, err := sb.store.GetOrg()
	if err != nil {
		return nil, err
	}
	scenario.org = org

	for _, user := range sb.vars.users {
		ghUser, err := sb.store.GetScenarioUser(*user)
		if err != nil {
			return nil, err
		}
		scenario.users[user.Name()] = ghUser
	}

	for _, team := range sb.vars.teams {
		ghTeam, err := sb.store.GetTeam(team)
		if err != nil {
			return nil, err
		}
		scenario.teams[team.Name()] = ghTeam
	}

	for _, repo := range sb.vars.repos {
		ghRepo, err := sb.store.GetRepo(repo)
		if err != nil {
			return nil, err
		}
		scenario.repos[repo.Name()] = ghRepo
	}

	return scenario, nil
}

func newGitHubScenarioVars(t *testing.T) *githubScenarioVars {
	t.Helper()
	return &githubScenarioVars{
		users: make(map[string]*GitHubScenarioUser),
		teams: make(map[string]*GitHubScenarioTeam),
		repos: make(map[string]*GitHubScenarioRepo),
	}
}

func (v *githubScenarioVars) AddUsers(users ...GitHubScenarioUser) {
	for _, user := range users {
		// we use the name of the user, because we want to replace users with similar names
		// and using ID or Key would mean we add new ones since those values are unique
		v.users[user.Name()] = &user
	}
}

func (v *githubScenarioVars) AddTeams(teams ...*GitHubScenarioTeam) {
	for _, team := range teams {
		v.teams[team.Name()] = team
	}
}

func (v *githubScenarioVars) AddRepos(repos ...*GitHubScenarioRepo) {
	for _, repo := range repos {
		v.repos[repo.Name()] = repo
	}
}

func createScenario() *GitHubScenario {
	return &GitHubScenario{
		org:   nil,
		users: make(map[string]*github.User),
		teams: make(map[string]*github.Team),
		repos: make(map[string]*github.Repository),
	}
}

func (s *GitHubScenario) GetOrg() (*github.Organization, error) {
	if s.org == nil {
		return nil, errors.Newf("org is nil")
	}

	return s.org, nil
}

func (s *GitHubScenario) GetClient() (*GitHubClient, error) {
	if s.client == nil {
		return nil, errors.Newf("client is nil")
	}

	return s.client, nil
}

func (s *GitHubScenario) Users() []*github.User {
	return mapValues(s.users)
}

func (s *GitHubScenario) Teams() []*github.Team {
	return mapValues(s.teams)
}

func (s *GitHubScenario) Repos() []*github.Repository {
	return mapValues(s.repos)
}

func (s *GitHubScenario) GetUser(user *GitHubScenarioUser) (*github.User, error) {
	if v, ok := s.users[user.Name()]; ok {
		return v, nil
	}

	return nil, errors.Newf("user '%q' not found", user.Name())
}

func (s *GitHubScenario) GetTeam(team *GitHubScenarioTeam) (*github.Team, error) {
	if v, ok := s.teams[team.Name()]; ok {
		return v, nil
	}

	return nil, errors.Newf("team '%q' not found", team.Name())
}

func (s *GitHubScenario) GetRepo(repo *GitHubScenarioRepo) (*github.Repository, error) {
	if v, ok := s.repos[repo.Name()]; ok {
		return v, nil
	}

	return nil, errors.Newf("repo '%q' not found", repo.Name())
}
