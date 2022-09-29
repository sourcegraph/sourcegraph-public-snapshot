DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 from pg_type WHERE typname = 'changeset_spec_published_enum') THEN
            CREATE TYPE changeset_spec_published_enum AS ENUM (
                'true',
                'false',
                '"draft"'
                );
        END IF;
    END
$$;

ALTER TABLE changeset_specs
    ALTER COLUMN published TYPE changeset_spec_published_enum USING published::text::changeset_spec_published_enum;
