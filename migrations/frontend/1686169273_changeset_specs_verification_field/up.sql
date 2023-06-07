ALTER TABLE changeset_specs
    ADD COLUMN IF NOT EXISTS commit_verification jsonb DEFAULT NULL;
