ALTER TABLE repo
  ALTER COLUMN vcs SET DEFAULT '',
  ALTER COLUMN default_branch SET DEFAULT '';

UPDATE repo SET vcs = '' WHERE vcs IS NULL;
UPDATE repo SET default_branch = '' WHERE default_branch IS NULL;

ALTER TABLE repo
  ALTER COLUMN vcs SET NOT NULL,
  ALTER COLUMN default_branch SET NOT NULL;
