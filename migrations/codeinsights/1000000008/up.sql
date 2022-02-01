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

comment on table insight_series is 'Data series that comprise code insights.';

comment on column insight_series.id is 'Primary key ID of this series';
comment on column insight_series.series_id is 'Unique Series ID represents a globally unique identifier for this series.';
comment on column insight_series.query is 'Query string that generates this series';
comment on column insight_series.created_at is 'Timestamp when this series was created';
comment on column insight_series.oldest_historical_at is 'Timestamp representing the oldest point of which this series is backfilled.';
comment on column insight_series.last_recorded_at is 'Timestamp when this series was last recorded (non-historical).';
comment on column insight_series.next_recording_after is 'Timestamp when this series should next record (non-historical).';
comment on column insight_series.recording_interval_days is 'Number of days that should pass between recordings (non-historical)';
comment on column insight_series.deleted_at is 'Timestamp of a soft-delete of this row.';

CREATE UNIQUE INDEX insight_series_series_id_unique_idx ON insight_series (series_id);
CREATE INDEX insight_series_deleted_at_idx ON insight_series (deleted_at);
CREATE INDEX insight_series_next_recording_after_idx ON insight_series (next_recording_after);

CREATE TABLE insight_view
(
    id          SERIAL NOT NULL PRIMARY KEY,
    title       TEXT,
    description TEXT,
    unique_id   TEXT NOT NULL
);

comment on table insight_view is 'Views for insight data series. An insight view is an abstraction on top of an insight data series that allows for lightweight modifications to filters or metadata without regenerating the underlying series.';

comment on column insight_view.id is 'Primary key ID for this view';
comment on column insight_view.title is 'Title of the view. This may render in a chart depending on the view type.';
comment on column insight_view.description is 'Description of the view. This may render in a chart depending on the view type.';
comment on column insight_view.unique_id is 'Globally unique identifier for this view that is externally referencable.';

CREATE UNIQUE INDEX insight_view_unique_id_unique_idx ON insight_view (unique_id);

CREATE TABLE insight_view_series
(
    insight_view_id   INT NOT NULL,
    insight_series_id INT NOT NULL,
    label             TEXT,
    stroke            TEXT,
    PRIMARY KEY (insight_view_id, insight_series_id)
);

comment on table insight_view_series is 'Join table to correlate data series with insight views';
comment on column insight_view_series.insight_view_id is 'Foreign key to insight view.';
comment on column insight_view_series.insight_series_id is 'Foreign key to insight data series.';
comment on column insight_view_series.label is 'Label text for this data series. This may render in a chart depending on the view type.';
comment on column insight_view_series.stroke is 'Stroke color metadata for this data series. This may render in a chart depending on the view type.';

ALTER TABLE insight_view_series
    ADD FOREIGN KEY (insight_view_id) REFERENCES insight_view (id);

ALTER TABLE insight_view_series
    ADD FOREIGN KEY (insight_series_id) REFERENCES insight_series (id);

COMMIT;
