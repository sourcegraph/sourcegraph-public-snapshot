ALTER TABLE repo ALTER COLUMN created_at SET DEFAULT now();
UPDATE repo SET created_at=now() WHERE created_at IS NULL;
ALTER TABLE repo ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE repo DROP COLUMN "blocked";
ALTER TABLE repo DROP COLUMN "owner";
ALTER TABLE repo DROP COLUMN "name";