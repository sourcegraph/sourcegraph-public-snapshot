package stackimpact

import (
	"sync"
	"time"
)

type Message struct {
	topic   string
	content map[string]interface{}
	addedAt int64
}

type MessageQueue struct {
	agent               *Agent
	queue               []Message
	queueLock           *sync.Mutex
	lastUploadTimestamp int64
	backoffSeconds      int
}

func newMessageQueue(agent *Agent) *MessageQueue {
	mq := &MessageQueue{
		agent:               agent,
		queue:               make([]Message, 0),
		queueLock:           &sync.Mutex{},
		lastUploadTimestamp: 0,
		backoffSeconds:      0,
	}

	return mq
}

func (mq *MessageQueue) start() {
	flushTicker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			ph := mq.agent.panicHandler()
			defer ph()

			select {
			case <-flushTicker.C:
				if len(mq.queue) > 0 && (mq.lastUploadTimestamp+int64(mq.backoffSeconds) < time.Now().Unix()) {
					mq.expire()
					mq.flush()
				}
			}
		}
	}()
}

func (mq *MessageQueue) expire() {
	now := time.Now().Unix()

	mq.queueLock.Lock()
	for i := len(mq.queue) - 1; i >= 0; i-- {
		if mq.queue[i].addedAt < now-10*60 {
			mq.queue = mq.queue[i+1:]
			break
		}
	}
	mq.queueLock.Unlock()
}

func (mq *MessageQueue) flush() {
	mq.queueLock.Lock()
	outgoing := mq.queue
	mq.queue = mq.queue[:0]
	mq.queueLock.Unlock()

	messages := make([]interface{}, 0)
	for _, m := range outgoing {
		message := map[string]interface{}{
			"topic":   m.topic,
			"content": m.content,
		}

		messages = append(messages, message)
	}

	payload := map[string]interface{}{
		"messages": messages,
	}

	mq.lastUploadTimestamp = time.Now().Unix()

	if _, err := mq.agent.apiRequest.post("upload", payload); err == nil {
		// reset backoff
		mq.backoffSeconds = 0
	} else {
		// prepend outgoing messages back to the queue
		mq.queueLock.Lock()
		mq.queue = append(outgoing, mq.queue...)
		mq.queueLock.Unlock()

		// increase backoff up to 1 minute
		mq.agent.log("Error uploading messages to dashboard, backing off next upload")
		if mq.backoffSeconds == 0 {
			mq.backoffSeconds = 10
		} else if mq.backoffSeconds*2 < 60 {
			mq.backoffSeconds *= 2
		}

		mq.agent.error(err)
	}
}

func (mq *MessageQueue) addMessage(topic string, message map[string]interface{}) {
	m := Message{
		topic:   topic,
		content: message,
		addedAt: time.Now().Unix(),
	}

	mq.queueLock.Lock()
	mq.queue = append(mq.queue, m)
	mq.queueLock.Unlock()

	mq.agent.log("Added message to the queue for topic: %v", topic)
	mq.agent.log("%v", message)
}
