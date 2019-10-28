-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

ALTER TABLE lsif_dumps ADD COLUMN uploaded_at timestamp with time zone NOT NULL DEFAULT now();
CREATE INDEX lsif_dumps_uploaded_at ON lsif_dumps(uploaded_at);
CREATE INDEX lsif_dumps_visible_repository_commit ON lsif_dumps(repository, "commit") WHERE visible_at_tip;

COMMIT;
