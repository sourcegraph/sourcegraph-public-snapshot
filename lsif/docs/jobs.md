# LSIF worker jobs

The following job types are enqueued into [bull](https://github.com/OptimalBits/bull) and handled by the an instance of the worker process. Each job is enqueued with its job type, `name`, and its `args`.

### `convert({repository, commit, root, filepath})`

Convert an LSIF dump into a SQLite database and add cross-repository information into the cross-repository database (defined packages, imported references, and an LSIF data marker).

The repository, commit, and root arguments denote the portion of source code from which the LSIF upload was generated. The filepath argument denotes the path on disk where the LSIF upload data can be found. The file contents are assumed to be gzipped, and each line of the file contains a vertex or edge structure encoded as JSON.

### `clean-old-jobs({})`

Remove old job data from the system. This is based on a configurable age, `JOB_MAX_AGE`, within the worker process.
