# Radix

[![Build Status](https://travis-ci.org/mediocregopher/radix.v2.svg)](https://travis-ci.org/mediocregopher/radix.v2)
[![GoDoc](https://godoc.org/github.com/mediocregopher/radix.v2?status.svg)](https://godoc.org/github.com/mediocregopher/radix.v2)

Radix is a minimalistic [Redis][redis] client for Go. It is broken up into
small, single-purpose packages for ease of use.

* [redis](http://godoc.org/github.com/mediocregopher/radix.v2/redis) - A wrapper
  around a single, *non-thread-safe* redis connection. Supports normal
  commands/response as well as pipelining.

* [pool](http://godoc.org/github.com/mediocregopher/radix.v2/pool) - a simple,
  automatically expanding/cleaning connection pool. If you have multiple
  go-routines using the same redis instance you'll need this.

* [pubsub](http://godoc.org/github.com/mediocregopher/radix.v2/pubsub) - a
  simple wrapper providing convenient access to Redis Pub/Sub functionality.

* [sentinel](http://godoc.org/github.com/mediocregopher/radix.v2/sentinel) - a
  client for [redis sentinel][sentinel] which acts as a connection pool for a
  cluster of redis nodes. A sentinel client connects to a sentinel instance and
  any master redis instances that instance is monitoring. If a master becomes
  unavailable, the sentinel client will automatically start distributing
  connections from the slave chosen by the sentinel instance.

* [cluster](http://godoc.org/github.com/mediocregopher/radix.v2/cluster) - a
  client for a [redis cluster][cluster] which automatically handles interacting
  with a redis cluster, transparently handling redirects and pooling. This
  client keeps a mapping of slots to nodes internally, and automatically keeps
  it up-to-date.

* [util](http://godoc.org/github.com/mediocregopher/radix.v2/util) - a
  package containing a number of helper methods for doing common tasks with the
  radix package, such as SCANing either a single redis instance or every one in
  a cluster, or executing server-side lua

## Installation

    go get github.com/mediocregopher/radix.v2/...

## Testing

    go test github.com/mediocregopher/radix.v2/...

The test action assumes you have the following running:

* A redis server listening on port 6379

* A redis cluster node listening on port 7000, handling slots 0 through 8191

* A redis cluster node listening on port 7001, handling slots 8192 through 16383

* A redis server listening on port 8000

* A redis server listening on port 8001, slaved to the one on 8000

* A redis sentinel listening on port 28000, watching the one on port 8000 as a
  master named `test`.

The slot number are particularly important as the tests for the cluster package
do some trickery which depends on certain keys being assigned to certain nodes

You can do `make start` and `make stop` to automatically start and stop a test
environment matching these requirements.

## Why is this V2?

V1 of radix was started by [fzzy](https://github.com/fzzy) and can be found
[here](https://github.com/fzzy/radix). Some time in 2014 I took over the project
and reached a point where I couldn't make improvements that I wanted to make due
to past design decisions (mostly my own). So I've started V2, which has
redesigned some core aspects of the api and hopefully made things easier to use
and faster.

Here are some of the major changes since V1:

* Combining resp and redis packages

* Reply is now Resp

* Hash is now Map

* Append is now PipeAppend, GetReply is now PipeResp

* PipelineQueueEmptyError is now ErrPipelineEmpty

* Significant changes to pool, making it easier to use

* More functionality in cluster

## Copyright and licensing

Unless otherwise noted, the source files are distributed under the *MIT License*
found in the LICENSE.txt file.

[redis]: http://redis.io
[sentinel]: http://redis.io/topics/sentinel
[cluster]: http://redis.io/topics/cluster-spec
