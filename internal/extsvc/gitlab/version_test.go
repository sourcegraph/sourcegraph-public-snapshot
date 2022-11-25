package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	ctx := context.Background()
	client := createTestClient(t)

	have, err := client.GetVersion(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, have)
}
