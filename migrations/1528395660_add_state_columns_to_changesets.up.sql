BEGIN;

CREATE TYPE changeset_external_state AS ENUM (
    'UNKNOWN',
    'OPEN',
    'CLOSED',
    'MERGED',
    'DELETED'
    );

CREATE TYPE changeset_external_review_state AS ENUM (
    'UNKNOWN',
    'APPROVED',
    'CHANGES_REQUESTED',
    'PENDING',
    'COMMENTED',
    'DISMISSED'
    );

CREATE TYPE changeset_external_check_state AS ENUM (
    'UNKNOWN',
    'PENDING',
    'PASSED',
    'FAILED'
    );

ALTER TABLE changesets
    ADD COLUMN external_state changeset_external_state;

ALTER TABLE changesets
    ADD COLUMN external_review_state changeset_external_review_state;

ALTER TABLE changesets
    ADD COLUMN external_check_state changeset_external_check_state;

UPDATE changesets SET external_state = 'UNKNOWN';
UPDATE changesets SET external_review_state = 'UNKNOWN';
UPDATE changesets SET external_check_state = 'UNKNOWN';

COMMIT;
