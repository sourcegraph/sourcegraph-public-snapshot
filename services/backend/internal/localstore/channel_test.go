// +build pgsqltest

package localstore

import (
	"reflect"
	"sync"
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

func TestChannel_multi(t *testing.T) {
	ctx, _, done := testContext()
	defer done()

	s := &channel{}

	// Open listen channels on 2 channels.
	l0, unlisten0, err := s.Listen(ctx, "c0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if unlisten0 != nil {
			unlisten0()
		}
	}()
	l1, unlisten1, err := s.Listen(ctx, "c1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if unlisten1 != nil {
			unlisten1()
		}
	}()

	// Separate receivers as there is no guarantee in the order;
	// DeepEqual cares about order, and two recievers can guarantee order.
	var recv0 []store.ChannelNotification
	var recv1 []store.ChannelNotification
	var mu = &sync.Mutex{}

	go func() {
		for n := range l0 {
			mu.Lock()
			recv0 = append(recv0, n)
			mu.Unlock()
		}
	}()

	go func() {
		for n := range l1 {
			mu.Lock()
			recv1 = append(recv1, n)
			mu.Unlock()
		}
	}()

	want := []store.ChannelNotification{
		{Channel: "c0", Payload: "p0"},
		{Channel: "c1", Payload: "p1"},
		{Channel: "c0", Payload: "p2"},
		{Channel: "c1", Payload: "p3"},
		{Channel: "c0", Payload: "p4"},
	}

	for _, n := range want {
		if err := s.Notify(ctx, n.Channel, n.Payload); err != nil {
			t.Fatal(err)
		}

		// Send some dummy notifications as well.
		if err := s.Notify(ctx, n.Channel+"ignore", n.Payload); err != nil {
			t.Fatal(err)
		}
	}

	want0 := []store.ChannelNotification{
		{Channel: "c0", Payload: "p0"},
		{Channel: "c0", Payload: "p2"},
		{Channel: "c0", Payload: "p4"},
	}

	want1 := []store.ChannelNotification{
		{Channel: "c1", Payload: "p1"},
		{Channel: "c1", Payload: "p3"},
	}

	// Sleep to make sure all values have been received
	time.Sleep(time.Second * 5)

	// Ensure that after unlistening, the channel doesn't receive any
	// more notifs.
	unlisten0()
	unlisten0 = nil
	unlisten1()
	unlisten1 = nil

	// Ensure both channels are closed.
	select {
	case _, ok := <-l0:
		if ok {
			t.Errorf("l0 not closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out")
	}
	select {
	case _, ok := <-l1:
		if ok {
			t.Errorf("l1 not closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out")
	}

	mu.Lock()
	if !reflect.DeepEqual(recv0, want0) {
		t.Errorf("got %+v, want %+v", recv0, want0)
	}
	mu.Unlock()
	mu.Lock()
	if !reflect.DeepEqual(recv1, want1) {
		t.Errorf("got %+v, want %+v", recv1, want1)
	}
	mu.Unlock()
}
