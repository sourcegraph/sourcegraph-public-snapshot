BEGIN;

-- Default to now, since we can't have null times
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS invalidated_sessions_at timestamp with time zone DEFAULT now() NOT NULL;
-- Update the invalidated sessions at to be when the user was created
UPDATE users SET invalidated_sessions_at = created_at;

-- Create a procedure that invalidates sessions for the user that can be used for our trigger
-- Invalidates if anything password related is updated in the user table
CREATE OR REPLACE FUNCTION invalidate_session_for_userid_on_password_change_or_reset() RETURNS trigger
LANGUAGE plpgsql
    AS $$
    BEGIN
        IF (OLD.passwd != NEW.passwd OR OLD.passwd_reset_code != NEW.passwd_reset_code OR OLD.passwd_reset_time != NEW.passwd_reset_time) THEN
            NEW.invalidated_sessions_at = now();
            RETURN NEW;
        END IF;
    RETURN NEW;
    END;
$$;

-- Need to drop and create, since we can't create if not exstis
DROP TRIGGER IF EXISTS trig_invalidate_session_on_password_change_or_reset ON users;
-- Create a trigger to to invalidate sessions if the user's password is ever changed
CREATE TRIGGER trig_invalidate_session_on_password_change_or_reset BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE invalidate_session_for_userid_on_password_change_or_reset();

COMMIT;
