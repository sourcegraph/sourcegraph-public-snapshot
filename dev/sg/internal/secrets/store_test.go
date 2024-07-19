package secrets

import (
	"bytes"
	"context"
	"os"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go/v2"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mySecrets struct {
	ID     string
	Secret string
}

func TestSecrets(t *testing.T) {
	t.Run("Put and Get", func(t *testing.T) {
		data := mySecrets{ID: "foo", Secret: "bar"}
		store := newStore("")
		err := store.Put("foo", data)
		if err != nil {
			t.Fatalf("want no error, got %q", err)
		}

		want := data
		got := mySecrets{}
		err = store.Get("foo", &got)
		if err != nil {
			t.Fatalf("want no error when getting secret, but got: %q", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("wrong secret data. (-want +got):\n%s", diff)
		}
	})

	t.Run("LoadFile returns an error on invalid JSON", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
		}
		if _, err := f.WriteString(`{"foo":1`); err != nil {
			t.Fatalf("couldn't write in temp secret file: %q", err)
		}
		f.Close()
		filepath := f.Name()
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		_, err = LoadFromFile(filepath)
		if err == nil {
			t.Fatal("want an error but got none")
		}
	})

	t.Run("LoadFile doesn't fail when file is empty", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
		}
		f.Close()
		filepath := f.Name()
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		got, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("want no error when loading an empty file, but got %q instead", err)
		}
		if got == nil {
			t.Fatal("want store to not be nil")
		}
	})

	t.Run("SaveFile and LoadFile", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
		}
		f.Close()
		filepath := f.Name()
		_ = os.Remove(filepath) // we just want the path, not the file
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		// Assign a secret and save it
		s, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("couldn't load temp secret file: %q", err)
		}
		data := map[string]any{"key": "val"}
		if err := s.Put("foo", data); err != nil {
			t.Fatalf("want no error when putting secret, got %q", err)
		}
		err = s.SaveFile()
		if err != nil {
			t.Fatalf("failed to save secrets: %q", err)
		}

		// Fetch it back and compare
		got, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("couldn't load temp secret file: %q", err)
		}
		if diff := cmp.Diff(s.persistedData, got.persistedData); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})

	t.Run("GetExternal does not get saved by SaveFile", func(t *testing.T) {
		store := newStore("")
		store.secretmanagerOnce.Do(func() {
			store.secretmanager = &mockSecretManagerClient{
				payload: &secretmanagerpb.SecretPayload{
					Data: []byte(t.Name()),
				},
			}
		})

		es := ExternalSecret{
			Project: "foo",
			Name:    "bar",
		}
		secret, err := store.GetExternal(context.Background(), es)
		require.NoError(t, err)
		assert.Equal(t, t.Name(), secret)
		assert.Equal(t, secret, store.externalData[es.id()].Value, "Stored in-memory")
		assert.Len(t, store.persistedData, 0, "Nothing pending persistence to disk")

		var persisted bytes.Buffer
		require.NoError(t, store.Write(&persisted))
		autogold.Expect("{}\n").Equal(t, persisted.String())
	})
}

type mockSecretManagerClient struct {
	payload *secretmanagerpb.SecretPayload
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name:    req.GetName(),
		Payload: m.payload,
	}, nil
}
