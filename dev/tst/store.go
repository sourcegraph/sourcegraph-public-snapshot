package tst

import (
	"testing"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var castFailure error

//TODO(burmudar): The store needs to be reworked entirely. Initially it was thought of that the
// store would be populated by what the actions return. This didn't pan out as some actions require
// the output of other actions and need use their output. A classic example is the CreateOrg action.
// All child actions need to refer to the org and the org needs to have been created.
//
// The store currently knows far to much about what it is storing and then the key management is also
// just a nightmare. We'd be far better off declaring the exact variable we want instead of wondering
// under what key something is.

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
	var result *github.Organization
	if v, ok := s.store["org"]; ok {
		if t, ok := v.(*github.Organization); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", "org")
	}
	return result, nil
}

func (s *ScenarioStore) SetScenarioUserMapping(u *GitHubScenarioUser, user *github.User) {
	s.T.Helper()
	s.store[u.Key()] = user
}

func (s *ScenarioStore) SetUsers(users []*github.User) {
	s.T.Helper()
	s.store["all-users"] = users
}

func (s *ScenarioStore) GetUsers() ([]*github.User, error) {
	s.T.Helper()
	var result []*github.User
	if v, ok := s.store["org"]; ok {
		if t, ok := v.([]*github.User); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", "all-users")
	}
	return result, nil
}

func (s *ScenarioStore) GetScenarioUser(u GitHubScenarioUser) (*github.User, error) {
	s.T.Helper()
	var result *github.User
	if v, ok := s.store[u.Key()]; ok {
		if t, ok := v.(*github.User); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", u.Key())
	}
	return result, nil
}

func (s *ScenarioStore) SetTeam(gt *GitHubScenarioTeam, t *github.Team) {
	s.T.Helper()
	// Store it twice so we have two ways of retrieving a team
	s.store[gt.Name()] = t
	s.store[gt.Key()] = t
}

func (s *ScenarioStore) GetTeamByName(name string) (*github.Team, error) {
	s.T.Helper()
	var result *github.Team
	if v, ok := s.store[name]; ok {
		if t, ok := v.(*github.Team); ok {
			result = t
		} else {
			return result, castFailure
		}
	} else {
		return result, errors.Newf("%s not found - it might not have been loaded yet", name)
	}
	return result, nil
}

func (s *ScenarioStore) GetTeam(t *GitHubScenarioTeam) (*github.Team, error) {
	s.T.Helper()
	return s.GetTeamByName(t.Name())
}

func (s *ScenarioStore) SetRepo(r *GitHubScenarioRepo, repo *github.Repository) {
	s.T.Helper()
}

func (s *ScenarioStore) GetRepo(r *GitHubScenarioRepo) (*github.Repository, error) {
	s.T.Helper()
	// stub
	return nil, nil
}
