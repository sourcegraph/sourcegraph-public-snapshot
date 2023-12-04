-- we create a temporary column to store text for this migration
ALTER TABLE roles ADD COLUMN IF NOT EXISTS name_text text;

-- we then, update the created column to have the same values as `name`
UPDATE roles SET name_text = name;

DROP INDEX IF EXISTS unique_role_name;

-- Drop the current `name` which is of type `citext`
ALTER TABLE roles DROP COLUMN name;

DO $$
BEGIN
    -- Rename the newly created column to `name`.
    ALTER TABLE roles RENAME COLUMN name_text TO name;
    ALTER TABLE roles ADD CONSTRAINT roles_name UNIQUE (name);
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column name_text does not exist in table roles';
    WHEN duplicate_object THEN RAISE NOTICE 'constrant roles_name already exists';
END $$;

ALTER TABLE roles ALTER COLUMN name SET NOT NULL;
ALTER TABLE roles ADD CONSTRAINT name_not_blank CHECK (name <> ''::text);
