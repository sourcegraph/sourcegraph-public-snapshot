-- +goose Up
CREATE INDEX idx_reverse_lookup_user on tuple (store, object_type, relation, _user);

-- +goose Down
DROP INDEX IF EXISTS idx_reverse_lookup_user ;