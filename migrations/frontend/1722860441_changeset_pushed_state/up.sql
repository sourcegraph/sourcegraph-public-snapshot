ALTER TABLE changesets ADD COLUMN IF NOT EXISTS pushed boolean NOT NULL DEFAULT false;
ALTER TABLE changeset_specs ADD COLUMN IF NOT EXISTS pushed boolean NOT NULL DEFAULT false;
