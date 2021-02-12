# `src search`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-explain-json` | Explain the JSON output schema and exit. | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-json` | Whether or not to output results as JSON | `false` |
| `-less` | Pipe output to 'less -R' (only if stdout is terminal, and not json flag) | `true` |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |


## Usage

```
Usage of 'src search':
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -explain-json
    	Explain the JSON output schema and exit.
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -json
    	Whether or not to output results as JSON
  -less
    	Pipe output to 'less -R' (only if stdout is terminal, and not json flag) (default true)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing

Examples:

  Perform a search and get results:

    	$ src search 'repogroup:sample error'

  Perform a search and get results as JSON:

    	$ src search -json 'repogroup:sample error'

Other tips:

  Make 'type:diff' searches have colored diffs by installing https://colordiff.org
    - Ubuntu/Debian: $ sudo apt-get install colordiff
    - Mac OS:        $ brew install colordiff
    - Windows:       $ npm install -g colordiff

  Disable color output by setting NO_COLOR=t (see https://no-color.org).

  Force color output on (not on by default when piped to other programs) by setting COLOR=t

  Query syntax: https://about.sourcegraph.com/docs/search/query-syntax/


```
	