-- recreate old table with comments 
CREATE TABLE metadata (
	id bigint NOT NULL,
	metadata jsonb NOT NULL
);

ALTER TABLE ONLY metadata
    ADD CONSTRAINT metadata_pkey PRIMARY KEY (id);

COMMENT ON TABLE metadata IS 'Records arbitrary metadata about events. Stored in a separate table as it is often repeated for multiple events.';

COMMENT ON COLUMN metadata.id IS 'The metadata ID.';

COMMENT ON COLUMN metadata.metadata IS 'Metadata about some event, this can be any arbitrary JSON emtadata which will be returned when querying events, and can be filtered on and grouped using jsonb operators ?, ?&, ?|, and @>. This should be small data only.';

CREATE SEQUENCE metadata_id_seq
	START WITH 1
	INCREMENT BY 1
	NO MINVALUE
	NO MAXVALUE
	CACHE 1;

ALTER SEQUENCE metadata_id_seq OWNED BY metadata.id;

ALTER TABLE ONLY metadata ALTER COLUMN id SET DEFAULT nextval('metadata_id_seq'::regclass);

CREATE INDEX metadata_metadata_gin ON metadata USING gin (metadata);

CREATE UNIQUE INDEX metadata_metadata_unique_idx ON metadata USING btree (metadata);

ALTER TABLE IF EXISTS series_points
	ADD COLUMN metadata_id integer;

COMMENT ON COLUMN series_points.metadata_id IS 'Associated metadata for this event, if any.';

ALTER TABLE IF EXISTS series_points_snapshots
	ADD COLUMN metadata_id integer;

COMMENT ON COLUMN series_points_snapshots.metadata_id IS 'Associated metadata for this event, if any.';

ALTER TABLE ONLY series_points
    ADD CONSTRAINT series_points_metadata_id_fkey FOREIGN KEY (metadata_id) REFERENCES metadata(id) ON DELETE CASCADE DEFERRABLE;
