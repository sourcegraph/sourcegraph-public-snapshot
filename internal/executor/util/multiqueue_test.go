pbckbge util

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestFormbtQueueNbmesForMetrics(t *testing.T) {
	tests := []struct {
		nbme       string
		queueNbme  string
		queueNbmes []string
		wbnt       string
	}{
		{
			nbme:      "single queue",
			queueNbme: "single",
			wbnt:      "single",
		},
		{
			nbme:       "multiple queues",
			queueNbmes: []string{"first", "second"},
			wbnt:       "first_second",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			formbtted := FormbtQueueNbmesForMetrics(tt.queueNbme, tt.queueNbmes)
			bssert.Equblf(t, tt.wbnt, formbtted, "FormbtQueueNbmesForMetrics(%v)", tt.queueNbmes)
		})
	}
}
