BEGIN;

-- Attempt to cast `val` as JSON and return false if there is a type error.
CREATE FUNCTION is_json(val text) RETURNS boolean AS $$
DECLARE temp json;
BEGIN
    temp := val;
    RETURN TRUE;
EXCEPTION WHEN others THEN
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Make arguments nullable instead of requiring an empty string
ALTER TABLE event_logs ALTER COLUMN argument DROP NOT NULL;
UPDATE event_logs SET argument = NULL WHERE argument = '';

-- Convert any current valid-json arguments into jsonb.
ALTER TABLE event_logs ALTER COLUMN argument TYPE jsonb USING CASE WHEN is_json(argument) THEN argument::jsonb ELSE NULL END;

-- Drop temp function
DROP FUNCTION is_json(text);

COMMIT;
