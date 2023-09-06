DROP TABLE IF EXISTS github_user_repos;

DROP TABLE IF EXISTS github_user_teams;

DROP TABLE IF EXISTS github_user_orgs;

DROP TRIGGER IF EXISTS delete_repo_org ON github_repo_user;

DROP FUNCTION IF EXISTS delete_repo_org();

DROP TRIGGER IF EXISTS delete_repo_user ON github_repo_org;

DROP FUNCTION IF EXISTS delete_repo_user();

DROP TABLE IF EXISTS github_repo_user;

DROP TABLE IF EXISTS github_repo_teams;

DROP TABLE IF EXISTS github_repo_org;

DROP TABLE IF EXISTS github_team_details;

DROP TABLE IF EXISTS github_org_details;
