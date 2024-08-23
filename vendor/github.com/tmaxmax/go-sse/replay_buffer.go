package sse

import (
	"errors"
	"strconv"
	"strings"
)

// A buffer is the underlying storage for a provider. Its methods are used by the provider to implement
// the Provider interface.
type buffer interface {
	queue(message *Message, topics []string) *Message
	dequeue()
	front() *messageWithTopics
	len() int
	cap() int
	slice(EventID) []messageWithTopics
}

type bufferBase struct {
	buf []messageWithTopics
}

func (b *bufferBase) len() int {
	return len(b.buf)
}

func (b *bufferBase) cap() int {
	return cap(b.buf)
}

func (b *bufferBase) front() *messageWithTopics {
	if b.len() == 0 {
		return nil
	}
	return &b.buf[0]
}

func (b *bufferBase) queue(message *Message, topics []string) *Message {
	if len(topics) == 0 {
		panic(errors.New("go-sse: no topics provided for Message.\n" + formatMessagePanicString(message)))
	}

	b.buf = append(b.buf, messageWithTopics{message: message, topics: topics})

	return message
}

func (b *bufferBase) dequeue() {
	// It may seem at first glance that the backing array would grow indefinitely,
	// but factor in that when the slice is reallocated all the dequeued elements
	// from the beginning become reclaimable.
	b.buf = b.buf[1:]
}

type bufferNoID struct {
	lastRemovedID EventID
	bufferBase
}

func (b *bufferNoID) queue(message *Message, topics []string) *Message {
	if !message.ID.IsSet() {
		// We could maybe return this as an error and change the ReplayProvider
		// interface to return the error. The issue with that is the following:
		// even if we return this message as an error, providers can't handle it
		// in any meaningful manner – for example, Joe has no way to report
		// a replay.Put error, as that's not run on the main goroutine.
		// A panic seems fitting, as putting a message without an ID when using
		// a provider that doesn't add IDs is breaking the API contract – that is,
		// the provider expects a message with an ID. It seems to be an irrecoverable
		// error which should be caught in development.
		panicString := "go-sse: a Message without an ID was given to a provider that doesn't set IDs automatically.\n" + formatMessagePanicString(message)

		panic(errors.New(panicString))
	}

	return b.bufferBase.queue(message, topics)
}

func (b *bufferNoID) dequeue() {
	b.lastRemovedID = b.buf[0].message.ID
	b.bufferBase.dequeue()
}

func (b *bufferNoID) slice(atID EventID) []messageWithTopics {
	if !atID.IsSet() {
		return nil
	}
	if atID == b.lastRemovedID {
		return b.buf
	}
	index := -1
	for i := range b.buf {
		if atID == b.buf[i].message.ID {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}

	return b.buf[index+1:]
}

type bufferAutoID struct {
	bufferBase
	firstID    int64
	upcomingID int64
}

const autoIDBase = 10

func (b *bufferAutoID) queue(message *Message, topics []string) *Message {
	message = message.Clone()
	message.ID = ID(strconv.FormatInt(b.upcomingID, autoIDBase))
	b.upcomingID++

	return b.bufferBase.queue(message, topics)
}

func (b *bufferAutoID) dequeue() {
	b.firstID++
	b.bufferBase.dequeue()
}

func (b *bufferAutoID) slice(atID EventID) []messageWithTopics {
	id, err := strconv.ParseInt(atID.String(), autoIDBase, 64)
	if err != nil {
		return nil
	}
	index := id - b.firstID
	if index < -1 || index >= int64(len(b.buf)) {
		return nil
	}
	return b.buf[index+1:]
}

func getBuffer(autoIDs bool, capacity int) buffer {
	base := bufferBase{buf: make([]messageWithTopics, 0, capacity)}
	if autoIDs {
		return &bufferAutoID{bufferBase: base}
	}
	return &bufferNoID{bufferBase: base}
}

func formatMessagePanicString(m *Message) string {
	ret := "The message is the following:\n"
	for _, line := range strings.SplitAfter(m.String(), "\n") {
		if strings.TrimSpace(line) != "" {
			ret += "│ " + line
		}
	}
	return ret + "└─■"
}
