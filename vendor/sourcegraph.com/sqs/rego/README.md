# rego

rego reinstalls and reruns a Go program when its source files change.

Usage:

```
go get -u sourcegraph.com/sqs/rego
rego [-v] [-race] import-path [optional args to program...]
```

Unlike [rerun](https://github.com/skelterjohn/rerun), it doesn't
recreate and rewatch all files after each change. It has fewer
features than rerun, though. I made it because my Mac was complaining
about having too many watched files since removing files from kevent
watchers wasn't working for some reason, and it was easier to write
this than dig into the kernel internals.
