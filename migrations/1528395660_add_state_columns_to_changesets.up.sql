BEGIN;

ALTER TABLE changesets
    ADD COLUMN external_state text;

ALTER TABLE changesets
    ADD COLUMN external_review_state text;

ALTER TABLE changesets
    ADD COLUMN external_check_state text;

-- Set some sane defaults. If not, we'll get failures on the frontend as we'll have invalid
-- values in the DB. These values will update on next changeset sync or webhook.

UPDATE changesets set external_state = 'OPEN',
                      external_review_state = 'PENDING',
                      external_check_state = 'UNKNOWN';

COMMIT;
