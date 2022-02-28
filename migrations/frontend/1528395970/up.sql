ALTER TABLE cm_trigger_jobs
    DROP CONSTRAINT IF EXISTS search_results_is_array,
    ADD COLUMN IF NOT EXISTS search_results JSONB,
    ADD CONSTRAINT search_results_is_array CHECK (jsonb_typeof(search_results) = 'array');

COMMENT ON COLUMN cm_trigger_jobs.results IS 'DEPRECATED: replaced by len(search_results) > 0. Can be removed after version 3.37 release cut';
COMMENT ON COLUMN cm_trigger_jobs.num_results IS 'DEPRECATED: replaced by len(search_results). Can be removed after version 3.37 release cut';
