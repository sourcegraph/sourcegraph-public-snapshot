// Package objectstore implements support for tracking a mapping of object
// references to and from their instance ID. It tracks objects by proxy of their
// memory address (i.e: pointer value), in order to avoid the pitfalls of go's
// standard object equality mechanism (which is also reflect.Value's equality
// mechanism) causing distinct instances appearing to be equal (including when
// used as keys to a map).
package objectstore
