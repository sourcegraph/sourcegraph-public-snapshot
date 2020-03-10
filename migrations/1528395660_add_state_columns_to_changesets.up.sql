BEGIN;

ALTER TABLE changesets
    ADD COLUMN external_state text;

ALTER TABLE changesets
    ADD COLUMN external_review_state text;

ALTER TABLE changesets
    ADD COLUMN external_check_state text;

COMMIT;
