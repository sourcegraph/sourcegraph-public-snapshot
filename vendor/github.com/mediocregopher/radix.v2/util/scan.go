package util

import (
	"errors"
	"strings"

	"github.com/mediocregopher/radix.v2/cluster"
)

func scanSingle(r Cmder, ch chan string, cmd, key, pattern string) error {
	defer close(ch)
	cmd = strings.ToUpper(cmd)

	var keys []string
	cursor := "0"
	for {
		args := make([]interface{}, 0, 4)
		if cmd != "SCAN" {
			args = append(args, key)
		}
		args = append(args, cursor, "MATCH", pattern)

		parts, err := r.Cmd(cmd, args...).Array()
		if err != nil {
			return err
		}

		if len(parts) < 2 {
			return errors.New("not enough parts returned")
		}

		if cursor, err = parts[0].Str(); err != nil {
			return err
		}

		if keys, err = parts[1].List(); err != nil {
			return err
		}

		for i := range keys {
			ch <- keys[i]
		}

		if cursor == "0" {
			return nil
		}
	}
}

// scanCluster is like Scan except it operates over a whole cluster. Unlike Scan
// it only works with SCAN and as such only takes in a pattern string.
func scanCluster(c *cluster.Cluster, ch chan string, pattern string) error {
	defer close(ch)
	clients, err := c.GetEvery()
	if err != nil {
		return err
	}
	for _, client := range clients {
		defer c.Put(client)
	}

	for _, client := range clients {
		cch := make(chan string)
		var err error
		go func() {
			err = scanSingle(client, cch, "SCAN", "", pattern)
		}()
		for key := range cch {
			ch <- key
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Scan is a helper function for performing any of the redis *SCAN functions. It
// takes in a channel which keys returned by redis will be written to, and
// returns an error should any occur. The input channel will always be closed
// when Scan returns, and *must* be read until it is closed.
//
// The key argument is only needed if cmd isn't SCAN
//
// Example SCAN command
//
//	ch := make(chan string)
//	var err error
//	go func() {
//		err = util.Scan(r, ch, "SCAN", "", "*")
//	}()
//	for key := range ch {
//		// do something with key
//	}
//	if err != nil {
//		// handle error
//	}
//
// Example HSCAN command
//
//	ch := make(chan string)
//	var err error
//	go func() {
//		err = util.Scan(r, ch, "HSCAN", "somekey", "*")
//	}()
//	for key := range ch {
//		// do something with key
//	}
//	if err != nil {
//		// handle error
//	}
//
func Scan(r Cmder, ch chan string, cmd, key, pattern string) error {
	if rr, ok := r.(*cluster.Cluster); ok && strings.ToUpper(cmd) == "SCAN" {
		return scanCluster(rr, ch, pattern)
	}
	var cmdErr error
	err := withClientForKey(r, key, func(c Cmder) {
		cmdErr = scanSingle(r, ch, cmd, key, pattern)
	})
	if err != nil {
		return err
	}
	return cmdErr
}
