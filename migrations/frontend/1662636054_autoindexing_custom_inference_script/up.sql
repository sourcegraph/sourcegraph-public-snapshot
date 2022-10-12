CREATE TABLE IF NOT EXISTS codeintel_inference_scripts (
    insert_timestamp timestamptz NOT NULL default NOW(),
    script text NOT NULL
);

COMMENT ON table codeintel_inference_scripts IS 'Contains auto-index job inference Lua scripts as an alternative to setting via environment variables.';
