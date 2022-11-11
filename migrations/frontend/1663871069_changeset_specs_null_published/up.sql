UPDATE changeset_specs
SET published = NULL
WHERE published = 'null';

UPDATE changeset_specs
SET published = '"draft"'
WHERE published = 'draft';

ALTER TABLE changeset_specs
    DROP CONSTRAINT IF EXISTS changeset_specs_published_valid_values,
    ADD CONSTRAINT changeset_specs_published_valid_values CHECK (published = 'true' OR published = 'false' OR
                                                                 published = '"draft"' OR published IS NULL);
