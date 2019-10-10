-- Note: `commit` is a reserved word, so it's quoted.

SELECT remote_exec('_lsif', '
    CREATE INDEX IF NOT EXISTS commits_parent_commit ON commits(repository, parent_commit);
    CREATE INDEX IF NOT EXISTS commits_commit ON commits(repository, "commit");
');
