# `src repos delete`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |


## Usage

```
Usage of 'src repos delete'
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing

Examples:

   Delete one or more repositories:

    	$ src repos delete github.com/my/repo github.com/my/repo2

```
	