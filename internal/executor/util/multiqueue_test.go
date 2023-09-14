package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatQueueNamesForMetrics(t *testing.T) {
	tests := []struct {
		name       string
		queueName  string
		queueNames []string
		want       string
	}{
		{
			name:      "single queue",
			queueName: "single",
			want:      "single",
		},
		{
			name:       "multiple queues",
			queueNames: []string{"first", "second"},
			want:       "first_second",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatQueueNamesForMetrics(tt.queueName, tt.queueNames)
			assert.Equalf(t, tt.want, formatted, "FormatQueueNamesForMetrics(%v)", tt.queueNames)
		})
	}
}
