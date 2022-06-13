package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	DefaultFile = "sg.secrets.json"
)

var (
	ErrSecretNotFound = errors.New("secret not found")

	// externalSecretTTL declares how long external secrets are allowed to be persisted
	// once fetched.
	externalSecretTTL = 24 * time.Hour
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
		if time.Since(value.Fetched) < externalSecretTTL {
			return value.Value, nil
		}

		// If expired, remove the secret and fetch a new one.
		_ = s.Remove(secret.id())
		value = externalSecretValue{}
	}

	// Get secret from provider
	var err error
	switch secret.Provider {

	case ExternalProviderGCloud:
		client, err := s.getSecretmanagerClient(ctx)
		if err != nil {
			return "", err
		}
		var result *secretmanagerpb.AccessSecretVersionResponse
		result, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", secret.Project, secret.Name),
		})
		if err == nil {
			value.Value = string(result.Payload.Data)
		}

	case ExternalProvider1Pass:
		sessionToken, err := s.get1passSession(ctx)
		if err != nil {
			return "", err
		}
		value.Value, err = run.Cmd(ctx, "op read",
			run.Arg(fmt.Sprintf("op://%s/%s/%s", secret.Project, secret.Name, secret.Field)),
			`--session`, sessionToken,
			`--account="team-sourcegraph.1password.com"`).
			Run().String()

	default:
		return "", errors.Newf("Unknown secrets provider %q", secret.Provider)
	}

	if err != nil {
		errMessaage := fmt.Sprintf("%s: failed to access secret %q from %q",
			secret.Provider, secret.Name, secret.Project)
		// Some secret providers use their respective CLI, if not found the user might not
		// have run 'sg setup' to set up the relevant tool.
		if strings.Contains(err.Error(), "command not found") {
			errMessaage += "- you may need to run 'sg setup' again"
		}
		return "", errors.Wrap(err, errMessaage)
	}

	// Return and persist the fetched secret
	value.Fetched = time.Now()
	return value.Value, s.PutAndSave(secret.id(), &value)
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

// getSecretmanagerClient instantiates a Google Secrets Manager client once and returns it.
func (s *Store) getSecretmanagerClient(ctx context.Context) (*secretmanager.Client, error) {
	s.secretmanagerOnce.Do(func() {
		var err error
		s.secretmanager, err = secretmanager.NewClient(ctx)
		if err != nil {
			const defaultMessage = "failed to create Google Secrets Manager client"
			if strings.Contains(err.Error(), "could not find default credentials") {
				s.secretmanagerErr = errors.Errorf("%s: %v - you might need to run 'sg setup' again to set up 'gcloud'",
					defaultMessage, err)
			} else {
				s.secretmanagerErr = errors.Wrap(err, defaultMessage)
			}
		}
	})
	return s.secretmanager, s.secretmanagerErr
}

func (s *Store) get1passSession(ctx context.Context) (string, error) {
	var sessionToken map[string]string
	if err := s.Get("1pass-session", &sessionToken); err != nil {
		return "", err
	}
	return sessionToken["token"], nil
}
