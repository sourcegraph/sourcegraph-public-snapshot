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

// New returns a new Chunker that will use the given sendFunc to send chunks of messages.
func New[T Message](sendFunc func([]T) error) *Chunker[T] {
	return &Chunker[T]{sendFunc: sendFunc}
}

// Chunker lets you spread items you want to send over multiple chunks.
// This type is not thread-safe.
type Chunker[T Message] struct {
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
	if c.sizeBytes == 0 {
		c.clearBuffer()
	}

	itemSize := proto.Size(item)

	if itemSize+c.sizeBytes >= maxMessageSize {
		if err := c.sendResponseMsg(); err != nil {
			return err
		}

		c.clearBuffer()
	}

	c.append(item)
	c.sizeBytes += itemSize

	return nil
}

func (c *Chunker[T]) append(items ...T) {
	c.buffer = append(c.buffer, items...)
}

func (c *Chunker[T]) clearBuffer() {
	c.buffer = c.buffer[:0]
}

func (c *Chunker[T]) sendResponseMsg() error {
	c.sizeBytes = 0
	return c.sendFunc(c.buffer)
}

// Flush sends remaining items in the current chunk, if any.
func (c *Chunker[T]) Flush() error {
	if c.sizeBytes == 0 {
		return nil
	}

	return c.sendResponseMsg()
}
