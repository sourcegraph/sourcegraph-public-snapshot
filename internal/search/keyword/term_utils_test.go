package keyword

import (
	"testing"
)

func TestRemovePunctuation(t *testing.T) {
	tests := []struct {
		term        string
		wantTerm    string
	}{
		{
			term: "“abc123",
			wantTerm: "abc123",
		},
		{
			term: "“!!abc123",
			wantTerm: "abc123",
		},
		{
			term: "abc!”",
			wantTerm: "abc",
		},
		{
			term: "/abc/.*",
			wantTerm: "abc",
		},
		{
			term: "package/name",
			wantTerm: "package/name",
		},
		{
			term: "!!??",
			wantTerm: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			got := removePunctuation(tt.term)
			if got != tt.wantTerm {
				t.Errorf("incorrect result, got %s, want %s", got, tt.wantTerm)
			}
		})
	}
}
