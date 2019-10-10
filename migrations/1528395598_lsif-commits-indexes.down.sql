-- Note: `commit` is a reserved word, so it's quoted.

SELECT remote_exec('_lsif', '
    DROP INDEX IF EXISTS commits_parent_commit;
    DROP INDEX IF EXISTS commits_commit;
');
