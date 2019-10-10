SELECT remote_exec('_lsif', '
    DROP INDEX IF EXISTS commits_parent_commit;
    DROP INDEX IF EXISTS commits_commit;
');
