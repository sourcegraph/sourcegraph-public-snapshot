BEGIN;

-- Recreate lsif_indexes view with columns introduced in 1528395730_lsif_index_log_contents.up.sql.
DROP VIEW lsif_indexes_with_repository_name;

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_indexes u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;
