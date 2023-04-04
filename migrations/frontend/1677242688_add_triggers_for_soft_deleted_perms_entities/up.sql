CREATE OR REPLACE FUNCTION delete_user_repo_permissions_on_repo_soft_delete() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
    	DELETE FROM user_repo_permissions WHERE repo_id = NEW.id;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_repo_soft_delete ON repo;
CREATE TRIGGER trig_delete_user_repo_permissions_on_repo_soft_delete AFTER UPDATE ON repo FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_repo_soft_delete();

CREATE OR REPLACE FUNCTION delete_user_repo_permissions_on_external_account_soft_delete() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
    	DELETE FROM user_repo_permissions WHERE user_id = OLD.user_id AND user_external_account_id = OLD.id;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_external_account_soft_delete ON user_external_accounts;
CREATE TRIGGER trig_delete_user_repo_permissions_on_external_account_soft_delete AFTER UPDATE ON user_external_accounts FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_external_account_soft_delete();

CREATE OR REPLACE FUNCTION delete_user_repo_permissions_on_user_soft_delete() RETURNS trigger
	LANGUAGE plpgsql
  AS $$ BEGIN
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
    	DELETE FROM user_repo_permissions WHERE user_id = OLD.id;
    END IF;
    RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_user_soft_delete ON users;
CREATE TRIGGER trig_delete_user_repo_permissions_on_user_soft_delete AFTER UPDATE ON users FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_user_soft_delete();
