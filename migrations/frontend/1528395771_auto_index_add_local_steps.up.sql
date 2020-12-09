BEGIN;

ALTER TABLE lsif_indexes ADD COLUMN local_steps json;

UPDATE lsif_indexes SET local_steps = '{}';

DROP VIEW lsif_indexes_with_repository_name;

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_indexes u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;