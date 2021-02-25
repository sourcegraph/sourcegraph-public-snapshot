# Developing an out-of-band migration

Normal migrations apply a sequence of pure-SQL migrations (each in a transaction) to the database on application startup. This happens to block the new application version from reaching a working state until all migrations since the previous version have been applied. This is generally a non-issue as most schema migrations fast, and _most_ data migrations are fast enough to apply on application startup.

Out-of-band migrations allow for application-specific logic to exist in a migration that can't be easily performed within a SQL command. For example:

- generating key pairs
- re-hashing passwords
- decoding and interpreting opaque payloads
- fetching data from another remote API or data store based on existing data

Additionally, the estimated _scale_ of data may be too large to migrate on application startup in an efficient manner - in which case an out of band migration is suitable.

Remember - the longer we block application startup on migrations, the more vulnerable an instance will become to downtime as no _new_ frontend containers will be able to service requests. In these cases, you should define an _out of band_ migration, which is run in the background of the application over time instead of at startup.

## Overview

An out-of-band migration is defined with the following data:

- an owning _team_ (e.g. `codeintel`)
- an affected _component_ (e.g. `codeintel-db.lsif_data_documents`)
- a human-readable _description_ (e.g. `Populate num_diagnostics from gob-encoded payload`)
- the version when the migration was _introduced_ (e.g. `3.25.0`)
- the version when the migration was _deprecated_ (if any)
- a flag indicating whether or not the migration is _non-destructive_

Each out-of-band migration is associated with a _migrator_ instance, which periodically runs in the background of Sourcegraph instances between the version the migration was _introduced_ (inclusive) and the version the migration was marked as _deprecated_ (exclusive). Each migrator instance enables three behaviors:

- Perform a batch of the migration
- Perform a batch of the migration _in reverse_
- Determine the percentage of data that has been migrated

For every registered migrator, the migration runner will periodically check for the migration progress (even if it's finished). If the migrator has more work to do, then the runner will periodically have it perform a batch of the migration in the background. Errors that occur during migrations are logged to teh migration record and are exposed to the site-admin.

#### Upgrades

Site admins will be prevented from upgrading beyond the version where an incomplete migration has been deprecated. i.e., site-admin must wait for these migrations to finish before an upgrade. Otherwise, the instance will have data in a format that is no longer readable by the new version.

Today, this is not actually enforced - [in the future](https://github.com/sourcegraph/sourcegraph/issues/18240), we'll ensure that the instance gives plenty of warnings prior to an upgrade and ensure the instance does not start in such an invalid state.

#### Downgrades

If a migration is non-destructive (it only adds data), then it is safe to downgrade to a previous version. If a migration is destructive (it removes or changes the shape of data), then the migration must be run _in reverse_. We can do this by flipping the direction of the migration record. The runner takes care of the rest.

It is not valid for a user to downgrade beyond the version in which a non-0% migration was introduced. Site-admin must wait for these migrations to completely revert before a downgrade. Otherwise, the instance will have data i a format that is not yet readable by the old version.

#### Backwards compatibility

Because this happens in the background over time, the application must able to read both the new and old data formats until the migration has been deprecated. Advice for this is highly dependent on the type of data migration, but there are a few general options:

**Migrate to a new column**: Migrate data from `col` to `col_migrated` (a new column). If data is moved destructively (it may be expensive to keep both versions), then the down migration must be able to produce a payload in the previous shape to support downgrades. Progress can be determined by finding the ratio between the number of records with a non-null `col_migrated` and the total number of records. Reads can query both fields, and decode whichever field is populated. The old column can be dropped after the migration has been deprecated.

**Version each row**: Introduce a `schema_version` column (with a default of 1) to the table undergoing migration. Then, bump the schema_version on up migrations and decrement it on down migrations (along with other migration actions). Progress can be determined by finding the ratio between the number of records with `schema_version >= 2` and the total number of records. Reads can check the schema version of each row to decide how to interpret the raw data. Subsequent migrations on this table can use the same schema version column, but with higher numbers. This version column can be dropped after all migrations using that field have been deprecated.

## Adding a new migration

This guide will show you how to add a new out-of-band migration. We are migrating the following table:

```sql
CREATE TABLE skunk_payloads (
  id      serial primary key,
  payload text NOT NULL
);
```

#### Step 1: Add migration record

The first step is to register the migration in the database. This should be done with a normal in-band migration. If the data you are migrating needs additional columns or constraints, they can also be added in the same migration.

```sql
BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (
    42,                           -- This must be consistent across all Sourcegraph instances
    'skunkworks',                 -- Team owning migration
    'db.skunk_payloads',          -- Component being migrated
    'Re-encode our skunky data',  -- Description
    '3.28.0',                     -- The next minor release
    true                          -- Can be read with previous version without down migration
)
ON CONFLICT DO NOTHING;

-- we'll migrate payload -> payload2
ALTER TABLE skunk_payloads ADD COLUMN payload2 text;

COMMIT;
```

The migration record does not need to be _removed_ on database down migrations (hence the no-op `ON CONFLICT` clause in the up direction).

#### Step 2: Modify reads

Before we start writing records in a new format (either from migrated existing records or new writes), we need to ensure we can read the new and old formats. Consider queries with the given result shape.

```sql
SELECT payload, payload2 FROM skunk_payloads ...
```

We can create a scanner function to read this data as follows.

```go
func scanSkunkPayloads(rows *sql.Rows, queryErr error) ([]string, error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	var payloads []string
	for rows.Next() {
		var oldPayload, newPayload string
		if err := rows.Scan(&dbutil.NullString{S: &oldPayload}, &dbutil.NullString{S: &newPayload}); err != nil {
			return nil, err
		}

		if newPayload != "" {
			// Decode in new way
			payloads = append(payloads, decodeNew(newPayload))
		} else if oldPayload != nil {
			// Decode in old way
			payloads = append(payloads, decodeOld(oldPayload))
		}
	}

	return payloads, nil
}
```

#### Step 3: Modify writes

Now that we can read both formats, it is safe to start writing all _new_ records using the new format.

#### Step 4: Register migrator

Next, we need to move all of the existing data in the old format into the new format. We'll first define a migrator instance.

```go
import "github.com/sourcegraph/sourcegraph/internal/oobmigration"

type migrator struct {
	store *basestore.Store
}

func NewMigrator(store *basestore.Store) oobmigration.Migrator {
  return &migrator{store: store}
}
```

This migrator reports progress by counting the number of records with its new field populated over the total number of records (and special-cases empty tables as being completely converted as no records are in the old format).

```go
func (m *migrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(`
		SELECT
			CASE c2.count WHEN 0 THEN 1 ELSE 
				cast(c1.count as float) / cast(c2.count as float)
			END
		FROM
			(SELECT count(*) as count FROM skunk_payloads WHERE payload2 IS NOT NULL) c1
			(SELECT count(*) as count FROM skunk_payloads) c2
	`)))
	return progress, err
}
```

In the forward migration direction, we want to select a record that is in the previous format (we can tell here by the absence of a `payload2` field), and update that record with the result of some external computation. Here, we're going to rely on an oracle function `oldToNew` that converts the old format into the new format.

```go
const BatchSize = 500

func (m *migrator) Up(ctx context.Context) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	// Select and lock a single record within this transaction. This ensures
	// that many frontend instances can run the same migration concurrently
	// without them all trying to convert the same record.
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT id, payload FROM skunk_payloads WHERE payload2 IS NULL LIMIT %s FOR UPDATE SKIP LOCKED",
		BatchSize,
	))
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseROws(err) }()

	for rows.Next() {
		var id int
		var payload string
		if err := rows.Scan(&id, &payload); err != nil{
			return err
		}

		if err := tx.Exec(sqlf.Sprintf(
			"UPDATE skunk_payloads SET payload2 = %s WHERE id = %s", 
			oldToNew(payload), 
			ids,
		)); err != nil {
			return err
		}
	}

	return nil
}
```

In the reverse migration direction, we may think we can take a few short-cuts because we don't actually destroy any data in the `payload` field. However, it would be a mistake to simply remove the `payload2` data as we're only writing to `payload2` and **not** to the original `payload` column for new data. This would also be a problem if our up migration set `payload = NULL` on update to save space. Again, we rely on an oracle function `newToOld` that converts the new format into the old format.

```go
func (m *migrator) Down(ctx context.Context) (err error) {
  // Drop migrated records that still have the old payload
	if err := m.store.Exec(sqlf.Sprintf(
		"UPDATE skunk_payloads SET payload2 = NULL WHERE payload IS NOT NULL AND payload2 IS NOT NULL LIMIT %s",
		BatchSize,
	)); err != nil {
		return err
	}

	// For records that were written in the new format, we need to calculate
	// the old format from the new format. This should be exactly the inverse
	// of the forward migration.
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	rows, err := tx.Query(ctx, sqlf.Sprintf(
		"SELECT id, payload2 FROM skunk_payloads WHERE payload IS NULL LIMIT %s FOR UPDATE SKIP LOCKED",
		BatchSize,
	))
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseROws(err) }()

	for rows.Next() {
		var id int
		var payload string
		if err := rows.Scan(&id, &payload); err != nil{
			return err
		}

		if err := tx.Exec(sqlf.Sprintf(
			"UPDATE skunk_payloads SET payload = %s WHERE id = %s",
			newToOld(payload),
			ids,
		)); err != nil{
			return err
		}
	}

	return nil
}
```

Lastly, in order for this migration to run, we need to register it to the [outOfBandMigrator](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v3.25.0/-/blob/cmd/frontend/internal/cli/serve_cmd.go#L173:2). This instance is also available in the initialization functions for each component in the enterprise frontend.

```go
if err := outOfBandMigrationRunner.Register(42, NewMigrator(store), oobmigration.MigratorOptions{Interval: 3 * time.Second}); err != nil {
	return err
}
```

Here, we're telling the migration runner to invoke the `Up` or `Down` method periodically (once every three seconds) while the migration is active. The migrator batch size together with this interval is what controls the migration throughput.

#### Step 5: Mark deprecated

Once the engineering team has decided on which versions require the new format, old migrations can be marked with a concrete deprecation version. The deprecation version denotes the first Sourcegraph version that no longer runs the migration, and is no longer guaranteed to successfully read un-migrated records.

A new in-band migration can be created to update the deprecation field of the target record.

```sql
BEGIN;
UPDATE out_of_band_migrations SET deprecated = '3.38.0' WHERE id = 42;
COMMIT;
```

This date may be known at the time the migration is created, in which case it is fine to set both the introduced and the deprecated fields at the same time.

Note that it is not advised to set the deprecated version to the _next_ minor release of Sourcegraph. This will not given site-admins enough warning on the previous version that updating with an unfinished migration may cause issues at startup or data loss.

#### Step 6: Deprecation

On or after the deprecation version of a migration, we can begin clean-up. This involves:

- unregistering the migrator instance
- removing the migrator code
- cleaning up any backwards-compatible read routines to support only the new format
- dropping old unused columns that
