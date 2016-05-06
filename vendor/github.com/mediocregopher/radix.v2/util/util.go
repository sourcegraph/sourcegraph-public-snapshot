// Package util is a collection of helper functions for interacting with various
// parts of the radix.v2 package
package util

import (
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

// Cmder is an interface which can be used to interchangeably work with either
// redis.Client (the basic, single connection redis client), pool.Pool, or
// cluster.Cluster. All three implement a Cmd method (although, as is the case
// with Cluster, sometimes with different limitations), and therefore all three
// are Cmders
type Cmder interface {
	Cmd(cmd string, args ...interface{}) *redis.Resp
}

// withClientForKey is useful for retrieving a single client which can handle
// the given key and perform one or more requests on them, especially when the
// passed in Cmder is actually a Cluster or Pool.
//
// The function given takes a Cmder and not a Client because the passed in Cmder
// may not be one implemented in radix.v2, and in that case may not actually
// have a way of mapping to a Client. In that case it is simply passed directly
// through to fn.
func withClientForKey(c Cmder, key string, fn func(c Cmder)) error {
	var singleC Cmder

	switch cc := c.(type) {
	case *cluster.Cluster:
		client, err := cc.GetForKey(key)
		if err != nil {
			return err
		}
		defer cc.Put(client)
		singleC = client

	case *pool.Pool:
		client, err := cc.Get()
		if err != nil {
			return err
		}
		defer cc.Put(client)
		singleC = client

	default:
		singleC = cc
	}

	fn(singleC)
	return nil
}
