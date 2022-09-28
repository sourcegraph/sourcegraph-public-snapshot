UPDATE changeset_specs
SET published = NULL
WHERE published = 'null';

DO
$$
    BEGIN
        CREATE TYPE changeset_spec_published_enum AS ENUM (
            'true',
            'false',
            '"draft"'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

ALTER TABLE changeset_specs
    ALTER COLUMN published TYPE changeset_spec_published_enum USING published::text::changeset_spec_published_enum;
