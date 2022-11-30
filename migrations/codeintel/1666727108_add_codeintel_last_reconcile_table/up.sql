CREATE TABLE IF NOT EXISTS codeintel_last_reconcile (
    dump_id integer NOT NULL UNIQUE,
    last_reconcile_at timestamp with time zone NOT NULL
);

COMMENT ON TABLE codeintel_last_reconcile IS 'Stores the last time processed LSIF data was reconciled with the other database.';

CREATE INDEX IF NOT EXISTS codeintel_last_reconcile_last_reconcile_at_dump_id ON codeintel_last_reconcile(last_reconcile_at, dump_id);
