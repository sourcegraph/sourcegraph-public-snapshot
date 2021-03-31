# lsif-index-tester

## Project Setup

You should have a folder full of separate projects. Each project should have an `lsif_tests/` folder inside of it.

For example:

```
$ ls lsif-clang/functionaltests
project_1
project_2
```

Inside of each project, they should have:

- `setup_indexer.sh`

  - Bash script that setups up the project.
  - Can install dependencies, build project, etc.
  - This will be run every time before the indexer is run (so it can be worthwhile to make sure you cache the build if nothing changes)

- `lsif_tests/`
  - Within `lsif_tests/` you can have as many `.json` files as you'd like
  - Each `.json` file is a map that details different tests. See below for different support test tags.

### Test: `textDocument/definition`

```json
{
  "textDocument/definition": [
    {
      "name": "other_simple",
      "request": {
        "textDocument": "src/uses_header.c",
        "position": {
          "line": 6,
          "character": 8
        }
      },
      "response": {
        "textDocument": "src/uses_header.c",
        "range": {
          "start": { "line": 10, "character": 5 },
          "end": { "line": 10, "character": 22 }
        }
      }
    }
  ]
}
```

- `name`: String
  - The name of the test. Used when printing output
- `request`: `textDocument/request` request.
- `reseponse`: expected `textDocument/response` response.

See: [textDocument definition](https://microsoft.github.io/language-server-protocol/specification#textDocument_definition)

## Project Testing

```
go run . --indexer "lsif-clang compile_commands.json" --dir "/path/to/test_directory/"
```

- `--indexer` is the set of commands to actually run the indexer
- `--dir` is the root directory that contains an `lsif_tests` directory.
