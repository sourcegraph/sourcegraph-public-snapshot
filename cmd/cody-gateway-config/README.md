# Cody Gateway Config

This tool generates the JSON ModelConfiguration data which describes the set of
LLM models that Cody Gateway currently supports. This information is used by
Sourcegraph instances so they can pick up on any new LLM models as they get
released.

The output of this tool is a `.json` file which we check into this repo. It will
be embedded in Go binaries as a simple way of deploying the configuration blob.

## Usage

```zsh
# First do a dry run, and confirm any changes are reflected in the output.
go run ./cmd/cody-gateway-config -dry-run=true

# Run the tool and overwrite the existing file in your repo.
go run ./cmd/cody-gateway-config \
    -dry-run=false \
    -confirm-overwrite=true \
    -output-file="./internal/modelconfig/embedded/models.json"

# Run the unit tests which among other things verifies that the static
# ModelConfiguration document is well-formed.
go test ./cmd/frontend/internal/modelconfig/...
go test ./internal/modelconfig/...

# And finally, rerun the `sg lint` command, to ensure it is formatted
# as expected.
sg lint --fix format
```
