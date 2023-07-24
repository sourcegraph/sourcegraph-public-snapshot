// Package chunk provides a utility for sending sets of protobuf messages in
// groups of smaller chunks. This is useful for gRPC, which has limitations around the maximum
// size of a message that you can send.
//
// This code is adapted from the gitaly project, which is licensed
// under the MIT license. A copy of that license text can be found at
// https://mit-license.org/.
//
// The code this file was based off can be found here: https://gitlab.com/gitlab-org/gitaly/-/blob/v16.2.0/internal/helper/chunk/chunker.go
package chunk

import (
	"google.golang.org/protobuf/proto"
)

// Message is a protobuf message.
type Message interface {
	proto.Message
}

// Sender encapsulates a gRPC response stream and the current chunk
// that's being built.
//
// Reset, Append, [Append...], Send, Reset, Append, [Append...], Send, ...
type Sender[T Message] interface {
	// Reset should create a fresh response message.
	Reset()
	// Append should append the given items to the slice in the current response message
	Append(...T)
	// Send should send the current response message
	Send() error
}

// New returns a new Chunker.
func New[T Message](s Sender[T]) *Chunker[T] { return &Chunker[T]{s: s} }

// Chunker lets you spread items you want to send over multiple chunks.
// This type is not thread-safe.
type Chunker[T Message] struct {
	s    Sender[T]
	size int
}

// maxMessageSize is the maximum size per protobuf message
const maxMessageSize = 1 * 1024 * 1024

// Send will append the provided items to the current chunk, and send the chunk if it is full.
//
// Callers should ensure that they call Flush() after the last call to Send().
func (c *Chunker[T]) Send(items ...T) error {
	for _, it := range items {
		if c.size == 0 {
			c.s.Reset()
		}

		itSize := proto.Size(it)

		if itSize+c.size >= maxMessageSize {
			if err := c.sendResponseMsg(); err != nil {
				return err
			}
			c.s.Reset()
		}

		c.s.Append(it)
		c.size += itSize
	}

	return nil
}

func (c *Chunker[T]) sendResponseMsg() error {
	c.size = 0
	return c.s.Send()
}

// Flush sends remaining items in the current chunk, if any.
func (c *Chunker[T]) Flush() error {
	if c.size == 0 {
		return nil
	}

	return c.sendResponseMsg()
}
