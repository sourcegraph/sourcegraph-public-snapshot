ALTER TABLE settings
    ALTER COLUMN contents DROP NOT NULL;

ALTER TABLE settings
    ALTER COLUMN contents DROP DEFAULT,
    DROP CONSTRAINT IF EXISTS settings_no_empty_contents;
