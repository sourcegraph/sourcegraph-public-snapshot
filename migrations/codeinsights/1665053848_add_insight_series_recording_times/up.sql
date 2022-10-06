CREATE TABLE insight_series_recording_times(
	series_id text,
	recording_time timestamptz,
	unique(series_id, recording_time)
);