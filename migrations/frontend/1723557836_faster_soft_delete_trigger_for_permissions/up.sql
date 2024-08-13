DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_repo_soft_delete ON repo;
CREATE TRIGGER trig_delete_user_repo_permissions_on_repo_soft_delete
  AFTER UPDATE OF deleted_at ON repo
  FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_repo_soft_delete();

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_external_account_soft_delete ON user_external_accounts;
CREATE TRIGGER trig_delete_user_repo_permissions_on_external_account_soft_delete
  AFTER UPDATE OF deleted_at ON user_external_accounts
  FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_external_account_soft_delete();

DROP TRIGGER IF EXISTS trig_delete_user_repo_permissions_on_user_soft_delete ON users;
CREATE TRIGGER trig_delete_user_repo_permissions_on_user_soft_delete
  AFTER UPDATE OF deleted_at ON users
  FOR EACH ROW EXECUTE FUNCTION delete_user_repo_permissions_on_user_soft_delete();
