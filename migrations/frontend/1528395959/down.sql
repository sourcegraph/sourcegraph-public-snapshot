DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'batch_changes' AND column_name = 'creator_id')
    THEN
        ALTER TABLE batch_changes RENAME COLUMN creator_id TO initial_applier_id;
    END IF;
END $$;
