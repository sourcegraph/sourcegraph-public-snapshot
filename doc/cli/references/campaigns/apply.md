# `src campaigns apply`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-allow-unsupported` | Allow unsupported code hosts. | `false` |
| `-apply` | Ignored. | `false` |
| `-cache` | Directory for caching results and repository archives. | `~/.cache/sourcegraph/campaigns` |
| `-clean-archives` | If true, deletes downloaded repository archives after executing campaign steps. | `true` |
| `-clear-cache` | If true, clears the execution cache and executes all steps anew. | `false` |
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-f` | The campaign spec file to read. |  |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-j` | The maximum number of parallel jobs. Default is GOMAXPROCS. | `8` |
| `-keep-logs` | Retain logs after executing steps. | `false` |
| `-n` | Alias for -namespace. |  |
| `-namespace` | The user or organization namespace to place the campaign within. Default is the currently authenticated user. |  |
| `-skip-errors` | If true, errors encountered while executing steps in a repository won't stop the execution of the campaign spec but only cause that repository to be skipped. | `false` |
| `-timeout` | The maximum duration a single set of campaign steps can take. | `1h0m0s` |
| `-tmp` | Directory for storing temporary data, such as log files. Default is /tmp. Can also be set with environment variable SRC_CAMPAIGNS_TMP_DIR; if both are set, this flag will be used and not the environment variable. | `/tmp` |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |
| `-v` | print verbose output | `false` |
| `-workspace` | Workspace mode to use ("auto", "bind", or "volume") | `auto` |


## Usage

```
Usage of 'src campaigns apply':
  -allow-unsupported
    	Allow unsupported code hosts.
  -apply
    	Ignored.
  -cache string
    	Directory for caching results and repository archives. (default "~/.cache/sourcegraph/campaigns")
  -clean-archives
    	If true, deletes downloaded repository archives after executing campaign steps. (default true)
  -clear-cache
    	If true, clears the execution cache and executes all steps anew.
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -f string
    	The campaign spec file to read.
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -j int
    	The maximum number of parallel jobs. Default is GOMAXPROCS. (default 8)
  -keep-logs
    	Retain logs after executing steps.
  -n string
    	Alias for -namespace.
  -namespace string
    	The user or organization namespace to place the campaign within. Default is the currently authenticated user.
  -skip-errors
    	If true, errors encountered while executing steps in a repository won't stop the execution of the campaign spec but only cause that repository to be skipped.
  -timeout duration
    	The maximum duration a single set of campaign steps can take. (default 1h0m0s)
  -tmp string
    	Directory for storing temporary data, such as log files. Default is /tmp. Can also be set with environment variable SRC_CAMPAIGNS_TMP_DIR; if both are set, this flag will be used and not the environment variable. (default "/tmp")
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing
  -v	print verbose output
  -workspace string
    	Workspace mode to use ("auto", "bind", or "volume") (default "auto")

'src campaigns apply' is used to apply a campaign spec on a Sourcegraph
instance, creating or updating the described campaign if necessary.

Usage:

    src campaigns apply -f FILE [command options]

Examples:

    $ src campaigns apply -f campaign.spec.yaml
  
    $ src campaigns apply -f campaign.spec.yaml -namespace myorg



```
	