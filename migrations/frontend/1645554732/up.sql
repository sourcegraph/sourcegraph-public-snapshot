ALTER TABLE settings
    ALTER COLUMN contents SET DEFAULT '{}';

UPDATE settings
SET contents = '{}'
WHERE contents = NULL
    OR contents = '';

ALTER TABLE settings
    ALTER COLUMN contents SET NOT NULL,
    DROP CONSTRAINT IF EXISTS settings_no_empty_contents,
    ADD CONSTRAINT settings_no_empty_contents CHECK ( contents <> '' );
