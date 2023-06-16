ALTER TABLE changesets
    ADD COLUMN IF NOT EXISTS commit_verification jsonb DEFAULT '{}'::jsonb NOT NULL;
