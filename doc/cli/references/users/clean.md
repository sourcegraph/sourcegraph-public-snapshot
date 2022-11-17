# `src users clean`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-days` | Days threshold on which to remove users, must be 60 days or greater and defaults to this value  | `60` |
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-force` | skips user confirmation step allowing programmatic use | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-insecure-skip-verify` | Skip validation of TLS certificates against trusted chains | `false` |
| `-remove-admin` | clean admin accounts | `false` |
| `-remove-never-active` | removes users with null lastActive value | `false` |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |
| `-user-agent-telemetry` | Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph | `true` |


## Usage

```
Usage of 'src users clean':
  -days int
    	Days threshold on which to remove users, must be 60 days or greater and defaults to this value  (default 60)
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -force
    	skips user confirmation step allowing programmatic use
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -insecure-skip-verify
    	Skip validation of TLS certificates against trusted chains
  -remove-admin
    	clean admin accounts
  -remove-never-active
    	removes users with null lastActive value
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing
  -user-agent-telemetry
    	Include the operating system and architecture in the User-Agent sent with requests to Sourcegraph (default true)

This command removes users from a Sourcegraph instance who have been inactive for 60 or more days. Admin accounts are omitted by default.
	
Examples:

	$ src users clean -days 182
	
	$ src users clean -remove-admin -remove-never-active 


```
	