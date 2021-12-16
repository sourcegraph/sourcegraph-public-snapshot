# Bkstats

A crude script to compute statistics from our Buildkite pipelines.
Owned by the [DevX team](https://handbook.sourcegraph.com/departments/product-engineering/engineering/enablement/dev-experience).

## Usage

```
$ go run main.go -token $BUILDKITE_API_TOKEN -date 2021-10-22 -pipeline sourcegraph
# ...
On 2021-10-22, the pipeline was red for 1h8m32.856s
```

### Buildkite API token

1. Go over [your personal settings](https://buildkite.com/user/api-access-tokens)
2. Create a new token with the following permissions:

- check `sourcegraph` organization
- `read_builds`
- `read_pipelines`

## Computation details

**Red** time is the duration from the _end of a failed build_ to the _end of the first subsequent green build_.
