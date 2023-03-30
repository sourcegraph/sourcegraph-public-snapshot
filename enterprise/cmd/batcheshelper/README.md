# batcheshelper

`batcheshelper` is a helper script for Sourcegraph Batch Changes that is run in a workspace container before and after a
step is executed.

It is designed to replace Sourcegraph's `src` CLI for use in Executors.

## Usage

```shell
batcheshelper <pre|post> <step index> [OPTIONS]
OPTIONS:
  -input string
        The input JSON file for the workspace execution. Defaults to "input.json". (default "input.json")
  -previousStepPath string
        The path to the previous steps result file. Defaults to current working directory.
  -workspaceFiles string
        The path to the workspace files. Defaults to "/data/workspace-files". (default "/data/workspace-files")
```

### Arguments

| Argument | Place  | Description                       | Example Value         |
|----------|--------|-----------------------------------|-----------------------|
| Mode     | First  | The mode to run the script in.    | `pre` or `post`       |
| Step     | Second | The step that is being processed. | `0`, `1`, `5`, etc... |

### Options

| Flag                | Default Value           | Description                                                                  |
|---------------------|-------------------------|------------------------------------------------------------------------------|
| `-input`            | `input.json`            | The path to the input file. Defaults to "input.json". (default "input.json") |
| `-previousStepPath` | N/A                     | The path to the previous step's result file.                                 |
| `-workspaceFiles`   | `/data/workspace-files` | The path to the workspace files.                                             |

## Modes

There are two modes that this script can run in: `pre` and `post`.

### pre

The `pre` mode ...

```shell
batcheshelper pre 0
```

### post

The `post` mode ...

```shell
batcheshelper pre 0
```

## Build

To build the image for local development, run:

```shell
IMAGE=sourcegraph/batcheshelper:insiders ./build.sh
```
