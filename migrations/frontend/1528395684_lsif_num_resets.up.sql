BEGIN;

DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads ADD COLUMN num_resets integer NOT NULL DEFAULT 0;
ALTER TABLE lsif_indexes ADD COLUMN num_resets integer NOT NULL DEFAULT 0;

-- Recreate view with new columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
