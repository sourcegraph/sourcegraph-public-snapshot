-- +++
-- parent: 1528395908
-- +++

BEGIN;

-- Undo the changes corresponding to https://github.com/sourcegraph/sourcegraph/pull/25715
DELETE FROM out_of_band_migrations WHERE id=12 AND team='apidocs';

COMMIT;
