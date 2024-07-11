DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pattern_type') THEN
        CREATE TYPE pattern_type AS ENUM ('keyword', 'literal', 'regexp', 'standard', 'structural');
    END IF;
END
$$;

ALTER TABLE notebooks ADD COLUMN IF NOT EXISTS pattern_type pattern_type NOT NULL DEFAULT 'standard';
