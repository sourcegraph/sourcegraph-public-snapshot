DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_update;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_insert;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_delete;
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_update();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_insert();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_delete();

DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_update;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_insert;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_delete;
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_gitserver_repos_update();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_gitserver_repos_insert();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_gitserver_repos_delete();

DROP TABLE IF EXISTS repo_statistics;
