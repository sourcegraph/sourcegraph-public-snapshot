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

// New returns a new Chunker that will use the given sendFunc to send chunks of messages.
func New[T proto.Message](sendFunc func([]T) error) *Chunker[T] {
	return &Chunker[T]{sendFunc: sendFunc}
}

// Chunker lets you spread items you want to send over multiple chunks.
// This type is not thread-safe.
type Chunker[T proto.Message] struct {
	sendFunc func([]T) error // sendFunc is the function that will be invoked when a chunk is ready to be sent.

	buffer    []T // buffer stores the items that will be sent when the sendFunc is invoked.
	sizeBytes int // sizeBytes is the size of the current chunk in bytes.
}

// maxMessageSize is the maximum size per protobuf message
const maxMessageSize = 1 * 1024 * 1024 // 1 MiB

// Send will append the provided items to the current chunk, and send the chunk if it is full.
//
// Callers should ensure that they call Flush() after the last call to Send().
func (c *Chunker[T]) Send(items ...T) error {
	for _, item := range items {
		if err := c.sendOne(item); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chunker[T]) sendOne(item T) error {
	itemSize := proto.Size(item)

	if itemSize+c.sizeBytes >= maxMessageSize {
		if err := c.sendResponseMsg(); err != nil {
			return err
		}
	}

	c.buffer = append(c.buffer, item)
	c.sizeBytes += itemSize

	return nil
}

func (c *Chunker[T]) sendResponseMsg() error {
	c.sizeBytes = 0

	err := c.sendFunc(c.buffer)
	if err != nil {
		return err
	}

	c.buffer = c.buffer[:0]
	return nil
}

// Flush sends remaining items in the current chunk, if any.
func (c *Chunker[T]) Flush() error {
	if len(c.buffer) == 0 {
		return nil
	}

	err := c.sendResponseMsg()
	if err != nil {
		return err
	}

	return nil
}

// SendAll is a convenience function that immediately sends all provided items in smaller chunks using the provided
// sendFunc.
//
// See the documentation for Chunker.Send() for more information.
func SendAll[T proto.Message](sendFunc func([]T) error, items ...T) error {
	c := New(sendFunc)

	err := c.Send(items...)
	if err != nil {
		return err
	}

	return c.Flush()
}
