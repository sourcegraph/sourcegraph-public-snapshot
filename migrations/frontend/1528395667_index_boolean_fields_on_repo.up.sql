BEGIN;

CREATE INDEX IF NOT EXISTS repo_private ON repo(private);
CREATE INDEX IF NOT EXISTS repo_fork ON repo(fork);
CREATE INDEX IF NOT EXISTS repo_archived ON repo(archived);

COMMIT;
