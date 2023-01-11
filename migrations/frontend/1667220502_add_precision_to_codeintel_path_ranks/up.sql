DELETE FROM
    codeintel_path_ranks;

ALTER TABLE
    codeintel_path_ranks
ADD
    COLUMN IF NOT EXISTS precision float NOT NULL;
