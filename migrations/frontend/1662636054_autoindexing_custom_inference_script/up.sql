CREATE TABLE IF NOT EXISTS codeintel_inference_scripts (
    script text NOT NULL
);

INSERT INTO codeintel_inference_scripts VALUES ('');

COMMENT ON table codeintel_inference_scripts IS 'Contains auto-index job inference Lua scripts as an alternative to setting via environment variables.';
