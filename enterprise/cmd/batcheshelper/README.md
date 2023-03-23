# batcheshelper

This image is used to prepare a Job for running a batch change.

## Usage

```shell
batcheshelper <pre|post> <step index> [OPTIONS]
OPTIONS:
  -input string
        The path to the input file. Defaults to "input.json". (default "input.json")
```

### Arguments

| Argument | Place  | Description                       | Example Value         |
|----------|--------|-----------------------------------|-----------------------|
| Mode     | First  | The mode to run the script in.    | `pre` or `post`       |
| Step     | Second | The step that is being processed. | `0`, `1`, `5`, etc... |

### Options

| Flag     | Default Value | Description                                                                  |
|----------|---------------|------------------------------------------------------------------------------|
| `-input` | `input.json`  | The path to the input file. Defaults to "input.json". (default "input.json") |

## Modes

There are two modes that this script can run in: `pre` and `post`.

### pre

```shell
batcheshelper pre 0
```

### post

```shell
batcheshelper pre 0
```

## Building

## Usage in Executors
