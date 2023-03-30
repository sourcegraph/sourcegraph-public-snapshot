-- we create a temporary column to store citext for this migration
ALTER TABLE roles ADD COLUMN IF NOT EXISTS name_citext citext;

-- we then, update the created column to have the same values as `name`
UPDATE roles SET name_citext = name;

-- remove the previous constraint on roles.name
ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_name;

-- Drop the current `name` which is of type `text`
ALTER TABLE roles DROP COLUMN IF EXISTS name;

DO $$
BEGIN
    -- Rename the newly created column to `name`.
    ALTER TABLE roles RENAME COLUMN name_citext TO name;
EXCEPTION
    WHEN undefined_column THEN RAISE NOTICE 'column name_text does not exist in table roles';
END $$;

ALTER TABLE roles ALTER COLUMN name SET NOT NULL;

-- add a unique index
CREATE UNIQUE INDEX IF NOT EXISTS unique_role_name ON roles (name);
