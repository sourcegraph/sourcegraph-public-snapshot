# Bkstats 

A crude script to compute statistics from our Buildkite pipelines.

## Usage 

```
$ go run main.go -token BUILKITE_API_TOKEN -date 2021-10-22 -pipeline sourcegraph
# ...
On 2021-10-22, the pipeline was red for 1h8m32.856s
```

## Computation details

**Red** time is the duration from the _end of a failed build_ to the _end of the first subsequent green build_.
