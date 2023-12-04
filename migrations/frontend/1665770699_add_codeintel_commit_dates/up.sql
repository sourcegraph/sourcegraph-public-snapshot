CREATE TABLE IF NOT EXISTS codeintel_commit_dates(
    repository_id integer NOT NULL,
    commit_bytea bytea NOT NULL,
    committed_at timestamp with time zone,
    PRIMARY KEY(repository_id, commit_bytea)
);

COMMENT ON TABLE codeintel_commit_dates IS 'Maps commits within a repository to the commit date as reported by gitserver.';

COMMENT ON COLUMN codeintel_commit_dates.repository_id IS 'Identifies a row in the `repo` table.';

COMMENT ON COLUMN codeintel_commit_dates.commit_bytea IS 'Identifies the 40-character commit hash.';

COMMENT ON COLUMN codeintel_commit_dates.committed_at IS 'The commit date (may be -infinity if unresolvable).';

INSERT INTO
    codeintel_commit_dates (repository_id, commit_bytea, committed_at)
SELECT
    u.repository_id,
    decode(u.commit, 'hex'),
    MIN(u.committed_at)
FROM
    lsif_uploads u
WHERE
    u.committed_at IS NOT NULL
GROUP BY
    u.repository_id,
    u.commit
ON CONFLICT DO NOTHING;
