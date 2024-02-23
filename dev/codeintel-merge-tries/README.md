A little program to check how many rows we could save by merging tries inside our
`codeintel_scip_symbol_names` table.

After uploading an index to your local sourcegraph instance, run like so:
`psql -d sourcegraph -c "COPY (select * from codeintel_scip_symbol_names where upload_id = (select max(upload_id) from codeintel_scip_symbol_names)) TO STDOUT DELIMITER ',' CSV HEADER;" | cargo run --release`

For example when running after uploading an index for sourcegraph/sourcegraph:

```
$ psql --no-psqlrc --csv --dbname=sourcegraph --command="select * from codeintel_scip_symbol_names where upload_id = (select max(upload_id) from codeintel_scip_symbol_names)" | cargo run --release
   Compiling merge-tries v0.1.0 (/Users/creek/work/merge-tries)
    Finished release [optimized] target(s) in 1.96s
     Running `target/release/merge-tries`
from 125334 rows
to   100428 rows
reduction by 19.87%
```
