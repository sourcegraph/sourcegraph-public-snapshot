-- +++
-- parent: 1528395864
-- +++

BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE insights_query_runner_jobs_dependencies
(
    id             SERIAL    NOT NULL,
    job_id         INT       NOT NULL,
    recording_time TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    --  The delete cascade is intentional, these records only have meaning in context of the related job row.
    CONSTRAINT insights_query_runner_jobs_dependencies_fk_job_id FOREIGN KEY (job_id) REFERENCES insights_query_runner_jobs (id) ON DELETE CASCADE
);

COMMENT ON TABLE insights_query_runner_jobs_dependencies IS 'Stores data points for a code insight that do not need to be queried directly, but depend on the result of a query at a different point';

COMMENT ON COLUMN insights_query_runner_jobs_dependencies.job_id IS 'Foreign key to the job that owns this record.';
COMMENT ON COLUMN insights_query_runner_jobs_dependencies.recording_time IS 'The time for which this dependency should be recorded at using the parents value.';

COMMIT;
