package ui

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
)

func TestIntervalWriter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan string, 500)

	sink := func(data string) {
		ch <- data
	}

	ticker := glock.NewMockTicker(1 * time.Second)
	writer := newIntervalWriter(ctx, ticker, sink)

	writer.Write([]byte("1"))
	select {
	case <-ch:
		t.Fatalf("ch has data")
	default:
	}

	ticker.BlockingAdvance(1 * time.Second)

	select {
	case d := <-ch:
		if d != "1" {
			t.Fatalf("wrong data in sink")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("ch has NO data")
	}

	writer.Write([]byte("2"))
	writer.Write([]byte("3"))
	writer.Write([]byte("4"))
	writer.Write([]byte("5"))

	select {
	case <-ch:
		t.Fatalf("ch has data")
	default:
	}

	cancel()
	writer.Close()

	select {
	case d := <-ch:
		if d != "2345" {
			t.Fatalf("wrong data in sink")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("ch has NO data")
	}
}
