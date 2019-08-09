// Package pubsubutil is a lightweight wrapper around pubsub which initializes
// pubsub based on environment variables.
package pubsubutil

import (
	"context"
	"errors"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// PubSubProjectID is used to create a new pubsub client.
var PubSubProjectID = env.Get("PUBSUB_PROJECT_ID", "", "Pub/sub project ID is the id of the pubsub project.")

// PubSubTopicID is the topic ID of the topic that forwards messages from publishers to subscribers.
var PubSubTopicID = env.Get("PUBSUB_TOPIC_ID", "", "Pub/sub topic ID is the pub/sub topic id where messages are published.")

var client *pubsub.Client

// Enabled returns true if pubsub has been configured to run.
func Enabled() bool {
	return PubSubProjectID != ""
}

func init() {
	if !Enabled() {
		return
	}
	ctx := context.Background()
	pClient, err := pubsub.NewClient(ctx, PubSubProjectID)
	if err != nil {
		log15.Error("failed to create pubsub client.", "error", err)
		return
	}
	client = pClient
}

// Publish publishes msg to the topic asynchronously. Messages are batched and sent according to the topic's PublishSettings.
func Publish(topic string, msg string) error {
	if client == nil {
		return errors.New("failed to publish message: pubsub client is not initialized")
	}

	ctx := context.Background()
	t := client.Topic(topic)
	defer t.Stop()
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	_, err := result.Get(ctx)
	if err != nil {
		return err
	}
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_862(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
