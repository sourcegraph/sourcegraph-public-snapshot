# Godep forks

This file lists all forked repositories that are vendored
under their original names. (For multi-package dependencies, it is
easier to keep a fork at the same import path than to rewrite the
import paths.)

* gopkg.in/gcfg.v1: github.com/sqs/gcfg branch `skip-unrecognized`
* gopkg.in/gorp.v1: github.com/sourcegraph/gorp branch `v1`
* github.com/spf13/hugo: github.com/sqs/hugo branch `vfs`
* github.com/drone/drone-exec: github.com/sourcegraph/drone-exec branch `dev`
* github.com/drone/drone-plugin-go: github.com/sqs/drone-plugin-go branch `multiple-netrc-entries`
* gopkg.in/olebedev/go-duktape.v2: github.com/sourcegraph/go-duktape branch `v2`
* github.com/NYTimes/gziphandler: using gzip.BestSpeed for ~1.8x speedup in many HTTP API endpoints

# Older versions

* github.com/gorilla/mux is pinned at the older version
  dfc482b2558f4151fc8598139a686caf3fb2ef65 due to the issues mentioned
  in https://github.com/gorilla/mux/issues/143. Our ./app/router tests
  fail on mux at latest master, so an erroneous update of this dep
  would be caught.
