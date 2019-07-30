BEGIN;
ALTER TABLE repo ADD COLUMN score INT;
CREATE INDEX repo_score ON repo(score);
COMMIT;
