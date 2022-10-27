DROP TRIGGER IF EXISTS trig_create_zoekt_repo_on_repo_insert ON repo;
DROP FUNCTION IF EXISTS func_insert_zoekt_repo();

DROP TABLE IF EXISTS zoekt_repos;
