CREATE TABLE IF NOT EXISTS codeintel_global_ranks (
    payload jsonb NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_codeintel_global_ranks_updated_at_column() RETURNS TRIGGER AS
$$ BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_codeintel_global_ranks_updated_at ON codeintel_path_ranks;
CREATE TRIGGER update_codeintel_global_ranks_updated_at BEFORE UPDATE ON codeintel_path_ranks FOR EACH ROW EXECUTE PROCEDURE update_codeintel_global_ranks_updated_at_column();
