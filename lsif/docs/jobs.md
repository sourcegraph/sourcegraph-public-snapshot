# LSIF worker jobs

The following job types are enqueued into [bull](https://github.com/OptimalBits/bull) and handled by the an instance of the worker process. Each job is enqueued with a `name`, which determines how the job is performed, and an `args` object.

### `convert({repository, commit, root, filepath})`

Convert an LSIF dump into a SQLite database and add cross-repository information into the cross-repository database (defined packages, imported references, and an LSIF data marker). The dumped file contents are assumed to be gzipped, and each line of the file contains a vertex or edge structure encoded as JSON.

| Argument   | Description                                               |
| ---------- | --------------------------------------------------------- |
| repository | The repository name.                                      |
| commit     | The 40-character commit hash.                             |
| root       | The directory from which LSIF data was generated.         |
| filepath   | The path on disk where the LSIF upload data can be found. |

### `update-tips({})`

Fetch the tip of the default branch for each repository with LSIF data from gitserver and update the `lsif_dumps` table.

### `clean-old-jobs({})`

Remove old job data from the system. This is based on a configurable age, `JOB_MAX_AGE`, within the worker process.
