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
	writer := newIntervalProcessWriter(ctx, ticker, sink)

	stdoutWriter := writer.StdoutWriter()
	stderrWriter := writer.StderrWriter()
	stdoutWriter.Write([]byte("1"))
	stderrWriter.Write([]byte("1"))
	select {
	case <-ch:
		t.Fatalf("ch has data")
	default:
	}

	ticker.BlockingAdvance(1 * time.Second)

	select {
	case d := <-ch:
		want := "stdout: 1\nstderr: 1\n"
		if d != want {
			t.Fatalf("wrong data in sink. want=%q, have=%q", want, d)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("ch has NO data")
	}

	stdoutWriter.Write([]byte("2"))
	stderrWriter.Write([]byte("2"))
	stdoutWriter.Write([]byte("3"))
	stderrWriter.Write([]byte("3"))
	stdoutWriter.Write([]byte("4"))
	stderrWriter.Write([]byte("4"))
	stdoutWriter.Write([]byte("5"))
	stderrWriter.Write([]byte("5"))
	stdoutWriter.Write([]byte(`Hello world: 1
`))
	stderrWriter.Write([]byte(`Hello world: 1
`))

	select {
	case <-ch:
		t.Fatalf("ch has data")
	default:
	}

	cancel()
	writer.Close()

	select {
	case d := <-ch:
		want := "stdout: 2\nstderr: 2\n" +
			"stdout: 3\nstderr: 3\n" +
			"stdout: 4\nstderr: 4\n" +
			"stdout: 5\nstderr: 5\n" +
			"stdout: Hello world: 1\nstderr: Hello world: 1\n"

		if d != want {
			t.Fatalf("wrong data in sink. want=%q, have=%q", want, d)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("ch has NO data")
	}
}
