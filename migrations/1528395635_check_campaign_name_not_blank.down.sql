BEGIN;

ALTER TABLE campaigns DROP CONSTRAINT campaigns_name_not_blank;

COMMIT;
