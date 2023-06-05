package common

import (
	"container/list"
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewQueue[int](list.New())

	if !queue.Empty() {
		t.Error("Expected queue to be empty initially")
	}

	// Push 1, 2 and 3 into the queue.
	for i := 1; i < 4; i++ {
		v := i
		queue.Push(&v)
	}

	if queue.Empty() {
		t.Error("Expected queue to not be empty after pushing elements")
	}

	// Pop and expect 1, 2 and 3 in that order (FIFO queue).
	for i := 1; i < 4; i++ {
		value := queue.Pop()
		if *value != i {
			t.Errorf("Expected 1, got %d", *value)
		}
	}

	if !queue.Empty() {
		t.Error("Expected queue to be empty after popping all elements")
	}
}
