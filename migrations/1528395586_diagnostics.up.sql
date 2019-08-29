BEGIN;

CREATE TABLE thread_diagnostic_edges (
	id bigserial PRIMARY KEY,
	thread_id bigint NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
	--TODO!(sqs) location_repository_id integer NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
	type text NOT NULL,
	data jsonb NOT NULL
);
--TODO!(sqs) CREATE INDEX threads_diagnostics_location_repository_id ON threads_diagnostics(location_repository_id);
CREATE INDEX thread_diagnostic_edges_thread_id ON thread_diagnostic_edges(thread_id);

ALTER TABLE events ADD COLUMN thread_diagnostic_edge_id bigint REFERENCES thread_diagnostic_edges(id) ON DELETE SET NULL;

COMMIT;
