package codygateway

import (
	"net/http/httptest"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOverwriteErrorSource(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(500)
	originalErr := types.NewErrStatusNotOK("Foobar", rec.Result())

	err := overwriteErrSource(originalErr)
	require.Error(t, err)
	statusErr, ok := types.IsErrStatusNotOK(err)
	require.True(t, ok)
	autogold.Expect("Sourcegraph Cody Gateway").Equal(t, statusErr.Source)

	assert.NoError(t, overwriteErrSource(nil))
	assert.Equal(t, "asdf", overwriteErrSource(errors.New("asdf")).Error())
}
