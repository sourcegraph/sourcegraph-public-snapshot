// +build pgsqltest

package pgsql

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func isRegisteredClientNotFound(err error) bool {
	_, ok := err.(*store.RegisteredClientNotFoundError)
	return ok
}

// TestRegisteredClients_Get_existing tests the behavior of
// RegisteredClients.Get when called on a client that exists (i.e.,
// the successful outcome).
func TestRegisteredClients_Get_existing(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &registeredClients{}
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: "b"}); err != nil {
		t.Fatal(err)
	}

	client, err := s.Get(ctx, sourcegraph.RegisteredClientSpec{ID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Error("client == nil")
	}
	if want := "a"; client.ID != want {
		t.Errorf("got ID %q, want %q", client.ID, want)
	}
}

// TestRegisteredClients_Get_nonexistent tests the behavior of
// RegisteredClients.Get when called on a client that does not exist.
func TestRegisteredClients_Get_nonexistent(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &registeredClients{}
	client, err := s.Get(ctx, sourcegraph.RegisteredClientSpec{ID: "doesntexist"})
	if !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
	if client != nil {
		t.Error("client != nil")
	}
}

// TestRegisteredClients_GetByCredentials_ok tests the behavior of
// RegisteredClients.Get when called with the correct credentials.
func TestRegisteredClients_GetByCredentials_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &registeredClients{}
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: "b"}); err != nil {
		t.Fatal(err)
	}

	client, err := s.GetByCredentials(ctx, sourcegraph.RegisteredClientCredentials{ID: "a", Secret: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Error("client != nil")
	}
	if want := "a"; client.ID != want {
		t.Errorf("got ID %q, want %q", client.ID, want)
	}
}

func TestRegisteredClients_GetByCredentials_badID(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_badID(ctx, t, &registeredClients{})
}

func TestRegisteredClients_GetByCredentials_badSecret(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_badSecret(ctx, t, &registeredClients{})
}

func TestRegisteredClients_GetByCredentials_noSecretOrJWKS(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_noSecretOrJWKS(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Create_secret_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_secret_ok(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Create_jwks_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_jwks_ok(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Create_duplicate(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_duplicate(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Create_noID(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_noID(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Create_noSecretOrJWKS(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_noSecretOrJWKS(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Update_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_ok(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Update_secret(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_secret(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Update_nonexistent(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_nonexistent(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Delete_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Delete_ok(ctx, t, &registeredClients{})
}

func TestRegisteredClients_Delete_nonexistent(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Delete_nonexistent(ctx, t, &registeredClients{})
}
