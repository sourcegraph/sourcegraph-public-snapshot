CREATE OR REPLACE FUNCTION rockskip_get_symbol(repo_arg INTEGER, path_arg TEXT, name_arg TEXT, hops INTEGER[]) RETURNS INTEGER AS $$
DECLARE
    hop INTEGER;
    found_id INTEGER;
BEGIN
    FOREACH hop IN ARRAY hops LOOP
        SELECT id
            INTO found_id
            FROM rockskip_symbols
            WHERE
                repo_id    =  repo_arg AND
                path       =  path_arg AND
                name       =  name_arg AND
                ARRAY[hop] && added;

        IF FOUND THEN
            RETURN found_id;
        END IF;

        IF
            EXISTS (
                SELECT id
                FROM rockskip_symbols
                WHERE
                    repo_id    =  repo_arg AND
                    path       =  path_arg AND
                    name       =  name_arg AND
                    ARRAY[hop] && deleted
            )
            THEN
            RETURN -1;
        END IF;
    END LOOP;

    RETURN -1;
END; $$ IMMUTABLE language plpgsql;
