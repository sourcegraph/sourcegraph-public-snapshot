ALTER TABLE repo
  ALTER COLUMN vcs SET DEFAULT '',
  ALTER COLUMN vcs SET NOT NULL,
  ALTER COLUMN default_branch SET DEFAULT '',
  ALTER COLUMN default_branch SET NOT NULL;

-- Add default otherwise down will fail due to rows existing which are
-- null. Note in our code we always have these fields empty.
