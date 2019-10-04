-- Drop the LSIF database. See up migration for details on dblink.
SELECT remote_exec('', 'DROP DATABASE IF EXISTS "' || current_database() || '_lsif";');

-- Drop dblink helper function
DROP FUNCTION IF EXISTS remote_exec(text, text);
