# How to use `migrator` operation `drift`

During an upgrade you may run into the following message.

```
* Sourcegraph migrator v4.1.3
❌ Schema drift detected for frontend
💡 Before continuing with this operation, run the migrator's drift command and follow instructions to repair the schema. See https://docs.sourcegraph.com/admin/how-to/manual_database_migrations#drift for additional instructions.
```

This error indicates that `migrator` has detected some difference between the state of the schema in your database and the expected schema for the database in the `-from` or current version of your Sourcegraph instance.

When the schema [drift](./manual_database_migrations.md#drift) command is run you'll see a set of diffs representing the areas where your instance schema has diverged from the expected state as well as the SQL operations to fix these examples of drift. For example:

```
❌ Missing index "external_service_repos"."external_service_repos_repo_id_external_service_id_unique"
💡 Suggested action: define the index.

ALTER TABLE external_service_repos ADD CONSTRAINT
external_service_repos_repo_id_external_service_id_unique UNIQUE
(repo_id, external_service_id);
```

```
❌ Unexpected properties of column "batch_spec_resolution_jobs"."batch_spec_id"

schemas.ColumnDescription{
  	Name:                   "batch_spec_id",
  	Index:                  -1,
  	TypeName:               "integer",
- 	IsNullable:             false,
+ 	IsNullable:             true,
  	Default:                "",
  	CharacterMaximumLength: 0,
  	... // 5 identical fields
  }

💡 Suggested action: change the column nullability constraint.

ALTER TABLE batch_spec_resolution_jobs ALTER COLUMN
batch_spec_id SET NOT NULL;
```

To correct these errors in the database run the suggested SQL queries via `psql` in internal databases, or via the tools provided by your cloud database provider. 

*docker example*
```
docker exec -it pgsql psql -U sg -c 'ALTER TABLE external_service_repos ADD CONSTRAINT external_service_repos_repo_id_external_service_id_unique UNIQUE (repo_id, external_service_id);'
```
*kubernetes example*
```
kubectl -n ns-sourcegraph exec -it pgsql -- psql -U sg -c 'ALTER TABLE external_service_repos ADD CONSTRAINT external_service_repos_repo_id_external_service_id_unique UNIQUE (repo_id, external_service_id);'
```

Then check the database again with the `drift` command and proceed with your multiversion upgrade.

> Note: It is possible for the drift command to detect diffs which will not prevent prevent upgrades.

If migrator drift suggests SQL queries which don't make sense please report to support@sourcegraph.com or open an issue in the [`sourcegraph/sourcegraph`](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=) repo.


