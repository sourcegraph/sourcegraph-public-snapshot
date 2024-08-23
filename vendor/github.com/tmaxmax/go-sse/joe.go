package sse

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
)

// A ReplayProvider is a type that can replay older published events to new subscribers.
// Replay providers use event IDs, the topics the events were published to and optionally
// the events' expiration times or any other criteria to determine which are valid for replay.
//
// While providers can require events to have IDs beforehand, they can also set the IDs themselves,
// automatically - it's up to the implementation. Providers should ignore events without IDs,
// if they require IDs to be set.
//
// Replay providers are not required to be thread-safe - server providers are required to ensure only
// one operation is executed on the replay provider at any given time. Server providers may not execute
// replay operation concurrently with other operations, so make sure any action on the replay provider
// blocks for as little as possible. If a replay provider is thread-safe, some operations may be
// run in a separate goroutine - see the interface's method documentation.
//
// Executing actions that require waiting for a long time on I/O, such as HTTP requests or database
// calls must be handled with great care, so the server provider is not blocked. Reducing them to
// the minimum by using techniques such as caching or by executing them in separate goroutines is
// recommended, as long as the implementation fulfills the requirements.
//
// If not specified otherwise, the errors returned are implementation-specific.
type ReplayProvider interface {
	// Put adds a new event to the replay buffer. The Message that is returned may not have the
	// same address, if the replay provider automatically sets IDs.
	//
	// Put panics if the message couldn't be queued – if no topics are provided, or
	// a message without an ID is put into a ReplayProvider which does not
	// automatically set IDs.
	//
	// The Put operation may be executed by the replay provider in another goroutine only if
	// it can ensure that any Replay operation called after the Put goroutine is started
	// can replay the new received message. This also requires the replay provider implementation
	// to be thread-safe.
	//
	// Replay providers are not required to guarantee that after Put returns the new events
	// can be replayed. If an error occurs internally when putting the new message
	// and retrying the operation would block for too long, it can be aborted.
	// The errors aren't returned as the server providers won't be able to handle them in a useful manner.
	Put(message *Message, topics []string) *Message
	// Replay sends to a new subscriber all the valid events received by the provider
	// since the event with the listener's ID. If the ID the listener provides
	// is invalid, the provider should not replay any events.
	//
	// Replay operations must be executed in the same goroutine as the one it is called in.
	// Other goroutines may be launched from inside the Replay method, but the events must
	// be sent to the listener in the same goroutine that Replay is called in.
	//
	// If an error is returned, then at least some messages weren't successfully replayed.
	// The error is nil if there were no messages to replay for the particular subscription
	// or if all messages were replayed successfully.
	Replay(subscription Subscription) error
}

type (
	subscriber   chan<- error
	subscription struct {
		done subscriber
		Subscription
	}

	messageWithTopics struct {
		message *Message
		topics  []string
	}
)

// Joe is a basic server provider that synchronously executes operations by queueing them in channels.
// Events are also sent synchronously to subscribers, so if a subscriber's callback blocks, the others
// have to wait.
//
// Joe optionally supports event replaying with the help of a replay provider.
//
// If the replay provider panics, the subscription for which it panicked is considered failed
// and an error is returned, and thereafter the replay provider is not used anymore – no replays
// will be attempted for future subscriptions.
// If due to some other unexpected scenario something panics internally, Joe will remove all subscribers
// and close itself, so subscribers don't end up blocked.
//
// He serves simple use-cases well, as he's light on resources, and does not require any external
// services. Also, he is the default provider for Servers.
type Joe struct {
	message        chan messageWithTopics
	subscription   chan subscription
	unsubscription chan subscriber
	done           chan struct{}
	closed         chan struct{}
	subscribers    map[subscriber]Subscription

	// An optional replay provider that Joe uses to resend older messages to new subscribers.
	ReplayProvider ReplayProvider

	initDone sync.Once
}

// Subscribe tells Joe to send new messages to this subscriber. The subscription
// is automatically removed when the context is done, a callback error occurs
// or Joe is stopped.
func (j *Joe) Subscribe(ctx context.Context, sub Subscription) error {
	j.init()

	done := make(chan error, 1)

	select {
	case <-j.done:
		return ErrProviderClosed
	case j.subscription <- subscription{done: done, Subscription: sub}:
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}

	select {
	case err := <-done:
		return err
	case j.unsubscription <- done:
		return nil
	}
}

// Publish tells Joe to send the given message to the subscribers.
// When a message is published to multiple topics, Joe makes sure to
// not send the Message multiple times to clients that are subscribed
// to more than one topic that receive the given Message. Every client
// receives each unique message once, regardless of how many topics it
// is subscribed to or to how many topics the message is published.
func (j *Joe) Publish(msg *Message, topics []string) error {
	if len(topics) == 0 {
		return ErrNoTopic
	}

	j.init()

	// Waiting on done ensures Publish doesn't block the caller goroutine
	// when Joe is stopped and implements the required Provider behavior.
	select {
	case j.message <- messageWithTopics{message: msg, topics: topics}:
		return nil
	case <-j.done:
		return ErrProviderClosed
	}
}

// Stop signals Joe to close all subscribers and stop receiving messages.
// It returns when all the subscribers are closed.
//
// Further calls to Stop will return ErrProviderClosed.
func (j *Joe) Shutdown(ctx context.Context) (err error) {
	j.init()

	defer func() {
		if r := recover(); r != nil {
			err = ErrProviderClosed
		}
	}()

	close(j.done)

	select {
	case <-j.closed:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return
}

func (j *Joe) removeSubscriber(sub subscriber) {
	delete(j.subscribers, sub)
	close(sub)
}

func (j *Joe) start(replay ReplayProvider) {
	defer close(j.closed)
	// defer closing all subscribers instead of closing them when done is closed
	// so in case of a panic subscribers won't block the request goroutines forever.
	defer j.closeSubscribers()

	canReplay := true

	for {
		select {
		case msg := <-j.message:
			toDispatch := msg.message
			if canReplay {
				toDispatch = j.tryPut(msg, replay, &canReplay)
			}

			for done, sub := range j.subscribers {
				if topicsIntersect(sub.Topics, msg.topics) {
					err := sub.Client.Send(toDispatch)
					if err == nil {
						err = sub.Client.Flush()
					}

					if err != nil {
						done <- err
						j.removeSubscriber(done)
					}
				}
			}
		case sub := <-j.subscription:
			var err error
			if canReplay {
				err = j.tryReplay(sub.Subscription, replay, &canReplay)
			}

			if err != nil && err != errReplayPanicked { //nolint:errorlint // This is our error.
				sub.done <- err
				close(sub.done)
			} else {
				j.subscribers[sub.done] = sub.Subscription
			}
		case sub := <-j.unsubscription:
			j.removeSubscriber(sub)
		case <-j.done:
			return
		}
	}
}

func (j *Joe) closeSubscribers() {
	for done := range j.subscribers {
		j.removeSubscriber(done)
	}
}

var errReplayPanicked = errors.New("replay failed unexpectedly")

func (*Joe) tryReplay(sub Subscription, replay ReplayProvider, canReplay *bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			*canReplay = false
			err = errReplayPanicked
			log.Printf("panic: %v\n%s", r, debug.Stack())
		}
	}()

	err = replay.Replay(sub)

	return
}

func (*Joe) tryPut(msg messageWithTopics, replay ReplayProvider, canReplay *bool) *Message {
	defer func() {
		if r := recover(); r != nil {
			*canReplay = false
			log.Printf("panic: %v\n%s", r, debug.Stack())
		}
	}()

	return replay.Put(msg.message, msg.topics)
}

func (j *Joe) init() {
	j.initDone.Do(func() {
		j.message = make(chan messageWithTopics)
		j.subscription = make(chan subscription)
		j.unsubscription = make(chan subscriber)
		j.done = make(chan struct{})
		j.closed = make(chan struct{})
		j.subscribers = map[subscriber]Subscription{}

		replay := j.ReplayProvider
		if replay == nil {
			replay = noopReplayProvider{}
		}
		go j.start(replay)
	})
}
