package adminanalytics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopCache(t *testing.T) {
	// Error on get to simulate cache miss
	_, err := getItemFromCache[any](NoopCache{}, "test")
	require.Error(t, err)

	_, err = getArrayFromCache[any](NoopCache{}, "test")
	require.Error(t, err)

	// No error on set
	err = setDataToCache(NoopCache{}, "test", "test", 1)
	require.NoError(t, err)

	summary := "test"
	err = setItemToCache[string](NoopCache{}, "test", &summary)
	require.NoError(t, err)

	err = setArrayToCache[string](NoopCache{}, "test", []*string{&summary})
	require.NoError(t, err)
}
