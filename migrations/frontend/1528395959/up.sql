DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'batch_changes' AND column_name = 'initial_applier_id')
    THEN
        ALTER TABLE batch_changes RENAME COLUMN initial_applier_id TO creator_id;
    END IF;
END $$;
