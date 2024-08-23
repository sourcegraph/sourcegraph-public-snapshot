# go-ctags: universal-ctags wrapper for easy access in Go

Note: This library is meant only for Sourcegraph use.

To improve `type:symbol` results in Sourcegraph,
for languages with high quality Tree-sitter grammars,
prefer adding support in `scip-ctags` in the Sourcegraph
monorepo over adding support in this repo.

## Adding new ctags flags

Add/modify appropriate `.ctags` files in `ctagsdotd`
and re-run `./gen.sh` at the root of the repository.

## Testing

Requires: `universal-ctags` on PATH. For exact version,
see the [CI workflow](.github/workflows/ci.yml).

Run tests: `go test ./...`

Update snapshots: `go test ./... -update`

Clean up old snapshots: `go test ./... -update -clean`
