# lazyzip

Package lazyzip provides support for reading ZIP archives. It is a fork of
`archive/zip`. It differs from `archive/zip` since it does not read the full
file listing into memory, instead it provides an interface similiar to
`archive/tar`.

See
[Reader.Next](https://godoc.org/github.com/sourcegraph/lazyzip#Reader.Next)
godoc.
