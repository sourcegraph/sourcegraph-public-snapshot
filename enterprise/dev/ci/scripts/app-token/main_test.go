package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenJwtToken(t *testing.T) {

	appID := os.Getenv("GITHUB_APP_ID")
	require.NotEmpty(t, appID, "GITHUB_APP_ID must be set.")
	keyPath := os.Getenv("KEY_PATH")
	require.NotEmpty(t, keyPath, "KEY_PATH must be set.")

	jwt, err := genJwtToken(appID, keyPath)
	require.NoError(t, err)
	t.Log("%+s", jwt)
}

func TestGetInstallAccessToken(t *testing.T) {

}
