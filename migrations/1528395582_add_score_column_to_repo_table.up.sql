BEGIN;
-- A default of 0 is better than a default of NULL because NULL comes first in ORDER DESC.
ALTER TABLE repo ADD COLUMN score INT DEFAULT(0);
CREATE INDEX repo_score ON repo(score);
COMMIT;
