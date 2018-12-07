BEGIN;
ALTER TABLE global_state ADD COLUMN mgmt_password_plaintext TEXT NOT NULL DEFAULT '';
ALTER TABLE global_state ADD COLUMN mgmt_password_bcrypt TEXT NOT NULL DEFAULT '';
END;
