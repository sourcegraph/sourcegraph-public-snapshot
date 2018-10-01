DELETE FROM repo WHERE uri = '' OR uri IS NULL;

ALTER TABLE repo
      ALTER COLUMN uri SET NOT NULL,
      ADD CONSTRAINT check_uri_nonempty CHECK (uri <> '');
