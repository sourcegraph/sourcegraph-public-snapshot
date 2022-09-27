ALTER TABLE changeset_specs
    ALTER COLUMN published TYPE VARCHAR;

DROP TYPE IF EXISTS changeset_spec_published_enum;
