DELETE FROM global_dep WHERE repo_id IN (
	SELECT DISTINCT(global_dep.repo_id) FROM global_dep LEFT JOIN repo ON (global_dep.repo_id=repo.id) WHERE repo.id IS NULL
);
ALTER TABLE global_dep ADD CONSTRAINT global_dep_repo_id FOREIGN KEY (repo_id) REFERENCES repo (id) ON DELETE RESTRICT;

DELETE FROM pkgs WHERE repo_id IN (
	SELECT DISTINCT(pkgs.repo_id) FROM pkgs LEFT JOIN repo ON (pkgs.repo_id=repo.id) WHERE repo.id IS NULL
);
ALTER TABLE pkgs ADD CONSTRAINT pkgs_repo_id FOREIGN KEY (repo_id) REFERENCES repo (id) ON DELETE RESTRICT;
