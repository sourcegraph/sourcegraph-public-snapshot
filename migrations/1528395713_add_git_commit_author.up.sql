BEGIN;

-- Add default values for git commit author (name and email)

UPDATE changeset_specs SET spec = spec || '{ "commits": [{"authorName": "Sourcegraph", "authorEmail": "campaigns@sourcegraph.com"}] }' WHERE spec->'externalId' IS NULL;

COMMIT;
