# `src orgs members add`


## Flags

| Name | Description | Default Value |
|------|-------------|---------------|
| `-dump-requests` | Log GraphQL requests and responses to stdout | `false` |
| `-get-curl` | Print the curl command for executing this query and exit (WARNING: includes printing your access token!) | `false` |
| `-org-id` | ID of organization to which to add member. (required) |  |
| `-trace` | Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing | `false` |
| `-username` | Username of user to add as member. (required) |  |


## Usage

```
Usage of 'src orgs members add':
  -dump-requests
    	Log GraphQL requests and responses to stdout
  -get-curl
    	Print the curl command for executing this query and exit (WARNING: includes printing your access token!)
  -org-id string
    	ID of organization to which to add member. (required)
  -trace
    	Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing
  -username string
    	Username of user to add as member. (required)

Examples:

  Add a member (alice) to an organization (abc-org):

    	$ src orgs members add -org-id=$(src org get -f '{{.ID}}' -name=abc-org) -username=alice



```
	