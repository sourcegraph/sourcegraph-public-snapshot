# LSIF development and testing utilities

## Install

Assumes a working Go installation:

```
# lsif-validate
go get github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/cmd/lsif-validate

# lsif-visualize
go get github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/cmd/lsif-visualize
```

Resulting binary should then be in your `$GOPATH/bin` (conventionally `$HOME/go/bin`), so make sure thats in your `$PATH` or else invoke using absolute/relative location.

Binary releases coming soon™️

## lsif-validate

This command validates the output of an LSIF indexer. The following properties are validated:

- Element IDs are unique
- All references of element occur after its definition
- A single metadata vertex exists and is the firsts element in the dump
- The project root is a valid URL
- Each document URI is a URL relative to the project root
- Each range vertex has sane bounds (non-negative line/character values and the ending position occurs strictly after the starting position)
- 1-to-n edges have a non-empty `inVs` array
- Edges refer to identifiers attached to the correct element type, as follows:

  | label                     | inV(s)               | outV                | condition                      |
  | ------------------------- | -------------------- | ------------------- | ------------------------------ |
  | `contains`                | `range`              |                     | if outV is a `document`        |
  | `item`                    | `range`              |                     |                                |
  | `item`                    | `referenceResult`    |                     | if outV is a `referenceResult` |
  | `next`                    | `resultSet`          | `range`/`resultSet` |                                |
  | `textDocument/definition` | `definitionResult`   | `range`/`resultSet` |                                |
  | `textDocument/references` | `referenceResult`    | `range`/`resultSet` |                                |
  | `textDocument/hover`      | `hoverResult`        | `range`/`resultSet` |                                |
  | `moniker`                 | `moniker`            | `range`/`resultSet` |                                |
  | `nextMoniker`             | `moniker`            | `moniker`           |                                |
  | `packageInformation`      | `packageInformation` | `moniker`           |                                |

- Each vertex is reachable from a range or document vertex (_ignored: metadata, project, document, and event vertices_)
- Each range belongs to a unique document
- No two ranges belonging to the same document improperly overlap
- The inVs of each `item` edge belong to that document referred to by the edge's `document` field
