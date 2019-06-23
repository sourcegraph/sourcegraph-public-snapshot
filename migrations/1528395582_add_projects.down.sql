BEGIN;

DROP TABLE IF EXISTS labels_objects;
DROP TABLE IF EXISTS labels;

ALTER TABLE discussion_threads DROP COLUMN project_id;

DROP TABLE projects;

ALTER TABLE discussion_threads_target_repo DROP COLUMN is_ignored;

ALTER TABLE discussion_threads DROP COLUMN status;
ALTER TABLE discussion_threads DROP COLUMN type;
ALTER TABLE discussion_threads DROP COLUMN settings;

COMMIT;
