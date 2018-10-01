ALTER TABLE repo
      ALTER COLUMN uri DROP NOT NULL,
      DROP CONSTRAINT check_uri_nonempty;
