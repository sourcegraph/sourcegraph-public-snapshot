package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTakerGoBrrr(t *testing.T) {
	// We'll run the tests in this file in parallel, since they need to perform
	// brief blocks, and there's no reason we should run them sequentially.
	t.Parallel()

	// We'll set up an unlimited schedule, and then use that to verify that
	// delays are appropriately handled and that stopping the taker works as
	// expected.
	cfg, err := window.NewConfiguration(nil)
	if err != nil {
		t.Fatal(err)
	}
	taker := newTaker(cfg.Schedule())

	// Take three as quickly as we can, with no delays going back.
	for i := 0; i < 3; i++ {
		c := <-taker.C
		if c == nil {
			t.Errorf("unexpected nil channel")
		}
		c <- time.Duration(0)
	}

	// Now send back a 10 ms delay and ensure that it takes at least 10 ms to get the following message.
	delay := 10 * time.Millisecond
	now := time.Now()
	c := <-taker.C
	c <- delay

	c = <-taker.C
	if have := time.Since(now); have < delay {
		t.Errorf("unexpectedly short delay between takes: have=%v want>=%v", have, delay)
	}
	c <- time.Duration(0)

	// Finally, let's stop the taker and make sure that the channel is closed.
	taker.stop()
	if c := <-taker.C; c != nil {
		t.Errorf("unexpected non-nil channel: %v", c)
	}
}

func TestTakerRateLimited(t *testing.T) {
	t.Parallel()

	// We'll set up a 100/sec rate limit, and then ensure we take at least 10 ms
	// to take two messages without any other delays.
	cfg, err := window.NewConfiguration(&[]*schema.BatchChangeRolloutWindow{
		{Rate: "100/sec"},
	})
	if err != nil {
		t.Fatal(err)
	}
	taker := newTaker(cfg.Schedule())

	// We'll take two messages, which should be at least 10 ms apart.
	now := time.Now()
	c := <-taker.C
	c <- time.Duration(0)

	c = <-taker.C
	if have, want := time.Since(now), 10*time.Millisecond; have < want {
		t.Errorf("unexpectedly short delay between takes: have=%v want>=%v", have, want)
	}
	c <- time.Duration(0)

	// Finally, let's stop the taker and make sure that the channel is closed.
	taker.stop()
	if c := <-taker.C; c != nil {
		t.Errorf("unexpected non-nil channel: %v", c)
	}
}

func TestTakerZero(t *testing.T) {
	t.Parallel()

	// Set up a zero rate limit.
	cfg, err := window.NewConfiguration(&[]*schema.BatchChangeRolloutWindow{
		{Rate: "0/sec"},
	})
	if err != nil {
		t.Fatal(err)
	}
	taker := newTaker(cfg.Schedule())

	// Wait for taker.C, which should only ever return nil (since the channel
	// will be closed).
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c := <-taker.C; c != nil {
			t.Errorf("unexpected non-nil channel: %v", c)
		}
	}()

	// Wait 10 ms and then stop the taker.
	time.Sleep(10 * time.Millisecond)
	taker.stop()

	wg.Wait()
}
