// +build exectest

package cli_test

import (
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"

	app_router "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
)

func TestServeCmd_RegisterClient(t *testing.T) {
	// Start the server (#1) that our client (#2) will register with.
	a1, ctx1 := testserver.NewServer()
	defer a1.Close()

	// Start server #2.
	a2, ctx2 := testserver.NewUnstartedServer()
	a2.Config.Serve.RegisterURL = conf.AppURL(ctx1).String()
	if err := a2.Start(); err != nil {
		t.Fatal(err)
	}
	defer a2.Close()

	k2 := idkey.FromContext(ctx2)

	{
		// Server #2 should automatically register with #1 when it
		// starts up.
		for tries := 3; tries >= 0; tries-- {
			if _, err := a1.Client.RegisteredClients.Get(ctx1, &sourcegraph.RegisteredClientSpec{ID: k2.ID}); err != nil {
				if tries == 0 {
					t.Fatal(err)
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
	}

	{
		// Verify that #2's key can issue authenticated requests to
		// #1.
		ctxAuthed := sourcegraph.WithCredentials(ctx1,
			k2.TokenSource(context.Background(),
				conf.AppURL(ctx1).ResolveReference(app_router.Rel.URLTo(app_router.OAuth2ServerToken)).String(),
			),
		)
		authInfo, err := a1.Client.Auth.Identify(ctxAuthed, &pbtypes.Void{})
		if err != nil {
			t.Fatal(err)
		}
		if authInfo.ClientID != k2.ID {
			t.Errorf("got client ID %q want %q", authInfo.ClientID, k2.ID)
		}
	}
}
