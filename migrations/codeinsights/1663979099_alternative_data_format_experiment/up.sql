create table if not exists series_points_compressed
(
    id serial constraint compressed_series_points_pk primary key,
    series_id INT not NULL,
    repo_id int not null,
    data bytea,
    capture text
);

alter table insight_series add COLUMN if not exists data_format int not null default 1;

CREATE UNIQUE INDEX series_points_compressed_composite_key_idx on series_points_compressed(series_id, repo_id, capture)
