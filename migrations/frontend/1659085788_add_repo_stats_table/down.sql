DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_update ON repo;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_insert ON repo;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_delete ON repo;

DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_update();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_insert();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_delete();

DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_update ON gitserver_repos;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_insert ON gitserver_repos;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_delete ON gitserver_repos;

DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_update();
DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_insert();
DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_delete();

DROP TABLE IF EXISTS gitserver_repos_statistics;
DROP TABLE IF EXISTS repo_statistics;
