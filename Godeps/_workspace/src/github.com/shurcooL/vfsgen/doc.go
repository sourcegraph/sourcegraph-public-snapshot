/*
Package vfsgen generates a vfsdata.go file that statically implements the given virtual filesystem.

vfsgen is simple and minimalistic. You provide an input filesystem, and it generates an output .go file.

Features:

-	Efficient generated code without unneccessary overhead.

-	Uses gzip compression internally (selectively, only for files that compress well).

-	Enables direct access to internal gzip compressed bytes via an optional interface.

-	Outputs gofmt-compatible .go code.
*/
package vfsgen
