# Compact binary encoding for Go

The Go standard library package
[encoding/binary](http://golang.org/pkg/encoding/binary/) provides
encoding/decoding of *fixed-size* Go values or slices of same. This package
extends support to arbitrary, variable-sized values by prefixing these values
with their varint-encoded size, recursively. It expects the encoded type and
decoded type to match exactly and makes no attempt to reconcile or check for
any differences.
