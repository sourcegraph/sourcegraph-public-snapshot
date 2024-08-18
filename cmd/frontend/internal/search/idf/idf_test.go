package idf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStatsAggregator_ProcessDoc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTerm map[string]int
	}{
		{
			name:     "Empty input",
			input:    "",
			wantTerm: map[string]int{},
		},
		{
			name:     "Single word",
			input:    "func",
			wantTerm: map[string]int{"func": 1},
		},
		{
			name:     "Multiple words",
			input:    "func ProcessDoc(text string) {",
			wantTerm: map[string]int{"func": 1, "ProcessDoc": 1, "text": 1, "string": 1},
		},
		{
			name:     "Repeated words",
			input:    "for word := range words { word = word }",
			wantTerm: map[string]int{"for": 1, "word": 3, "range": 1, "words": 1},
		},
		{
			name:     "Invalid words",
			input:    "s.TermToDocCt[word]++",
			wantTerm: map[string]int{"TermToDocCt": 1, "word": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StatsAggregator{
				TermToDocCt: make(map[string]int),
			}
			s.ProcessDoc(tt.input)

			if diff := cmp.Diff(s.TermToDocCt, tt.wantTerm); diff != "" {
				t.Errorf("ProcessDoc() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
