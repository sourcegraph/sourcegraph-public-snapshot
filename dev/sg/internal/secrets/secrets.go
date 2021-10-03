package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

// Store holds secrets regardless on their form, as long as they are marshallable in JSON.
type Store struct {
	m map[string]interface{}
}

// New returns an empty store.
func New() *Store {
	return &Store{m: map[string]interface{}{}}
}

// Load deserialize data from a given io.Reader into a Store, returning an error if
// deserialization fails.
func Load(r io.Reader) (*Store, error) {
	s := New()
	dec := json.NewDecoder(r)
	err := dec.Decode(&s.m)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// LoadFile deserialize from a file into a Store, returning an error if
// deserialization fails.
func LoadFile(filepath string) (*Store, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Load(f)
}

// Save persists in a file the content of the store.
func (s *Store) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s.m)
}

// SaveFile persists in a file the content of the store.
func (s *Store) SaveFile(filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.Save(f)
}

// Put stores a value in memory.
func (s *Store) Put(key string, data interface{}) error {
	s.m[key] = data
	return nil
}

// Get fetches a value from memory.
func (s *Store) Get(key string) (interface{}, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("%w: %s not found", ErrSecretNotFound, key)
}
