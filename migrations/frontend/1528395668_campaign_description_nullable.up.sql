BEGIN;

ALTER TABLE campaigns ALTER COLUMN description DROP NOT NULL;
UPDATE campaigns SET description = NULL WHERE description = '';

COMMIT;
