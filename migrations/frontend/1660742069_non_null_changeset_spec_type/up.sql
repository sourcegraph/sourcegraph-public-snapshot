-- Ensure no unmigrated records exist.
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

ALTER TABLE changeset_specs ALTER COLUMN type SET NOT NULL;
ALTER TABLE changeset_specs ALTER COLUMN spec DROP NOT NULL;
UPDATE changeset_specs SET spec = NULL;
