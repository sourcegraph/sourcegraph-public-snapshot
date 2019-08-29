BEGIN;

ALTER TABLE events DROP COLUMN rule_id;

DROP TABLE IF EXISTS rules;

ALTER TABLE campaigns DROP COLUMN due_date;
ALTER TABLE campaigns DROP COLUMN start_date;
ALTER TABLE campaigns DROP COLUMN is_draft;
ALTER TABLE campaigns DROP COLUMN template_context;
ALTER TABLE campaigns DROP COLUMN template_id;

ALTER TABLE threads DROP COLUMN is_draft;

COMMIT;
