CREATE TABLE IF NOT EXISTS insight_series_incomplete_points
(
    id        SERIAL CONSTRAINT insight_series_incomplete_points_pk PRIMARY KEY,
    series_id INT                         NOT NULL,
    reason    TEXT                        NOT NULL,
    time      TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    repo_id   INT,

    CONSTRAINT insight_series_incomplete_points_series_id_fk
        FOREIGN KEY (series_id) REFERENCES insight_series (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS insight_series_incomplete_points_unique_idx
    ON insight_series_incomplete_points (series_id, reason, time, repo_id);
