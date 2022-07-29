DROP TABLE IF EXISTS repo_statistics;

DROP TRIGGER IF EXISTS trig_count_soft_deleted_repo ON repo;
DROP FUNCTION IF EXISTS count_soft_deleted_repo();

DROP TRIGGER IF EXISTS trig_count_inserted_repo ON repo;
DROP FUNCTION IF EXISTS count_inserted_repo();

DROP TRIGGER IF EXISTS trig_count_deleted_repo ON repo;
DROP FUNCTION IF EXISTS count_deleted_repo();

DROP TRIGGER IF EXISTS trig_count_cloned_gitserver_repos ON gitserver_repos;
DROP FUNCTION IF EXISTS count_cloned_gitserver_repos();

DROP TRIGGER IF EXISTS trig_count_deleted_gitserver_repos ON gitserver_repos;
DROP FUNCTION IF EXISTS count_deleted_gitserver_repos();
