DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_update;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_insert;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_delete;
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_update();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_insert();
DROP FUNCTION IF EXISTS recalc_repo_statistics_on_repo_delete();

DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_update;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_insert;
DROP TRIGGER IF EXISTS trig_recalc_gitserver_repos_statistics_on_delete;
DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_update() RETURNS trigger
DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_insert() RETURNS trigger
DROP FUNCTION IF EXISTS recalc_gitserver_repos_statistics_on_delete() RETURNS trigger

DROP TABLE IF EXISTS gitserver_repos_statistics;
DROP TABLE IF EXISTS repo_statistics;
