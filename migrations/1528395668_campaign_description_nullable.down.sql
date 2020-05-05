BEGIN;

UPDATE campaigns SET description = '' WHERE description IS NULL;

ALTER TABLE campaigns ALTER COLUMN description SET NOT NULL;

COMMIT;
