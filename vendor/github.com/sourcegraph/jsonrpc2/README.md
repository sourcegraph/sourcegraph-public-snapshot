# jsonrpc2: JSON-RPC 2.0 implementation for Go [![Build Status](https://travis-ci.org/sourcegraph/jsonrpc2.svg)](https://travis-ci.org/sourcegraph/jsonrpc2)

Package jsonrpc2 provides a [Go](https://golang.org) implementation of [JSON-RPC 2.0](http://www.jsonrpc.org/specification).

This package is **experimental** until further notice.

[**Open the code in Sourcegraph**](https://sourcegraph.com/github.com/sourcegraph/jsonrpc2)

## Known issues

* Batch requests and responses are not yet supported. A handler will panic if it receives a batch request. Because of this, you should not expose any server using this package to external, untrusted traffic (yet).
