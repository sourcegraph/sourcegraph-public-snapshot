To regenerate `lsif-go@foo.sql` you will need to:

1. `./dev/drop-entire-local-database-and-redis.sh`
2. Configure your `../dev-private` configuration to have:
   1. auto-indexing OFF (default)
   2. `"apidocs.search.indexing": "enabled",` under `experimentalFeatures`.
   3. `sourcegraph/lsif-go` as a repository.
3. Start the `sg start enterprise-codeintel` server
4. Run `lsif-go` (master revision) manually in a checkout of the `lsif-go` repository at the revision mentioned in the `lsif-go@rev.sql` filename.
5. Use `src lsif upload` against your local server to upload the LSIF bundle.
6. Run the following to generate the SQL file:

```
pg_dump --data-only  --format plain --inserts --dbname sg --table 'lsif_data_*' --file test.sql
```

Then rename `test.sql` as needed.
