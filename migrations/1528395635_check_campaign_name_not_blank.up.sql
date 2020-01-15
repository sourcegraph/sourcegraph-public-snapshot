BEGIN;

UPDATE campaigns SET name = 'Campaign #' || campaigns.id WHERE campaigns.name = '';
ALTER TABLE campaigns ADD CONSTRAINT campaigns_name_not_blank CHECK (name <> ''::text);

COMMIT;
