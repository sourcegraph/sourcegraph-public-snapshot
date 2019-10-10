BEGIN;

DROP VIEW IF EXISTS lsif_commits_with_lsif_data_markers;

DROP INDEX IF EXISTS lsif_commits_repo_commit_parent_commit_unique;
DROP INDEX IF EXISTS lsif_commits_repo_commit;
DROP INDEX IF EXISTS lsif_commits_repo_parent_commit;
DROP INDEX IF EXISTS lsif_packages_package_unique;
DROP INDEX IF EXISTS lsif_packages_repo_commit;
DROP INDEX IF EXISTS lsif_references_package;
DROP INDEX IF EXISTS lsif_references_repo_commit;

DROP TABLE IF EXISTS lsif_commits;
DROP TABLE IF EXISTS lsif_data_markers;
DROP TABLE IF EXISTS lsif_packages;
DROP TABLE IF EXISTS lsif_references;

COMMIT;
