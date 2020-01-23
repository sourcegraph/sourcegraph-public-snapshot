BEGIN;

-- Attempt to cast `val` as JSON and return false if there is a type error.
CREATE FUNCTION is_json(val varchar) RETURNS boolean AS $$ DECLARE temp json;
BEGIN
    BEGIN
        temp := val;
    EXCEPTION WHEN others THEN
        RETURN FALSE;
    END;
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Make arguments nullable instead of requiring an empty string
ALTER TABLE event_logs ALTER COLUMN argument DROP NOT NULL;
UPDATE event_logs SET argument = NULL WHERE argument = '';

-- Convert any current valid-json arguments into jsonb.
ALTER TABLE event_logs ALTER COLUMN argument TYPE jsonb USING CASE WHEN is_json(argument) THEN argument::jsonb ELSE NULL END;

-- Drop temp function
DROP FUNCTION is_json;

COMMIT;
