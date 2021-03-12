package graphqlbackend

import (
	"testing"
	"time"
)

// Compare BenchmarkWithoutTimer and BenchmarkWithTimer to see how much overhead
// is added by keeping track of "time to first event".
func BenchmarkWithoutTimer(b *testing.B) {
	s := StreamFunc(func(event SearchEvent) {})
	se := SearchEvent{Results: make([]SearchResultResolver, 10)}
	for i := 0; i < b.N; i++ {
		s.Send(se)
	}
}

func BenchmarkWithTimer(b *testing.B) {
	s := WithTimer(StreamFunc(func(event SearchEvent) {}), time.Now())
	se := SearchEvent{Results: make([]SearchResultResolver, 10)}
	for i := 0; i < b.N; i++ {
		s.Send(se)
	}
}

func TestWithTimer(t *testing.T) {
	s := WithTimer(StreamFunc(func(event SearchEvent) {}), time.Now())
	time.Sleep(1 * time.Millisecond)
	if got := s.Latency(); got != 0 {
		t.Fatalf("want 0, got %d ", got)
	}
	s.Send(SearchEvent{Results: make([]SearchResultResolver, 1)})
	if got := s.Latency().Milliseconds(); got < 1 {
		t.Fatalf("want >=1, got %d ", got)
	}
}
