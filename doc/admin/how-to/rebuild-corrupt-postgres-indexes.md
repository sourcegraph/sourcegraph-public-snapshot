# How to rebuild corrupt Postgres indexes after upgrading to 3.30 or 3.30.1

## Background

The 3.30 release introduced a `pgsql` and `codeinteldb` base image change from debian to alpine which changed the default OS locale.
This caused corruption in indexes that have collatable key columns (e.g. any index with a `text` column). Read more about this here: https://postgresql.verite.pro/blog/2018/08/27/glibc-upgrade.html

After we found the root-cause of the [issues many customers were seeing](https://github.com/sourcegraph/sourcegraph/issues/23288), we cut [a patch release](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md#3303) that reverted the images to be based on debian, buying us time to change the alpine based version of the images to [reindex affected indexes on startup, before accepting new connections](https://github.com/sourcegraph/sourcegraph/issues/23310).

However, those customers that had already upgraded need assistance in fixing their already corrupt databases. Below is a guide to do so.

## Recovery guide

### Rebuild indexes in the main sourcegraph db

We need to ensure there's nothing writing or reading from/to the database before performing the next steps.

In Kubernetes, you can accomplish this by deleting the database service to prevent new connections from being established, followed by a query to kill existing connections.

```shell
kubectl delete svc/pgsql
kubectl port-forward deploy/pgsql 3333:5432 # doesn't use the service that we just deleted
psql -U sg -d sg -h localhost -p 3333
```

Kill existing client connections first:

```sql
select pg_terminate_backend(pg_stat_activity.pid)
from pg_stat_activity where datname = 'sg'
```

With a Postgres client connected to the database, we now start by reindexing system catalog indexes which may have been affected.

```sql
reindex (verbose) system sg;
```

Now we need to rebuild the sourcegraph db indexes. We execute each of these sequentially until the first failure, most likely due to duplicates, at which point we must delete those duplicates before trying again.
You can find an example duplicate deletion query further down in this file. After deleting those duplicates, just resume the index rebuilding from where we left off, commenting out or deleting the previouse reindex statements.
We prefer to reindex explicit indexes rather that the whole database in order to allow resuming where we left off after encountering a duplicates error and fixing it.

```sql
reindex (verbose) index batch_changes_site_credentials_unique;
reindex (verbose) index batch_spec_executions_rand_id;
reindex (verbose) index batch_specs_rand_id;
reindex (verbose) index changeset_events_changeset_id_kind_key_unique;
reindex (verbose) index changeset_jobs_bulk_group_idx;
reindex (verbose) index changeset_jobs_state_idx;
reindex (verbose) index changeset_specs_external_id;
reindex (verbose) index changeset_specs_head_ref;
reindex (verbose) index changeset_specs_rand_id;
reindex (verbose) index changeset_specs_title;
reindex (verbose) index changesets_external_state_idx;
reindex (verbose) index changesets_external_title_idx;
reindex (verbose) index changesets_publication_state_idx;
reindex (verbose) index changesets_reconciler_state_idx;
reindex (verbose) index changesets_repo_external_id_unique;
reindex (verbose) index discussion_mail_reply_tokens_pkey;
reindex (verbose) index discussion_threads_target_repo_repo_id_path_idx;
reindex (verbose) index event_logs_anonymous_user_id;
reindex (verbose) index event_logs_name;
reindex (verbose) index event_logs_source;
reindex (verbose) index external_service_sync_jobs_state_idx;
reindex (verbose) index feature_flag_overrides_unique_org_flag;
reindex (verbose) index feature_flag_overrides_unique_user_flag;
reindex (verbose) index feature_flags_pkey;
reindex (verbose) index gitserver_repos_last_error_idx;
reindex (verbose) index insights_query_runner_jobs_state_btree;
reindex (verbose) index kind_cloud_default;
reindex (verbose) index lsif_packages_scheme_name_version_dump_id;
reindex (verbose) index lsif_references_scheme_name_version_dump_id;
reindex (verbose) index lsif_uploads_repository_id_commit_root_indexer;
reindex (verbose) index lsif_uploads_state;
reindex (verbose) index names_pkey;
reindex (verbose) index orgs_name;
reindex (verbose) index phabricator_repos_repo_name_key;
reindex (verbose) index registry_extension_releases_registry_extension_id;
reindex (verbose) index registry_extension_releases_version;
reindex (verbose) index registry_extensions_publisher_name;
reindex (verbose) index repo_external_unique_idx;
reindex (verbose) index repo_name_unique;
reindex (verbose) index repo_pending_permissions_perm_unique;
reindex (verbose) index repo_permissions_perm_unique;
reindex (verbose) index repo_uri_idx;
reindex (verbose) index search_context_repos_search_context_id_repo_id_revision_unique;
reindex (verbose) index search_contexts_name_namespace_org_id_unique;
reindex (verbose) index search_contexts_name_namespace_user_id_unique;
reindex (verbose) index search_contexts_name_without_namespace_unique;
reindex (verbose) index security_event_logs_anonymous_user_id;
reindex (verbose) index security_event_logs_name;
reindex (verbose) index security_event_logs_source;
reindex (verbose) index user_credentials_domain_user_id_external_service_type_exter_key;
reindex (verbose) index user_emails_no_duplicates_per_user;
reindex (verbose) index user_emails_unique_verified_email;
reindex (verbose) index user_external_accounts_account;
reindex (verbose) index user_pending_permissions_service_perm_object_unique;
reindex (verbose) index user_permissions_perm_object_unique;
reindex (verbose) index users_billing_customer_id;
reindex (verbose) index users_username;
reindex (verbose) index versions_pkey;
```

The above statements were produced by this query. **No need to run this, it's just informational.**

```sql
-- This query lists all indexes in the database, excluding catalogue indexes,
-- that have key columns of collatable types. In other words, it lists all indexes
-- that need to be rebuilt. You don't need to run this query since we have done that
-- for you, with the output below.
select
    distinct('reindex (verbose) index ' || i.relname || ';') as stmt
from
    pg_class t,
    pg_class i,
    pg_index ix,
    pg_attribute a,
    pg_namespace n
where
    t.oid = ix.indrelid
    and i.oid = ix.indexrelid
    and n.oid = i.relnamespace
    and a.attrelid = t.oid
    and a.attnum = ANY(ix.indkey)
    and t.relkind = 'r'
    and n.nspname = 'public'
    and ix.indcollation != oidvectorin(repeat('0 ', ix.indnkeyatts)::cstring)
order by stmt;
```

As for the _duplicate deletion query_, here's an example of for the `repo` table that needs to be adapated to the specific table and index we need to remove duplicates in.

```sql
begin;

-- We must disable index scans before deleting so that we avoid
-- using the corrupt indexes to find the rows to delete. The database then
-- does a sequential scan, which is what we want in order to accomplish that.

set enable_indexscan = 'off';
set enable_bitmapscan = 'off';

delete from repo t1
using repo t2
where t1.ctid > t2.ctid
and (
  t1.name = t2.name or
  (
    t1.external_service_type = t2.external_service_type and
    t1.external_service_id = t2.external_service_id and
    t1.external_id = t2.external_id
  )
);

commit;
```

At the very end of the index rebuilding process, as a last sanity check, we use the amcheck extension to verify there are no corrupt indexes.

```sql
create extension amcheck;

select bt_index_parent_check(c.oid, true), c.relname, c.relpages
from pg_index i
join pg_opclass op ON i.indclass[0] = op.oid
join pg_am am ON op.opcmethod = am.oid
join pg_class c ON i.indexrelid = c.oid
join pg_namespace n ON c.relnamespace = n.oid
where am.amname = 'btree'
-- Don't check temp tables, which may be from another session:
and c.relpersistence != 't'
-- Function may throw an error when this is omitted:
and i.indisready AND i.indisvalid;
```

### Rebuild indexes in the codeintel db

We need to ensure there's nothing writing or reading from/to the database before performing the next steps.

In Kubernetes, you can accomplish this by deleting the database service to prevent new connections from being established, followed by a query to kill existing connections.

```shell
kubectl delete svc/codeintel-db
kubectl port-forward deploy/codeintel-db 3333:5432 # doesn't use the service that we just deleted
psql -U sg -d sg -h localhost -p 3333
```

Kill existing client connections first:

```sql
select pg_terminate_backend(pg_stat_activity.pid)
from pg_stat_activity where datname = 'sg'
```

With a Postgres client connected to the database, we now start by reindexing system catalog indexes which may have been affected.

```sql
reindex (verbose) system sg;
```

Now we need to rebuild the codeintel db indexes. We execute each of these sequentially until the first failure, most likely due to duplicates, at which point we must delete those duplicates before trying again.
You can find an example duplicate deletion query above in this file. After deleting those duplicates, just resume the index rebuilding from where we left off, commenting out or deleting the previouse reindex statements.
We prefer to reindex explicit indexes rather that the whole database in order to allow resuming where we left off after encountering a duplicates error and fixing it.

```sql
reindex (verbose) index lsif_data_definitions_pkey;
reindex (verbose) index lsif_data_documentation_mappings_pkey;
reindex (verbose) index lsif_data_documentation_pages_pkey;
reindex (verbose) index lsif_data_documentation_path_info_pkey;
reindex (verbose) index lsif_data_documents_pkey;
reindex (verbose) index lsif_data_references_pkey;
```

At the very end of the index rebuilding process, as a last sanity check, we use the amcheck extension to verify there are no corrupt indexes.

```sql
create extension amcheck;

select bt_index_parent_check(c.oid, true), c.relname, c.relpages
from pg_index i
join pg_opclass op ON i.indclass[0] = op.oid
join pg_am am ON op.opcmethod = am.oid
join pg_class c ON i.indexrelid = c.oid
join pg_namespace n ON c.relnamespace = n.oid
where am.amname = 'btree'
-- Don't check temp tables, which may be from another session:
and c.relpersistence != 't'
-- Function may throw an error when this is omitted:
and i.indisready AND i.indisvalid;
```

After the indexes have been rebuilt and the index integrity query doesn't return any errors, we can start all Sourcegraph services again. The way you do this is dependent on your specific deployment.
