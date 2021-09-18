BEGIN;

ALTER TABLE IF EXISTS repo ADD COLUMN blocked jsonb;

CREATE OR REPLACE FUNCTION repo_block(reason text, at timestamptz) RETURNS jsonb AS
$$
SELECT jsonb_build_object(
    'reason', reason,
    'at', extract(epoch from timezone('utc', at))::bigint
);
$$ LANGUAGE SQL STRICT IMMUTABLE;

CREATE INDEX repo_blocked_idx ON repo USING BTREE ((blocked IS NOT NULL));
CREATE INDEX repo_is_not_blocked_idx ON repo USING BTREE ((blocked IS NULL));

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
