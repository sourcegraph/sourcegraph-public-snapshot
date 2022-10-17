CREATE OR REPLACE FUNCTION merge_audit_log_transitions(internal hstore, arrayhstore hstore[]) RETURNS hstore AS $$
    DECLARE
        trans hstore;
    BEGIN
      FOREACH trans IN ARRAY arrayhstore
      LOOP
          internal := internal || hstore(trans->'column', trans->'new');
      END LOOP;

      RETURN internal;
    END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE AGGREGATE snapshot_transition_columns(HSTORE[]) (
    SFUNC = merge_audit_log_transitions,
    STYPE = HSTORE,
    INITCOND = ''
);
