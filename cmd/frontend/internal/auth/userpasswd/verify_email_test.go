package userpasswd

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachEmailVerificationToPasswordReset(t *testing.T) {
	resetURL, err := url.Parse("/password-reset?code=foo&userID=42")
	require.NoError(t, err)

	db := dbmocks.NewMockUserEmailsStore()
	db.SetLastVerificationFunc.SetDefaultReturn(nil)

	newURL, err := AttachEmailVerificationToPasswordReset(context.Background(), db, *resetURL, 42, "foobar@bobheadxi.dev")
	assert.NoError(t, err)

	rendered := newURL.String()
	t.Log(rendered)
	assert.NotEqual(t, resetURL.String(), rendered)
	assert.True(t, strings.Contains(rendered, "userID=42"))
	assert.True(t, strings.Contains(rendered, "code=foo"))
	assert.True(t, strings.Contains(rendered, "email=foobar%40bobheadxi.dev"))
	assert.True(t, strings.Contains(rendered, "emailVerifyCode="))
}
