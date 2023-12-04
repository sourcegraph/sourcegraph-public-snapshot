CREATE TABLE IF NOT EXISTS insight_series_backfill
(
    id               SERIAL
        CONSTRAINT insight_series_backfill_pk PRIMARY KEY,
    series_id        INT  NOT NULL,
    repo_iterator_id INT,
    estimated_cost   DOUBLE PRECISION,
    state            TEXT NOT NULL DEFAULT 'new',

    CONSTRAINT insight_series_backfill_series_id_fk
        FOREIGN KEY (series_id) REFERENCES insight_series (id) ON DELETE CASCADE
);

ALTER TABLE insights_background_jobs
    ADD COLUMN IF NOT EXISTS backfill_id INT REFERENCES insight_series_backfill(id) ON DELETE CASCADE;

CREATE OR REPLACE VIEW insights_jobs_backfill_in_progress AS
SELECT jobs.*, isb.state AS backfill_state, isb.estimated_cost
FROM insights_background_jobs jobs
         JOIN insight_series_backfill isb ON jobs.backfill_id = isb.id
WHERE isb.state = 'processing';

CREATE OR REPLACE VIEW insights_jobs_backfill_new AS
SELECT jobs.*, isb.state AS backfill_state, isb.estimated_cost
FROM insights_background_jobs jobs
         JOIN insight_series_backfill isb ON jobs.backfill_id = isb.id
WHERE isb.state = 'new';
