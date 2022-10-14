CREATE TABLE IF NOT EXISTS insight_series_recording_times(
	series_id int,
	recording_time timestamptz,
	unique(series_id, recording_time),
	constraint series_id_fkey foreign key(series_id) references insight_series(id)
);