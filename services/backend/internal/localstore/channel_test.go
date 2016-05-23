// +build pgsqltest

package localstore

import (
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

// TestChannel_Notify_noListeners tests the behavior of Channel.Notify when
// there are no listeners.
func TestChannel_Notify_noListeners(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &channel{}

	if err := s.Notify(ctx, "c", "p"); err != nil {
		t.Fatal(err)
	}
	if err := s.Notify(ctx, "c2", "p"); err != nil {
		t.Fatal(err)
	}
}

// TestChannel_single tests the behavior of the Channel store when
// there is a single channel.
func TestChannel_single(t *testing.T) {
	t.Parallel()

	ctx, _, done := testContext()
	defer done()

	s := &channel{}

	l, unlisten, err := s.Listen(ctx, "c")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if unlisten != nil {
			unlisten()
		}
	}()

	// Send a notif and check that we receive it.
	if err := s.Notify(ctx, "c", "p"); err != nil {
		t.Fatal(err)
	}
	select {
	case n := <-l:
		want := store.ChannelNotification{Channel: "c", Payload: "p"}
		if !reflect.DeepEqual(n, want) {
			t.Errorf("got %+v, want %+v", n, want)
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}

	// Ensure that after unlistening, the channel doesn't receive any
	// more notifs.
	unlisten()
	unlisten = nil
	if err := s.Notify(ctx, "c", "p2"); err != nil {
		t.Fatal(err)
	}
	select {
	case n, ok := <-l:
		if ok {
			t.Errorf("got %+v (and channel not closed), want nothing received and channel closed", n)
		}
	case <-time.After(time.Millisecond * 200):
		t.Error("timed out")
	}
}
