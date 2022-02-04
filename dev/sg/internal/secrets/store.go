package secrets

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/cockroachdb/errors"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

// Store holds secrets regardless on their form, as long as they are marshallable in JSON.
type Store struct {
	filepath string
	m        map[string]json.RawMessage
}

type storeKey struct{}

// FromContext fetches a store from context.
func FromContext(ctx context.Context) *Store {
	if store, ok := ctx.Value(storeKey{}).(*Store); ok {
		return store
	}
	return nil
}

// WithContext stores a Store in the context.
func WithContext(ctx context.Context, store *Store) context.Context {
	return context.WithValue(ctx, storeKey{}, store)
}

// New returns an empty store that if saved, will be written at filepath.
func New(filepath string) *Store {
	return &Store{filepath: filepath, m: map[string]json.RawMessage{}}
}

// LoadFile deserialize from a file into a Store, returning an error if
// deserialization fails.
func LoadFile(filepath string) (*Store, error) {
	s := New(filepath)
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
func (s *Store) Put(key string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.m[key] = b
	return nil
}

// PutAndSave saves automatically after calling Put.
func (s *Store) PutAndSave(key string, data interface{}) error {
	err := s.Put(key, data)
	if err != nil {
		return err
	}
	return s.SaveFile()
}

// Get fetches a value from memory and uses the given target to deserialize it.
func (s *Store) Get(key string, target interface{}) error {
	if v, ok := s.m[key]; ok {
		return json.Unmarshal(v, target)
	}
	return errors.Newf("%w: %s not found", ErrSecretNotFound, key)
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
