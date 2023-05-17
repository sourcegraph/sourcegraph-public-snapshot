-- Hey you know what, deal with it.

DROP SEQUENCE IF EXISTS codeintel_ranking_references_processed_id_seq CASCADE;
CREATE SEQUENCE codeintel_ranking_references_processed_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- ALTER SEQUENCE codeintel_ranking_references_processed_id_seq OWNED BY codeintel_ranking_references_processed.id;
ALTER SEQUENCE codeintel_ranking_references_processed_id_seq as integer MAXVALUE 2147483647;
ALTER TABLE ONLY codeintel_ranking_references_processed ALTER COLUMN id TYPE integer USING (id::integer);
ALTER TABLE ONLY codeintel_ranking_references_processed ALTER COLUMN id SET DEFAULT nextval('codeintel_ranking_references_processed_id_seq'::regclass);
-- ALTER SEQUENCE codeintel_ranking_references_processed_id_seq as bigint MAXVALUE 2147483647;
