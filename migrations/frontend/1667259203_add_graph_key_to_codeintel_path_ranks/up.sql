ALTER TABLE
    codeintel_path_ranks
ADD
    COLUMN IF NOT EXISTS graph_key text;
