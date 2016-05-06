// Package pubsub provides a wrapper around a normal redis client which makes
// interacting with publish/subscribe commands much easier
package pubsub

import (
	"container/list"
	"errors"
	"fmt"

	"github.com/mediocregopher/radix.v2/redis"
)

// SubRespType describes the type of the response  being returned from one of
// the methods in this package
type SubRespType uint8

// The different kinds of SubRespTypes
const (
	Error SubRespType = iota
	Subscribe
	Unsubscribe
	Message
)

// SubClient wraps a Redis client to provide convenience methods for Pub/Sub
// functionality.
type SubClient struct {
	Client   *redis.Client
	messages *list.List
}

// SubResp wraps a Redis resp and provides convenient access to Pub/Sub info.
type SubResp struct {
	*redis.Resp // Original Redis resp

	Type     SubRespType
	Channel  string // Channel resp is on (Message)
	Pattern  string // Pattern which was matched for publishes captured by a PSubscribe
	SubCount int    // Count of subs active after this action (Subscribe or Unsubscribe)
	Message  string // Publish message (Message)
	Err      error  // SubResp error (Error)
}

// Timeout determines if this SubResp is an error type
// due to a timeout reading from the network
func (r *SubResp) Timeout() bool {
	return redis.IsTimeout(r.Resp)
}

// NewSubClient takes an existing, connected redis.Client and wraps it in a
// SubClient, returning that. The passed in redis.Client should not be used as
// long as the SubClient is also being used
func NewSubClient(client *redis.Client) *SubClient {
	return &SubClient{client, &list.List{}}
}

// Subscribe makes a Redis "SUBSCRIBE" command on the provided channels
func (c *SubClient) Subscribe(channels ...interface{}) *SubResp {
	return c.filterMessages("SUBSCRIBE", channels...)
}

// PSubscribe makes a Redis "PSUBSCRIBE" command on the provided patterns
func (c *SubClient) PSubscribe(patterns ...interface{}) *SubResp {
	return c.filterMessages("PSUBSCRIBE", patterns...)
}

// Unsubscribe makes a Redis "UNSUBSCRIBE" command on the provided channels
func (c *SubClient) Unsubscribe(channels ...interface{}) *SubResp {
	return c.filterMessages("UNSUBSCRIBE", channels...)
}

// PUnsubscribe makes a Redis "PUNSUBSCRIBE" command on the provided patterns
func (c *SubClient) PUnsubscribe(patterns ...interface{}) *SubResp {
	return c.filterMessages("PUNSUBSCRIBE", patterns...)
}

// Receive returns the next publish resp on the Redis client. It is possible
// Receive will timeout, and the *SubResp will be an Error. You can use the
// Timeout() method on SubResp to easily determine if that is the case. If this
// is the case you can call Receive again to continue listening for publishes
func (c *SubClient) Receive() *SubResp {
	return c.receive(false)
}

func (c *SubClient) receive(skipBuffer bool) *SubResp {
	if c.messages.Len() > 0 && !skipBuffer {
		v := c.messages.Remove(c.messages.Front())
		return v.(*SubResp)
	}
	r := c.Client.ReadResp()
	return c.parseResp(r)
}

func (c *SubClient) filterMessages(cmd string, names ...interface{}) *SubResp {
	r := c.Client.Cmd(cmd, names...)
	var sr *SubResp
	for i := 0; i < len(names); i++ {
		// If nil we know this is the first loop
		if sr == nil {
			sr = c.parseResp(r)
		} else {
			sr = c.receive(true)
		}
		if sr.Type == Message {
			c.messages.PushBack(sr)
			i--
		}
	}
	return sr
}

func (c *SubClient) parseResp(resp *redis.Resp) *SubResp {
	sr := &SubResp{Resp: resp}
	var elems []*redis.Resp

	switch {
	case resp.IsType(redis.Array):
		elems, _ = resp.Array()
		if len(elems) < 3 {
			sr.Err = errors.New("resp is not formatted as a subscription resp")
			sr.Type = Error
			return sr
		}

	case resp.IsType(redis.Err):
		sr.Err = resp.Err
		sr.Type = Error
		return sr

	default:
		sr.Err = errors.New("resp is not formatted as a subscription resp")
		sr.Type = Error
		return sr
	}

	rtype, err := elems[0].Str()
	if err != nil {
		sr.Err = fmt.Errorf("resp type: %s", err)
		sr.Type = Error
		return sr
	}

	//first element
	switch rtype {
	case "subscribe", "psubscribe":
		sr.Type = Subscribe
		count, err := elems[2].Int()
		if err != nil {
			sr.Err = fmt.Errorf("subscribe count: %s", err)
			sr.Type = Error
		} else {
			sr.SubCount = int(count)
		}

	case "unsubscribe", "punsubscribe":
		sr.Type = Unsubscribe
		count, err := elems[2].Int()
		if err != nil {
			sr.Err = fmt.Errorf("unsubscribe count: %s", err)
			sr.Type = Error
		} else {
			sr.SubCount = int(count)
		}

	case "message", "pmessage":
		var chanI, msgI int

		if rtype == "message" {
			chanI, msgI = 1, 2
		} else { // "pmessage"
			chanI, msgI = 2, 3
			pattern, err := elems[1].Str()
			if err != nil {
				sr.Err = fmt.Errorf("message pattern: %s", err)
				sr.Type = Error
				return sr
			}
			sr.Pattern = pattern
		}

		sr.Type = Message
		channel, err := elems[chanI].Str()
		if err != nil {
			sr.Err = fmt.Errorf("message channel: %s", err)
			sr.Type = Error
			return sr
		}
		sr.Channel = channel
		msg, err := elems[msgI].Str()
		if err != nil {
			sr.Err = fmt.Errorf("message msg: %s", err)
			sr.Type = Error
		} else {
			sr.Message = msg
		}
	default:
		sr.Err = errors.New("suscription multiresp has invalid type: " + rtype)
		sr.Type = Error
	}
	return sr
}
