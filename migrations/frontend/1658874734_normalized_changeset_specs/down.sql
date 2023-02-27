ALTER TABLE changeset_specs
    DROP COLUMN IF EXISTS diff,
    DROP COLUMN IF EXISTS base_rev,
    DROP COLUMN IF EXISTS base_ref,
    DROP COLUMN IF EXISTS body,
    DROP COLUMN IF EXISTS published,
    DROP COLUMN IF EXISTS commit_message,
    DROP COLUMN IF EXISTS commit_author_name,
    DROP COLUMN IF EXISTS commit_author_email,
    DROP COLUMN IF EXISTS type;
