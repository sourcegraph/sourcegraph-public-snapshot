package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	DefaultFile = "sg.secrets.json"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

// Store holds secrets regardless on their form, as long as they are marshallable in JSON.
type Store struct {
	filepath string
	m        map[string]json.RawMessage

	secretmanagerOnce sync.Once
	secretmanager     *secretmanager.Client
	secretmanagerErr  error
}

type storeKey struct{}

// FromContext fetches a store from context. In sg, a store is set in the command context
// when sg starts - if the load fails, an error is printed and a store is not set.
func FromContext(ctx context.Context) (*Store, error) {
	if store, ok := ctx.Value(storeKey{}).(*Store); ok {
		return store, nil
	}
	return nil, errors.New("secrets store not available")
}

// WithContext stores a Store in the context.
func WithContext(ctx context.Context, store *Store) context.Context {
	return context.WithValue(ctx, storeKey{}, store)
}

// newStore returns an empty store that if saved, will be written at filepath.
func newStore(filepath string) *Store {
	return &Store{filepath: filepath, m: map[string]json.RawMessage{}}
}

// LoadFromFile deserialize from a file into a Store, returning an error if
// deserialization fails.
func LoadFromFile(filepath string) (*Store, error) {
	s := newStore(filepath)
	f, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return s, dec.Decode(&s.m)
}

// Write serializes the store content in the given writer.
func (s *Store) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s.m)
}

// SaveFile persists in a file the content of the store.
func (s *Store) SaveFile() error {
	f, err := os.OpenFile(s.filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.Write(f)
}

// Put stores serialized data in memory.
func (s *Store) Put(key string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.m[key] = b
	return nil
}

// PutAndSave saves automatically after calling Put.
func (s *Store) PutAndSave(key string, data any) error {
	err := s.Put(key, data)
	if err != nil {
		return err
	}
	return s.SaveFile()
}

// Get fetches a value from memory and uses the given target to deserialize it.
func (s *Store) Get(key string, target any) error {
	if v, ok := s.m[key]; ok {
		return json.Unmarshal(v, target)
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

func (s *Store) GetExternal(ctx context.Context, secret ExternalSecret) (string, error) {
	var value externalSecretValue

	// Check if we already have this secret
	if err := s.Get(secret.id(), &value); err == nil {
		return value.Value, nil
	}

	if secret.Provider != "gcloud" {
		return "", errors.Newf("Unknown secrets provider %q", secret.Provider)
	}

	client, err := s.getSecretmanagerClient(ctx)
	if err != nil {
		return "", err
	}
	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", secret.Project, secret.Name),
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to access secret %q from %q", secret.Name, secret.Project)
	}

	// cache value, but don't save - TBD if we want to persist these secrets
	value.Fetched = time.Now()
	value.Value = string(result.Payload.Data)
	s.Put(secret.id(), &value)

	return value.Value, nil
}

// Remove deletes a value from memory.
func (s *Store) Remove(key string) error {
	if _, exists := s.m[key]; exists {
		delete(s.m, key)
		return nil
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

// Keys returns out all keys
func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.m))
	for key := range s.m {
		keys = append(keys, key)
	}
	return keys
}

func (s *Store) getSecretmanagerClient(ctx context.Context) (*secretmanager.Client, error) {
	s.secretmanagerOnce.Do(func() {
		var err error
		s.secretmanager, err = secretmanager.NewClient(ctx)
		if err != nil {
			s.secretmanagerErr = errors.Errorf("failed to create secretmanager client: %v", err)
		}
	})
	return s.secretmanager, s.secretmanagerErr
}
