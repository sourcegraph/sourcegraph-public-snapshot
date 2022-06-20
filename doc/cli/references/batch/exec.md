# `src batch exec`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-cache` | Directory to read cached results from. |  |
| `-f` | The workspace execution input file to read. |  |
| `-repo` | Path of the checked out repo on disk. |  |
| `-sourcegraphVersion` | Sourcegraph backend version. |  |
| `-timeout` | The maximum duration a single batch spec step can take. | `1h0m0s` |
| `-tmp` | Directory for storing temporary data. |  |


## Usage

```
Usage of 'src batch exec':
  -cache string
    	Directory to read cached results from.
  -f string
    	The workspace execution input file to read.
  -repo string
    	Path of the checked out repo on disk.
  -sourcegraphVersion string
    	Sourcegraph backend version.
  -timeout duration
    	The maximum duration a single batch spec step can take. (default 1h0m0s)
  -tmp string
    	Directory for storing temporary data.

INTERNAL USE ONLY: 'src batch exec' executes the given raw batch spec in the given workspaces.

The input file contains a JSON dump of the WorkspacesExecutionInput struct in
github.com/sourcegraph/sourcegraph/lib/batches.

Usage:

    src batch exec -f FILE -repo DIR [command options]

Examples:

    $ src batch exec -f batch-spec-with-workspaces.json



```
	