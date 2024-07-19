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
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	DefaultFile     = "sg.secrets.json"
	LocalDevProject = "sourcegraph-local-dev"
)

var (
	ErrSecretNotFound = errors.New("secret not found")

	// externalSecretTTL declares how long external secrets are allowed to be persisted
	// once fetched.
	externalSecretTTL = 24 * time.Hour
)

type FallbackFunc func(context.Context) (string, error)

type secretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

// Store holds secrets regardless on their form, as long as they are marshallable in JSON.
type Store struct {
	filepath string
	// persistedData holds secrets that should be persisted to filepath.
	persistedData map[string]json.RawMessage

	// externalData holds secrets that are fetched from external sources. This
	// is NOT persisted to disk, as it should be fetched from the external source
	// on every use for security.
	externalData map[string]externalSecretValue

	secretmanagerOnce sync.Once
	secretmanager     secretManagerClient
	secretmanagerErr  error
}

type storeKey struct{}

// SecretErr is the error that occurs when we fail to get a secret. It contains the original
// error as well as the secret key that failed to fetch.
type SecretErr struct {
	// Err is the original error
	Err error
	// Key is the secret key that failed to fetch
	Key string
}

func (se SecretErr) Error() string {
	return fmt.Sprintf("failed to get secret %q: %v", se.Key, se.Err)
}

// GoogleSecretErr is an error that occurs when we fail to fetch a secret from Google Secret Manger in a particular GCP Project.
// It contains the key that failed to fetch, the original error and the GCP project name.
type GoogleSecretErr struct {
	SecretErr
	// Project is the GCP project where we failed to fetch the secret
	Project string
}

func (gse GoogleSecretErr) Error() string {
	return fmt.Sprintf("google(%s): %s", gse.Project, gse.SecretErr.Error())
}

// CommandErr is an error that occurs when we fail to get a secret by executing some CLI Command.
// It contains the original error as well as the secret key that failed to fetch.
type CommandErr struct {
	SecretErr
}

func (ce CommandErr) Error() string {
	return fmt.Sprintf("command error: %v", ce.SecretErr.Error())
}

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
	return &Store{
		filepath:      filepath,
		persistedData: map[string]json.RawMessage{},
		externalData:  map[string]externalSecretValue{},
	}
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
	if err := dec.Decode(&s.persistedData); err != nil {
		// Ignore EOF which is returned when the file is empty, we just pretend the file isn't there.
		// Note that invalid JSON might still return "unexpected EOF" and
		// we let that one get through.
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}
	var deletedExpired bool
	for k := range s.persistedData {
		// Migration: external secrets that were persisted before we stopped
		// persisting external secrets
		var value externalSecretValue
		if err := s.Get(k, &value); err != nil {
			continue
		}
		if !value.Fetched.IsZero() && value.Value != "" {
			deletedExpired = true
			delete(s.persistedData, k)
		}
	}
	if deletedExpired {
		return s, s.SaveFile()
	}
	return s, nil
}

// Write serializes the store content in the given writer.
func (s *Store) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s.persistedData)
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
	s.persistedData[key] = b
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
	if v, ok := s.persistedData[key]; ok {
		return json.Unmarshal(v, target)
	}
	if v, ok := s.externalData[key]; ok {
		// Must be an *externalSecretValue
		if tv, ok := target.(*externalSecretValue); ok {
			*tv = v
			return nil
		}
		return errors.Newf("target must be *externalSecretValue, got %T", target)
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

func (s *Store) GetExternal(ctx context.Context, secret ExternalSecret, fallbacks ...FallbackFunc) (string, error) {
	var value externalSecretValue

	// Check if we already have this secret
	if err := s.Get(secret.id(), &value); err == nil {
		if time.Since(value.Fetched) < externalSecretTTL {
			return value.Value, nil
		}

		// If expired, fetch a new one. We don't need to remove anything because
		// external secrets are only stored in-memory.
		value = externalSecretValue{}
	}

	// Get secret from provider
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

	// Failed to get the secret normally, so lets try getting it with the fallback if it exists
	if err != nil && len(fallbacks) > 0 {
		for _, fallback := range fallbacks {
			val, fallbackErr := fallback(ctx)

			if fallbackErr != nil {
				err = errors.Wrap(err, fallbackErr.Error())
			} else {
				value.Value = val
				// Since we were able to get a secret using the fallback, we set the error to nil
				// this also ensures that the fallback value is also saved to the store
				err = nil
				break
			}
		}
	}

	if err != nil {
		secretErr := SecretErr{Err: err, Key: secret.Name}
		// Some secret providers use their respective CLI, if not found the user might not
		// have run 'sg setup' to set up the relevant tool.
		if strings.Contains(err.Error(), "command not found") {
			return "", CommandErr{SecretErr: secretErr}
		}
		return "", GoogleSecretErr{
			SecretErr: secretErr,
			Project:   secret.Project,
		}
	}

	// Return and persist the fetched secret
	value.Fetched = time.Now()
	s.externalData[secret.id()] = value
	return value.Value, nil
}

// Remove deletes a value from memory.
func (s *Store) Remove(key string) error {
	if _, exists := s.persistedData[key]; exists {
		delete(s.persistedData, key)
		return nil
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
}

// Keys returns out all keys
func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.persistedData))
	for key := range s.persistedData {
		keys = append(keys, key)
	}
	return keys
}

// getSecretmanagerClient instantiates a Google Secrets Manager client once and returns it.
func (s *Store) getSecretmanagerClient(ctx context.Context) (secretManagerClient, error) {
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
