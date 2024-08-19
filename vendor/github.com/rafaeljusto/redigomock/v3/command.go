// Copyright 2014 Rafael Dantas Justo. All rights reserved.
// Use of this source code is governed by a GPL
// license that can be found in the LICENSE file.

package redigomock

import (
	"fmt"
	"reflect"
	"sync"
)

// response struct that represents single response from `Do` call.
type response struct {
	response interface{} // Response to send back when this command/arguments are called
	err      error       // Error to send back when this command/arguments are called
	panicVal interface{} // Panic to throw when this command/arguments are called
}

// ResponseHandler dynamic handles the response for the provided arguments.
type ResponseHandler func(args []interface{}) (interface{}, error)

// Cmd stores the registered information about a command to return it later
// when request by a command execution
type Cmd struct {
	// name and args must not be mutated after creation.
	name      string        // Name of the command
	args      []interface{} // Arguments of the command
	responses []response    // Slice of returned responses
	called    bool          // State for this command called or not
	mu        sync.Mutex    // hold while accessing responses and called
}

// cmdHash stores a unique identifier of the command
type cmdHash string

// equal verify if a command/arguments is related to a registered command
func equal(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.name || len(args) != len(cmd.args) {
		return false
	}

	for pos := range cmd.args {
		if implementsFuzzy(cmd.args[pos]) && implementsFuzzy(args[pos]) {
			if reflect.TypeOf(cmd.args[pos]) != reflect.TypeOf(args[pos]) {
				return false
			}
		} else if implementsFuzzy(cmd.args[pos]) || implementsFuzzy(args[pos]) {
			return false
		} else {
			if !reflect.DeepEqual(cmd.args[pos], args[pos]) {
				return false
			}
		}
	}
	return true
}

// match check if provided arguments can be matched with any registered
// commands
func match(commandName string, args []interface{}, cmd *Cmd) bool {
	if commandName != cmd.name || len(args) != len(cmd.args) {
		return false
	}

	for pos := range cmd.args {
		if implementsFuzzy(cmd.args[pos]) {
			if !cmd.args[pos].(FuzzyMatcher).Match(args[pos]) {
				return false
			}
		} else if !reflect.DeepEqual(cmd.args[pos], args[pos]) {
			return false
		}
	}
	return true
}

// Expect sets a response for this command. Every time a Do or Receive method
// is executed for a registered command this response or error will be
// returned. Expect call returns a pointer to Cmd struct, so you can chain
// Expect calls. Chained responses will be returned on subsequent calls
// matching this commands arguments in FIFO order
func (c *Cmd) Expect(resp interface{}) *Cmd {
	c.expect(response{resp, nil, nil})
	return c
}

// expect appends the argument to the response-slice, holding the lock
func (c *Cmd) expect(resp response) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responses = append(c.responses, resp)
}

// ExpectMap works in the same way of the Expect command, but has a key/value
// input to make it easier to build test environments
func (c *Cmd) ExpectMap(resp map[string]string) *Cmd {
	var values []interface{}
	for key, value := range resp {
		values = append(values, []byte(key))
		values = append(values, []byte(value))
	}
	c.expect(response{values, nil, nil})
	return c
}

// ExpectError allows you to force an error when executing a
// command/arguments
func (c *Cmd) ExpectError(err error) *Cmd {
	c.expect(response{nil, err, nil})
	return c
}

// ExpectPanic allows you to force a panic when executing a
// command/arguments
func (c *Cmd) ExpectPanic(msg interface{}) *Cmd {
	c.expect(response{nil, nil, msg})
	return c
}

// ExpectSlice makes it easier to expect slice value
// e.g - HMGET command
func (c *Cmd) ExpectSlice(resp ...interface{}) *Cmd {
	ifaces := []interface{}{}
	ifaces = append(ifaces, resp...)
	c.expect(response{ifaces, nil, nil})
	return c
}

// ExpectStringSlice makes it easier to expect a slice of strings, plays nicely
// with redigo.Strings
func (c *Cmd) ExpectStringSlice(resp ...string) *Cmd {
	ifaces := []interface{}{}
	for _, r := range resp {
		ifaces = append(ifaces, []byte(r))
	}
	c.expect(response{ifaces, nil, nil})
	return c
}

// Handle registers a function to handle the incoming arguments, generating an
// on-the-fly response.
func (c *Cmd) Handle(fn ResponseHandler) *Cmd {
	c.expect(response{fn, nil, nil})
	return c
}

// hash generates a unique identifier for the command
func (c *Cmd) hash() cmdHash {
	output := c.name
	for _, arg := range c.args {
		output += fmt.Sprintf("%v", arg)
	}
	return cmdHash(output)
}

// Called returns true if the command-mock was ever called.
func (c *Cmd) Called() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.called
}

// getResponse marks the command as used, and gets the next response to return.
func (c *Cmd) getResponse() *response {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.called = true
	if len(c.responses) == 0 {
		return nil
	}

	resp := c.responses[0]
	if len(c.responses) > 1 {
		c.responses = c.responses[1:]
	}
	return &resp
}
