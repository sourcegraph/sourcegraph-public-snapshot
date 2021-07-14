BEGIN;

CREATE TABLE insight_series
(
    id                      SERIAL    NOT NULL PRIMARY KEY,
    series_id               TEXT      NOT NULL,
    query                   TEXT      NOT NULL,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    oldest_historical_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP - INTERVAL '1 year',
    last_recorded_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP - INTERVAL '10 year',
    next_recording_after    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recording_interval_days INT       NOT NULL DEFAULT 1,
    deleted_at              TIMESTAMP
);

CREATE UNIQUE INDEX insight_series_series_id_unique_idx ON insight_series (series_id);
CREATE INDEX insight_series_deleted_at_idx ON insight_series (deleted_at);
CREATE INDEX insight_series_next_recording_after_idx ON insight_series (next_recording_after);

CREATE TABLE insight_view
(
    id          SERIAL NOT NULL PRIMARY KEY,
    title       TEXT,
    description TEXT,
    unique_id   TEXT
);

CREATE UNIQUE INDEX insight_view_unique_id_unique_idx ON insight_view (unique_id);

CREATE TABLE insight_view_series
(
    insight_view_id   INT,
    insight_series_id INT,
    label             TEXT,
    stroke            TEXT,
    PRIMARY KEY (insight_view_id, insight_series_id)
);

ALTER TABLE insight_view_series
    ADD FOREIGN KEY (insight_view_id) REFERENCES insight_view (id);

ALTER TABLE insight_view_series
    ADD FOREIGN KEY (insight_series_id) REFERENCES insight_series (id);

COMMIT;
