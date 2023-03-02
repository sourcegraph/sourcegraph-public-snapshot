package worker

import (
	"testing"
)

func TestHandler_PreDequeue(t *testing.T) {
	//logger := logtest.Scoped(t)

	tests := []struct {
		name              string
		options           Options
		expectedDequeue   bool
		expectedExtraArgs any
		expectedErr       error
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
		})
	}
}
