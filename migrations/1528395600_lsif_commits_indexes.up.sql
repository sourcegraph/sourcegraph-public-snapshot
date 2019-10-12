-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

CREATE INDEX IF NOT EXISTS lsif_commits_parent_commit ON lsif_commits(repository, parent_commit);
CREATE INDEX IF NOT EXISTS lsif_commits_commit ON lsif_commits(repository, "commit");

END;
