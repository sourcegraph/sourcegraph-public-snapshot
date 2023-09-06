CREATE TABLE IF NOT EXISTS github_org_details(
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    org_id INT NOT NULL,
    default_repository_permission TEXT NOT NULL,
    PRIMARY KEY(base_url, org_id)
);

CREATE TABLE IF NOT EXISTS github_team_details(
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    team_id INT NOT NULL,
    parent_id INT,
    org_id INT NOT NULL,
    PRIMARY KEY (base_url, team_id)
);

CREATE TABLE IF NOT EXISTS github_repo_org (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    repo_id TEXT NOT NULL,
    org_id INT NOT NULL,
    PRIMARY KEY (base_url, repo_id, org_id)
);

CREATE TABLE IF NOT EXISTS github_repo_teams (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    repo_id TEXT NOT NULL,
    team_id INT NOT NULL,
    PRIMARY KEY (base_url, repo_id, team_id)
);

CREATE TABLE IF NOT EXISTS github_repo_user (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    repo_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (base_url, repo_id)
);

CREATE OR REPLACE FUNCTION delete_repo_user()
RETURNS trigger AS $$
BEGIN
  DELETE FROM github_repo_user
  WHERE base_url = NEW.base_url AND repo_id = NEW.repo_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER delete_repo_user
AFTER INSERT ON github_repo_org
FOR EACH ROW
EXECUTE PROCEDURE delete_repo_user();

CREATE OR REPLACE FUNCTION delete_repo_org()
RETURNS trigger AS $$
BEGIN
  DELETE FROM github_repo_org
  WHERE base_url = NEW.base_url AND repo_id = NEW.repo_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER delete_repo_org
AFTER INSERT ON github_repo_user
FOR EACH ROW
EXECUTE PROCEDURE delete_repo_org();

CREATE TABLE IF NOT EXISTS github_user_orgs (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    user_id TEXT NOT NULL,
    org_id INT NOT NULL,
    PRIMARY KEY (base_url, user_id, org_id)
);

CREATE TABLE IF NOT EXISTS github_user_teams (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    user_id TEXT NOT NULL,
    team_id INT NOT NULL,
    PRIMARY KEY (base_url, user_id, team_id)
);

CREATE TABLE IF NOT EXISTS github_user_repos (
    base_url TEXT NOT NULL CHECK (right(base_url, 1) = '/'),
    user_id TEXT NOT NULL,
    repo_id TEXT NOT NULL,
    PRIMARY KEY (base_url, user_id, repo_id)
);
