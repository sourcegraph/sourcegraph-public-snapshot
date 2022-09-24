drop table if exists series_points_compressed;

alter table insight_series drop COLUMN if exists data_format;

