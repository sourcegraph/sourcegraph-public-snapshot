ALTER TABLE
    repo_permissions
ADD
    COLUMN IF NOT EXISTS unrestricted boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS repo_permissions_unrestricted_true_idx ON repo_permissions
    USING btree (unrestricted) WHERE unrestricted;

