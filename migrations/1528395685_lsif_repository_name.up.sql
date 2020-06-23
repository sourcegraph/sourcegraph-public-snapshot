BEGIN;

CREATE VIEW lsif_dumps_with_repository_name AS SELECT u.*, regexp_replace(r.name, '^DELETED-\d+\.\d+-', '') as repository_name FROM lsif_dumps u JOIN repo r ON r.id = u.repository_id;
CREATE VIEW lsif_uploads_with_repository_name AS SELECT u.*, regexp_replace(r.name, '^DELETED-\d+\.\d+-', '') as repository_name FROM lsif_uploads u JOIN repo r ON r.id = u.repository_id;
CREATE VIEW lsif_indexes_with_repository_name AS SELECT u.*, regexp_replace(r.name, '^DELETED-\d+\.\d+-', '') as repository_name FROM lsif_indexes u JOIN repo r ON r.id = u.repository_id;

COMMIT;
