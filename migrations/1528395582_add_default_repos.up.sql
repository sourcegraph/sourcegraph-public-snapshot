BEGIN;
CREATE TABLE default_repos (
	repo_id INT REFERENCES repo(id)
);
COMMIT;
