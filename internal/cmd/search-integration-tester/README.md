# Search Integration Tester

This utility helps keep high-level search functionality predictable with integration tests.

## Setup

Simply make sure these repos exist on your instance (In `Site Admin` -> `Manage repositories -> GitHub`:

```
  "repos": [
    "rvantonderp/beego-mux",
    "rvantonderp/adjust-go-wrk",
    "rvantonderp/sourcegraph-testing-caddy",
    "rvantonderp/auth0-go-jwt-middleware",
    "rvantonderp/DirectXMan12-k8s-prometheus-adapter",
    "rvantonderp/sgtest-mux",
    "facebookarchive/httpcontrol"
  ],
```

And set the following environment variables:

```
	endpoint  = env.Get("ENDPOINT", "http://127.0.0.1:3080", "Sourcegraph frontend endpoint")
	token     = env.Get("ACCESS_TOKEN", "", "Access token") // The access token configured in the instance site admin.
```

## Running

Simply run

```
go build && ./search-integration-tester
```

No output means all tests passed.

**Failed runs.** If a test fails, you'll see a diff of the new output, and the previous output. A test may fail for these reasons:

1. New functionality introduced a bug, in which case the bug should be fixed so that the test suite passes.
1. New functionality is correct and changed the output, in which case the test output should be updated. To update the output for the most recently failing test, run:

```
go build && ./search-integration-tester -update
```

1. It is the first time you are running this command, in which case you need to first generate the initial expected output. Do this by running:

```
go build && ./search-integration-tester -update-all
```

## Adding tests

Add test queries to `search_tests.go`, for example:

```
        {
                Name:  `Search timeout option, alert raised`,
                Query: `router index:no timeout:1ns`,
        },
```

- After adding a test, run `go build && ./search-integration-tester`. If the output looks good, run `./search-integration-tester -update`.
- Test names must be unique, it is used in the slug for a file path.
- Test against the existing repo set above unless there is a very good reason to add an additional repo. Keeping the repo set small helps to set things up and run tests quickly.
- If your test depends on something in the GQL result type that isn't currently queried, just add it to the GQL request in `search.go`.

## Example output

When a test fails, you'll see output that looks something like this, which mentions the test name, the query (which may be copy-pasted into the search bar to reproduce) and the GQL result type.

```
TEST FAILURE: Global search, double-quoted pattern, nonzero result
Query: "error type:\nasfjdjdafjdaf" patterntype:regexp count:1 stable:yes
  map[string]interface{}{
  	"data": map[string]interface{}{
  		"search": map[string]interface{}{
  			"results": map[string]interface{}{
  				"alert": nil,
  				"dynamicFilters": []interface{}{
- 					map[string]interface{}{"count": float64(1), "value": string("-file:(^|/)vendor/")},
- 					map[string]interface{}{"count": float64(1), "value": string("lang:go")},
- 					map[string]interface{}{
- 						"count": float64(1),
- 						"value": string(`repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$`),
- 					},
  				},
- 				"limitHit":   bool(true),
+ 				"limitHit":   bool(false),
- 				"matchCount": float64(1),
+ 				"matchCount": float64(0),
- 				"results": []interface{}{
- 					map[string]interface{}{
- 						"__typename": string("FileMatch"),
- 						"limitHit":   bool(false),
- 						"lineMatches": []interface{}{
- 							map[string]interface{}{
- 								"lineNumber":       float64(863),
- 								"offsetAndLengths": []interface{}{[]interface{}{float64(3), float64(11)}},
- 								"preview":          string("// Error type:"),
- 							},
- 						},
- 						"resource": string("git://github.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter#vendor/k8s.io/client-go/rest/request.go"),
- 					},
- 				},
+ 				"results": []interface{}{},
  			},
  		},
  	},
  }
```
