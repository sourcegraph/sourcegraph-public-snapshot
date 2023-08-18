package search

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExhaustiveSearchRepoHandler_Handle(t *testing.T) {
	logger := logtest.Scoped(t)

	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "Default",
			expectedErr: errors.New("not implemented"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := exhaustiveSearchRepoHandler{}
			err := handler.Handle(context.Background(), logger, nil)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
