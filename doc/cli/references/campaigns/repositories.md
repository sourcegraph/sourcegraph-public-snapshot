# `src campaigns repositories`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-f` | The campaign spec file to read. |  |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |


## Usage

```
Usage of 'src campaigns repositories':
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -f string
    	The campaign spec file to read.
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing

'src campaigns repositories' works out the repositories that a campaign spec
would apply to.

Usage:

    src campaigns repositories -f FILE

Examples:

    $ src campaigns repositories -f campaign.spec.yaml



```
	