create table if not exists series_points_compressed
(
    id serial constraint compressed_series_points_pk primary key,
    series_id INT not NULL,
    repo_id int not null,
    repo_name_id INTEGER REFERENCES repo_names(id) ON DELETE CASCADE DEFERRABLE,
    data bytea not null,
    capture text
);

alter table insight_series add COLUMN if not exists data_format int not null default 1;
