ALTER TABLE changeset_specs
    ALTER COLUMN published TYPE text;

DROP TYPE IF EXISTS changeset_spec_published_enum;
