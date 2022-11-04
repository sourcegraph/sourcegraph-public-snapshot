ALTER TABLE changeset_specs
    ADD COLUMN IF NOT EXISTS diff bytea,
    ADD COLUMN IF NOT EXISTS base_rev TEXT,
    ADD COLUMN IF NOT EXISTS base_ref TEXT,
    ADD COLUMN IF NOT EXISTS body TEXT,
    ADD COLUMN IF NOT EXISTS published TEXT,
    ADD COLUMN IF NOT EXISTS commit_message TEXT,
    ADD COLUMN IF NOT EXISTS commit_author_name TEXT,
    ADD COLUMN IF NOT EXISTS commit_author_email TEXT,
    ADD COLUMN IF NOT EXISTS type TEXT;

UPDATE
    changeset_specs
SET
    diff = convert_to(spec->'commits'->0->>'diff', 'UTF8'),
    base_rev = spec->>'baseRev',
    base_ref = spec->>'baseRef',
    body = spec->>'body',
    published = spec->>'published',
    commit_message = spec->'commits'->0->>'message',
    commit_author_name = spec->'commits'->0->>'authorName',
    commit_author_email = spec->'commits'->0->>'authorEmail',
    type = CASE WHEN spec->>'externalID' IS NOT NULL THEN 'existing' ELSE 'branch' END;
