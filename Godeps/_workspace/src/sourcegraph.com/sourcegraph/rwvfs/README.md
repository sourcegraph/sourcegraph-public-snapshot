rwvfs
=====

Package rwvfs augments vfs to support write operations.

* [rwvfs documentation](https://sourcegraph.com/sourcegraph/rwvfs)
* [golang.org/x/tools/godoc/vfs documentation](http://godoc.org/golang.org/x/tools/godoc/vfs)

[![Build Status](https://travis-ci.org/sourcegraph/rwvfs.png)](https://travis-ci.org/sourcegraph/rwvfs)


## HTTP VFS

This repository includes HTTP server and client implementations of a
read-write virtual filesystem. The HTTP server is a Go http.Handler
that exposes any underlying RWVFS via HTTP. The HTTP client is a Go
implementation of rwvfs.FileSystem that communicates with an external
HTTP server. They both speak the same API on top of HTTP; see the code
for details.

To spin up a test client and server, follow these instructions. (They require Go and a `$GOPATH` set; as a quickstart, install the latest Go from [golang.org](http://golang.org) and then run `export GOPATH=$HOME; export PATH=$PATH:$GOPATH/bin`.)

**Server**

```
go get sourcegraph.com/sourcegraph/rwvfs/httpvfs
httpvfs &

go get sourcegraph.com/sourcegraph/rwvfs/httpvfs-client
httpvfs-client ls .
echo hello | httpvfs-client put foo
httpvfs-client cat foo
httpvfs-client rm foo
```


Future roadmap
----

* Add corresponding implementations of `rwvfs.FileSystem` for mapfs and zipfs
