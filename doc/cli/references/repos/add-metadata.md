# `src repos add-metadata`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-insecure-skip-verify` | Skip validation of TLS certificates against trusted chains | `false` |
| `-key` | The name of the key to add (required) |  |
| `-repo` | The ID of the repo to add the key-value pair to (required) |  |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |
| `-user-agent-telemetry` | Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph | `true` |
| `-value` | The value associated with the key. Defaults to null. |  |


## Usage

```
Usage of 'src repos add-metadata':
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -insecure-skip-verify
    	Skip validation of TLS certificates against trusted chains
  -key string
    	The name of the key to add (required)
  -repo string
    	The ID of the repo to add the key-value pair to (required)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing
  -user-agent-telemetry
    	Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph (default true)
  -value string
    	The value associated with the key. Defaults to null.

Examples:

  Add a key-value pair metadata to a repository:

    	$ src repos add-metadata -repo=repoID -key=mykey -value=myvalue

  Omitting -value will create a tag (a key with a null value).


```
	