SELECT remote_exec('_lsif', '
    DROP VIEW IF EXISTS commit_with_lsif_markers;
    DROP INDEX IF EXISTS commits_repo_commit_parent_commit_unique;
    DROP INDEX IF EXISTS commits_repo_commit;
    DROP INDEX IF EXISTS commits_repo_parent_commit;
    DROP TABLE IF EXISTS commits;
    DROP TABLE IF EXISTS lsif_data_markers;
');
