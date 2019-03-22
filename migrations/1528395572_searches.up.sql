BEGIN;

CREATE TABLE searches (
	id serial PRIMARY KEY,
	-- pkg/version.Version()
	sourcegraph_version text NOT NULL,
	query text NOT NULL,
	created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION delete_old_rows() RETURNS trigger
	LANGUAGE plpgsql
	AS $$
BEGIN
	DELETE FROM searches WHERE id <= (SELECT MAX(id) FROM SEARCHES) - 100000;
	RETURN NULL;
END;
$$;

CREATE TRIGGER trigger_delete_old_rows
	AFTER INSERT ON searches
	EXECUTE PROCEDURE delete_old_rows();

COMMIT;
