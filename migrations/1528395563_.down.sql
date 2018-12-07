BEGIN;
ALTER TABLE global_state DROP COLUMN mgmt_password_plaintext;
ALTER TABLE global_state DROP COLUMN mgmt_password_bcrypt;
END;
