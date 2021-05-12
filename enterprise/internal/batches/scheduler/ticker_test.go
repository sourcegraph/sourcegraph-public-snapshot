package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTickerGoBrrr(t *testing.T) {
	// We'll run the tests in this file in parallel, since they need to perform
	// brief blocks, and there's no reason we should run them sequentially.
	t.Parallel()

	// We'll set up an unlimited schedule, and then use that to verify that
	// delays are appropriately handled and that stopping the ticker works as
	// expected.
	cfg, err := window.NewConfiguration(nil)
	if err != nil {
		t.Fatal(err)
	}
	ticker := newTicker(cfg.Schedule())

	// Take three as quickly as we can, with no delays going back.
	for i := 0; i < 3; i++ {
		c := <-ticker.C
		if c == nil {
			t.Errorf("unexpected nil channel")
		}
		c <- time.Duration(0)
	}

	// Now send back a 10 ms delay and ensure that it takes at least 10 ms to
	// get the following message.
	delay := 10 * time.Millisecond
	now := time.Now()
	c := <-ticker.C
	c <- delay

	c = <-ticker.C
	if have := time.Since(now); have < delay {
		t.Errorf("unexpectedly short delay between takes: have=%v want>=%v", have, delay)
	}
	c <- time.Duration(0)

	// Finally, let's stop the ticker and make sure that the channel is closed.
	ticker.stop()
	// Also read from the now-closed `done` to synchronize, since closing a
	// channel is non-blocking.
	<-ticker.done
	// Now make sure that the channel is closed.
	if c := <-ticker.C; c != nil {
		t.Errorf("unexpected non-nil channel: %v", c)
	}
}

func TestTickerRateLimited(t *testing.T) {
	t.Parallel()

	// We'll set up a 500/sec rate limit, and then ensure we take at least 2 ms
	// between ticks without any other delays.
	cfg, err := window.NewConfiguration(&[]*schema.BatchChangeRolloutWindow{
		{Rate: "500/sec"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ticker := newTicker(cfg.Schedule())

	// We'll take eleven ticks, which should take ten delay periods. Since the
	// delay between ticks is 2 ms, this should be about 20 ms. We'll be happy
	// with anything from 15 to 30, since we can't promise other things aren't
	// happening at the same time.
	now := time.Now()
	for i := 0; i < 11; i++ {
		c := <-ticker.C
		c <- time.Duration(0)
	}

	have := time.Since(now)
	if wantMin, wantMax := 15*time.Millisecond, 30*time.Millisecond; have < wantMin || have > wantMax {
		t.Errorf("unexpected delay across a group of ticks: have=%v want >=%v && <=%v", have, wantMin, wantMax)
	}

	// Finally, let's stop the ticker
	ticker.stop()
	// Also read from the now-closed `done` to synchronize, since closing a
	// channel is non-blocking.
	<-ticker.done
	// Now make sure that the channel is closed.
	if c := <-ticker.C; c != nil {
		t.Errorf("unexpected non-nil channel: %v", c)
	}
}

func TestTickerZero(t *testing.T) {
	t.Parallel()

	// Set up a zero rate limit.
	cfg, err := window.NewConfiguration(&[]*schema.BatchChangeRolloutWindow{
		{Rate: "0/sec"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ticker := newTicker(cfg.Schedule())

	// Wait for ticker.C, which should only ever return nil (since the channel
	// will be closed).
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if c := <-ticker.C; c != nil {
			t.Errorf("unexpected non-nil channel: %v", c)
		}
	}()

	// Wait 10 ms and then stop the ticker.
	time.Sleep(10 * time.Millisecond)
	ticker.stop()

	wg.Wait()
}
