package githubdb

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type Store interface {
	basestore.ShareableStore
	SetOrgDetails(ctx context.Context, baseURL string, orgID int, defaultRepoPerm string) error
	SetTeamDetails(ctx context.Context, baseURL string, teamID int, orgID int) error
	SetTeamDetailsWithParent(ctx context.Context, baseURL string, teamID, parentID, orgID int) error
	SetRepoOrg(ctx context.Context, baseURL string, repoID string, orgID int) error
	SetRepoTeam(ctx context.Context, baseURL string, repoID string, teamID int) error
	SetRepoTeams(ctx context.Context, baseURL string, repoID string, teamIDs []int) error
	SetRepoUser(ctx context.Context, baseURL string, repoID string, userID string) error
	SetUserOrg(ctx context.Context, baseURL string, userID string, orgID int) error
	SetUserTeam(ctx context.Context, baseURL string, userID string, teamID int) error
	SetUserOrgs(ctx context.Context, baseURL string, userID string, orgIDs []int) error
	SetUserTeams(ctx context.Context, baseURL string, userID string, teamIDs []int) error
	SetUserRepos(ctx context.Context, baseURL string, userID string, repoIDs []string) error
}

type store struct {
	*basestore.Store
}

// SetOrgDefaultPermission implements Store.
func (s *store) SetOrgDetails(ctx context.Context, baseURL string, orgID int, defaultRepoPerm string) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_org_details (base_url, org_id, default_repository_permission)
	VALUES (%s, %s, %s)
	ON CONFLICT (base_url, org_id) DO UPDATE
	SET default_repository_permission = EXCLUDED.default_repository_permission`, baseURL, orgID, defaultRepoPerm)

	return s.Exec(ctx, qry)
}

func (s *store) SetTeamDetails(ctx context.Context, baseURL string, teamID int, orgID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_team_details (base_url, team_id, org_id)
	VALUES (%s, %s, %s)
	ON CONFLICT (base_url, team_id) DO UPDATE
	SET parent_id = NULL, org_id = EXCLUDED.org_id`, baseURL, teamID, orgID)

	return s.Exec(ctx, qry)
}

func (s *store) SetTeamDetailsWithParent(ctx context.Context, baseURL string, teamID, parentID, orgID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_team_details (base_url, team_id, parent_id, org_id)
	VALUES (%s, %s, %s, %s)
	ON CONFLICT (base_url, team_id) DO UPDATE
	SET parent_id = EXCLUDED.parent_id, org_id = EXCLUDED.org_id`, baseURL, teamID, parentID, orgID)

	return s.Exec(ctx, qry)
}

// SetRepoOrg implements Store.
func (s *store) SetRepoOrg(ctx context.Context, baseURL string, repoID string, orgID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_repo_org (base_url, repo_id, org_id)
	VALUES (%s, %s, %s)
	ON CONFLICT (base_url, repo_id) DO UPDATE
	SET org_id = EXCLUDED.org_id`, baseURL, repoID, orgID)

	return s.Exec(ctx, qry)
}

// SetRepoTeam implements Store.
func (s *store) SetRepoTeam(ctx context.Context, baseURL string, repoID string, teamID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_repo_teams (base_url, repo_id, team_id)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING`, baseURL, repoID, teamID)

	return s.Exec(ctx, qry)
}

// SetRepoUser implements Store.
func (s *store) SetRepoUser(ctx context.Context, baseURL string, repoID string, userID string) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_repo_user (base_url, repo_id, user_id)
	VALUES (%s, %s, %s)
	ON CONFLICT (base_url, repo_id) DO UPDATE
	SET user_id = EXCLUDED.user_id`, baseURL, repoID, userID)

	return s.Exec(ctx, qry)
}

func (s *store) SetUserOrg(ctx context.Context, baseURL string, userID string, orgID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_user_orgs (base_url, user_id, org_id)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING`, baseURL, userID, orgID)

	return s.Exec(ctx, qry)
}

func (s *store) SetUserTeam(ctx context.Context, baseURL string, userID string, teamID int) error {
	qry := sqlf.Sprintf(`
	INSERT INTO github_user_teams (base_url, user_id, team_id)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING`, baseURL, userID, teamID)

	return s.Exec(ctx, qry)
}

func (s *store) SetUserOrgs(ctx context.Context, baseURL string, userID string, orgIDs []int) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(`
	DELETE FROM github_user_orgs
	WHERE base_url = %s AND user_id = %s`, baseURL, userID)); err != nil {
		return err
	}

	var values []*sqlf.Query
	for _, orgID := range orgIDs {
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", baseURL, userID, orgID))
	}
	qry := sqlf.Sprintf(`
	INSERT INTO github_user_orgs (base_url, user_id, org_id)
	VALUES %s
	ON CONFLICT DO NOTHING`, sqlf.Join(values, ", "))

	if err := tx.Exec(ctx, qry); err != nil {
		return err
	}

	return nil
}

func (s *store) SetUserTeams(ctx context.Context, baseURL string, userID string, teamIDs []int) error {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(`
	DELETE FROM github_user_teams
	WHERE base_url = %s AND user_id = %s`, baseURL, userID)); err != nil {
		return err
	}

	var values []*sqlf.Query
	for _, teamID := range teamIDs {
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", baseURL, userID, teamID))
	}

	qry := sqlf.Sprintf(`
	INSERT INTO github_user_teams (base_url, user_id, team_id)
	VALUES %s
	ON CONFLICT DO NOTHING`, sqlf.Join(values, ", "))

	if err := tx.Exec(ctx, qry); err != nil {
		return err
	}

	return nil
}

func (s *store) SetUserRepos(ctx context.Context, baseURL string, userID string, repoIDs []string) error {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(`
	DELETE FROM github_user_repos
	WHERE base_url = %s AND user_id = %s`, baseURL, userID)); err != nil {
		return err
	}

	var values []*sqlf.Query
	for _, repoID := range repoIDs {
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", baseURL, userID, repoID))
	}

	qry := sqlf.Sprintf(`
	INSERT INTO github_user_repos (base_url, user_id, repo_id)
	VALUES %s
	ON CONFLICT DO NOTHING`, sqlf.Join(values, ", "))

	if err := tx.Exec(ctx, qry); err != nil {
		return err
	}

	return nil
}

func (s *store) SetRepoTeams(ctx context.Context, baseURL string, repoID string, teamIDs []int) error {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(`
	DELETE FROM github_repo_teams
	WHERE base_url = %s AND repo_id = %s`, baseURL, repoID)); err != nil {
		return err
	}

	var values []*sqlf.Query
	for _, teamID := range teamIDs {
		values = append(values, sqlf.Sprintf("(%s, %s, %s)", baseURL, repoID, teamID))
	}

	qry := sqlf.Sprintf(`
	INSERT INTO github_repo_teams (base_url, repo_id, team_id)
	VALUES %s
	ON CONFLICT DO NOTHING`, sqlf.Join(values, ", "))

	if err := tx.Exec(ctx, qry); err != nil {
		return err
	}

	return nil
}

func NewStore(other basestore.ShareableStore) Store {
	return &store{Store: basestore.NewWithHandle(other.Handle())}
}
