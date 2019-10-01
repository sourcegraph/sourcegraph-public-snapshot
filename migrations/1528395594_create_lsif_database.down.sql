SELECT dblink_exec('dbname=' || current_database() || ' user=' || current_user, 'DROP DATABASE sourcegraph_lsif;');
