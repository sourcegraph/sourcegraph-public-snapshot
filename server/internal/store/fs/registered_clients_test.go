package fs

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestRegisteredClients_Get_existing(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Get_existing(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Get_nonexistent(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Get_nonexistent(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_GetByCredentials_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_ok(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_GetByCredentials_badID(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_badID(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_GetByCredentials_badSecret(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_badSecret(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_GetByCredentials_noSecretOrJWKS(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_GetByCredentials_noSecretOrJWKS(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Create_secret_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_secret_ok(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Create_jwks_ok(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_jwks_ok(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Create_duplicate(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_duplicate(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Create_noID(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_noID(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Create_noSecretOrJWKS(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Create_noSecretOrJWKS(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Update_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_ok(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Update_secret(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_secret(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Update_nonexistent(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Update_nonexistent(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Delete_ok(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Delete_ok(ctx, t, &RegisteredClients{})
}

func TestRegisteredClients_Delete_nonexistent(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.RegisteredClients_Delete_nonexistent(ctx, t, &RegisteredClients{})
}
