-- +++
-- parent: 1000000001
-- +++

BEGIN;

COMMENT ON TABLE repo_names IS 'Records repository names, both historical and present, using a unique repository _name_ ID (unrelated to the repository ID.)';
COMMENT ON COLUMN repo_names.id IS 'The repository _name_ ID.';
COMMENT ON COLUMN repo_names.name IS 'The repository name string, with unique constraint for table entry deduplication and trigram index for e.g. regex filtering.';

COMMENT ON TABLE metadata IS 'Records arbitrary metadata about events. Stored in a separate table as it is often repeated for multiple events.';
COMMENT ON COLUMN metadata.id IS 'The metadata ID.';
COMMENT ON COLUMN metadata.metadata IS 'Metadata about some event, this can be any arbitrary JSON emtadata which will be returned when querying events, and can be filtered on and grouped using jsonb operators ?, ?&, ?|, and @>. This should be small data only.';

COMMENT ON TABLE series_points IS 'Records events over time associated with a repository (or none, i.e. globally) where a single numerical value is going arbitrarily up and down.  Repository association is based on both repository ID and name. The ID can be used to refer toa specific repository, or lookup the current name of a repository after it has been e.g. renamed. The name can be used to refer to the name of the repository at the time of the events creation, for example to trace the change in a gauge back to a repository being renamed.';
COMMENT ON COLUMN series_points.series_id IS 'A unique identifier for the series of data being recorded. This is not an ID from another table, but rather just a unique identifier.';
COMMENT ON COLUMN series_points.time IS 'The timestamp of the recorded event.';
COMMENT ON COLUMN series_points.value IS 'The floating point value at the time of the event.';
COMMENT ON COLUMN series_points.metadata_id IS 'Associated metadata for this event, if any.';
COMMENT ON COLUMN series_points.repo_id IS 'The repository ID (from the main application DB) at the time the event was created. Note that the repository may no longer exist / be valid at query time, however.';
COMMENT ON COLUMN series_points.repo_name_id IS 'The most recently known name for the repository, updated periodically to account for e.g. repository renames. If the repository was deleted, this is still the most recently known name.  null if the event was not for a single repository (i.e. a global gauge).';
COMMENT ON COLUMN series_points.original_repo_name_id IS 'The repository name as it was known at the time the event was created. It may have been renamed since.';

COMMIT;
