package testsuite

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// RegisteredClients_Create_noID tests the behavior of
// RegisteredClients.Create when called with an empty ID.
func RegisteredClients_Create_noID(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "", ClientSecret: "b"}); err == nil {
		t.Fatal("err == nil")
	}
}

// RegisteredClients_Create_noSecretOrJWKS tests the behavior of
// RegisteredClients.Create when called with an empty secret.
func RegisteredClients_Create_noSecretOrJWKS(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: ""}); err == nil {
		t.Fatal("err == nil")
	}
}

// RegisteredClients_Update_ok tests the behavior of
// RegisteredClients.Update when called with correct args.
func RegisteredClients_Update_ok(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	want := sourcegraph.RegisteredClient{
		ID:           "a",
		ClientSecret: "b",
		ClientURI:    "https://example.com/1",
		RedirectURIs: []string{"https://example.com/2"},
		ClientName:   "t",
		Description:  "d",
		Meta:         map[string]string{"k1": "v1", "k2": "v2"},
		Type:         sourcegraph.RegisteredClientType_SourcegraphServer,
	}
	if err := s.Create(ctx, want); err != nil {
		t.Fatal(err)
	}

	want.ClientSecret = ""
	want.ClientURI += "!"
	want.RedirectURIs[0] += "!"
	want.ClientName += "!"
	want.Description += "!"
	want.Meta["k1"] += "!"
	want.Meta["k3"] = "v3"

	if err := s.Update(ctx, want); err != nil {
		t.Fatal(err)
	}

	updated, err := s.Get(ctx, sourcegraph.RegisteredClientSpec{ID: "a"})
	if err != nil {
		t.Fatal(err)
	}

	// Don't check equality of these non-deterministic fields.
	want.UpdatedAt = pbtypes.Timestamp{}
	want.UpdatedAt = pbtypes.Timestamp{}
	updated.UpdatedAt = pbtypes.Timestamp{}
	updated.UpdatedAt = pbtypes.Timestamp{}

	if !reflect.DeepEqual(*updated, want) {
		t.Errorf("Update: got %+v, want %+v", *updated, want)
	}
}

// RegisteredClients_Update_secret tests the behavior of
// RegisteredClients.Update when called with an new Secret value.
func RegisteredClients_Update_secret(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: "b"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Update(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: "b!"}); err == nil {
		t.Fatal(err)
	}

	// Get with the initial (still valid) credentials.
	valid, err := s.GetByCredentials(ctx, sourcegraph.RegisteredClientCredentials{ID: "a", Secret: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "a"; valid.ID != want {
		t.Errorf("got %q, want %q", valid.ID, want)
	}

	// Fail to get with the newly update-attempted credentials.
	invalid, err := s.GetByCredentials(ctx, sourcegraph.RegisteredClientCredentials{ID: "a", Secret: "b!"})
	if !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
	if invalid != nil {
		t.Error("invalid != nil")
	}
}

// RegisteredClients_Update_nonexistent tests the behavior of
// RegisteredClients.Update when called with a nonexistent client ID.
func RegisteredClients_Update_nonexistent(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Update(ctx, sourcegraph.RegisteredClient{ID: "a", ClientName: "t"}); !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
}

// RegisteredClients_Delete_ok tests the behavior of
// RegisteredClients.Delete when called with correct args.
func RegisteredClients_Delete_ok(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Create(ctx, sourcegraph.RegisteredClient{ID: "a", ClientSecret: "b"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(ctx, sourcegraph.RegisteredClientSpec{ID: "a"}); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Get(ctx, sourcegraph.RegisteredClientSpec{ID: "a"}); !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
}

// RegisteredClients_Delete_nonexistent tests the behavior of
// RegisteredClients.Delete when called with a nonexistent client ID.
func RegisteredClients_Delete_nonexistent(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Delete(ctx, sourcegraph.RegisteredClientSpec{ID: "doesntexist"}); !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
}

func isRegisteredClientNotFound(err error) bool {
	_, ok := err.(*store.RegisteredClientNotFoundError)
	return ok
}
