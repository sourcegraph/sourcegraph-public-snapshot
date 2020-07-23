BEGIN;

DROP TABLE lsif_commits;

---
---
---

DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_uploads_with_repository_name;
DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads DROP COLUMN visible_at_tip;

CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

CREATE VIEW lsif_dumps_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_dumps u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_uploads_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_uploads u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

--
--
--

CREATE TABLE lsif_nearest_uploads (
    repository_id integer NOT NULL,
    "commit" text NOT NULL,
    upload_id integer NOT NULL,
    distance integer NOT NULL
);

CREATE TABLE lsif_uploads_visible_at_tip (
    repository_id integer NOT NULL,
    upload_id integer NOT NULL
);

CREATE TABLE lsif_dirty_repositories (
    repository_id integer PRIMARY KEY,
    dirty_token int NOT NULL,
    update_token int NOT NULL
);

CREATE INDEX lsif_nearest_uploads_repository_id_commit ON lsif_nearest_uploads(repository_id, "commit");
CREATE INDEX lsif_uploads_visible_at_tip_repository_id ON lsif_uploads_visible_at_tip(repository_id);

COMMIT;
