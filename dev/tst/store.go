package tst

import (
	"testing"

	"github.com/google/go-github/v53/github"
)

var castFailure error

type ScenarioStore struct {
	T     *testing.T
	store map[string]any
}

func NewStore(t *testing.T) *ScenarioStore {
	return &ScenarioStore{
		T:     t,
		store: make(map[string]any),
	}
}

func (s *ScenarioStore) SetOrg(org *github.Organization) {
	s.T.Helper()
	s.store["org"] = org
}
func (s *ScenarioStore) GetOrg() (*github.Organization, error) {
	s.T.Helper()
	//stub
	return nil, nil
}

func (s *ScenarioStore) SetScenarioUserMapping(u *GitHubScenarioUser, user *github.User) {
	s.T.Helper()
}

func (s *ScenarioStore) SetUsers(users []*github.User) {
	s.T.Helper()
}

func (s *ScenarioStore) GetUsers() ([]*github.User, error) {
	s.T.Helper()
	//stub
	return nil, nil
}

func (s *ScenarioStore) GetScenarioUser(u GitHubScenarioUser) (*github.User, error) {
	s.T.Helper()
	//stub
	return nil, nil
}

func (s *ScenarioStore) SetTeam(gt *GitHubScenarioTeam, t *github.Team) {
	s.T.Helper()
}

func (s *ScenarioStore) GetTeamByName(name string) (*github.Team, error) {
	s.T.Helper()
	// stub
	return nil, nil
}

func (s *ScenarioStore) GetTeam(t *GitHubScenarioTeam) (*github.Team, error) {
	s.T.Helper()
	// stub
	return nil, nil
}

func (s *ScenarioStore) SetRepo(r *GitHubScenarioRepo, repo *github.Repository) {
	s.T.Helper()
}

func (s *ScenarioStore) GetRepo(r *GitHubScenarioRepo) (*github.Repository, error) {
	s.T.Helper()
	// stub
	return nil, nil
}
