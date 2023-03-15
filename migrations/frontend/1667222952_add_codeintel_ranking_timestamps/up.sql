ALTER TABLE
    codeintel_path_ranks
ADD
    COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS codeintel_path_ranks_updated_at ON codeintel_path_ranks(updated_at) INCLUDE (repository_id);

CREATE OR REPLACE FUNCTION update_codeintel_path_ranks_updated_at_column() RETURNS TRIGGER AS
$$ BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_codeintel_path_ranks_updated_at ON codeintel_path_ranks;
CREATE TRIGGER update_codeintel_path_ranks_updated_at BEFORE UPDATE ON codeintel_path_ranks FOR EACH ROW EXECUTE PROCEDURE update_codeintel_path_ranks_updated_at_column();
