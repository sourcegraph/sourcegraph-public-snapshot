-- Create and backfill new bigint identity column (with temporary name)
ALTER TABLE codeintel_ranking_references_processed ADD COLUMN idx bigint;
UPDATE codeintel_ranking_references_processed SET idx = id;

-- Alter integer sequence to be backed by bigint
ALTER SEQUENCE codeintel_ranking_references_processed_id_seq as bigint MAXVALUE 9223372036854775807;

-- Register sequence as column default (POST BACKFILL)
ALTER TABLE codeintel_ranking_references_processed ALTER COLUMN idx SET DEFAULT nextval('codeintel_ranking_references_processed_id_seq'::regclass);

-- Swap primary key constraint
ALTER TABLE codeintel_ranking_references_processed DROP CONSTRAINT codeintel_ranking_references_processed_pkey;
ALTER TABLE codeintel_ranking_references_processed ADD PRIMARY KEY (idx);

-- Swap ownership (sequence can't be owned by id before we drop it)
ALTER SEQUENCE codeintel_ranking_references_processed_id_seq OWNED BY codeintel_ranking_references_processed.idx;

-- Swap id columns
ALTER TABLE codeintel_ranking_references_processed DROP COLUMN id;
ALTER TABLE codeintel_ranking_references_processed RENAME COLUMN idx TO id;
