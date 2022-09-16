ALTER TABLE changeset_specs ALTER COLUMN type DROP NOT NULL;
ALTER TABLE changeset_specs ADD COLUMN IF NOT EXISTS spec jsonb;

UPDATE
    changeset_specs
SET
    spec = jsonb_build_object(
        'baseRev', base_rev,
        'baseRef', base_ref,
        'externalID', external_id,
        'headRef', head_ref,
        'title', title,
        'body', body,
        'published', published,
        'commits', json_build_array(
            jsonb_build_object(
                'diff', encode(diff, 'escape'),
                'message', commit_message,
                'authorName', commit_author_name,
                'authorEmail', commit_author_email
            )
        )
    );

ALTER TABLE changeset_specs ALTER COLUMN spec SET NOT NULL;
