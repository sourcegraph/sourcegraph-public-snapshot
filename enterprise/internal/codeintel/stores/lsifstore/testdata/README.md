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

7. Rename `test.sql` as needed.
8. IMPORTANT: Remove these lines from the top of the file or else tests will fail:

```
--
-- PostgreSQL database dump
--

-- Dumped from database version 13.1
-- Dumped by pg_dump version 13.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
```
