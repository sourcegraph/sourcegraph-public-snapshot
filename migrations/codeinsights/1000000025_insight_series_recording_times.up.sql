BEGIN;

-- loose schema
create table if not exists insight_series_recording_times
(
    id SERIAL
        constraint insight_series_recording_times_pk
            primary key,
    insight_series_id int not null,
    recording_time timestamp not null,
    repository_offset int not null DEFAULT 0,
--     compression_group INT not null DEFAULT NEXTVAL('insight_series_recording_times_group_seq'),

    CONSTRAINT insight_series_recording_times_insight_series_id_fk FOREIGN KEY(insight_series_id) REFERENCES insight_series(id)
);

COMMIT;
