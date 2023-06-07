package shared

import (
	"testing"
)

func TestIsContextRequiredForChatQuery(t *testing.T) {
	cases := []struct {
		name         string
		query        string
		embedding    []float32
		embeddingErr error
		want         bool
	}{
		{
			name:      "query matches no context regex",
			query:     "that answer looks incorrect",
			embedding: []float32{0.0, 1.0}, // unused
			want:      false,
		},
		{
			name:      "query requires context",
			query:     "what directory contains the cody plugin",
			embedding: []float32{1.0, 0.0}, // unused
			want:      true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := isContextRequiredForChatQuery(tt.query)
			if got != tt.want {
				t.Fatalf("expected context required to be %t but was %t", tt.want, got)
			}
		})
	}
}
