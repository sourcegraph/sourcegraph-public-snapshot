DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_repo_soft_delete ON repo;
CREATE TRIGGER trig_delete_user_repo_permissions_on_repo_soft_delete
  AFTER UPDATE OF deleted_at ON repo
  FOR EACH ROW
  WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
  EXECUTE FUNCTION delete_user_repo_permissions_on_repo_soft_delete();

-- This all runs in a transaction and acquires the table lock
-- ShareRowExclusiveLock, so to prevent deadlocks with other queries we commit
-- after each trigger change.
COMMIT AND CHAIN;

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_external_account_soft_delete ON user_external_accounts;
CREATE TRIGGER trig_delete_user_repo_permissions_on_external_account_soft_delete
  AFTER UPDATE OF deleted_at ON user_external_accounts
  FOR EACH ROW
  WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
  EXECUTE FUNCTION delete_user_repo_permissions_on_external_account_soft_delete();

-- See above comment
COMMIT AND CHAIN;

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_user_soft_delete ON users;
CREATE TRIGGER trig_delete_user_repo_permissions_on_user_soft_delete
  AFTER UPDATE OF deleted_at ON users
  FOR EACH ROW
  WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
  EXECUTE FUNCTION delete_user_repo_permissions_on_user_soft_delete();
