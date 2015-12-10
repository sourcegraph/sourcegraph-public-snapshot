package idkeystore

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"src.sourcegraph.com/sourcegraph/auth/idkey"

	"gopkg.in/inconshreveable/log15.v2"
)

// GenerateOrGetIDKey reads the server's ID key (or creates one on-demand)
func GenerateOrGetIDKey(idKeyData string, idKeyFile string) (*idkey.IDKey, error) {
	stringStore := &stringStore{idKeyData}
	fileStore := &fileStore{idKeyFile}

	getStores := []idStoreGet{stringStore, fileStore}
	for _, s := range getStores {
		k, err := s.Get()
		if k == nil {
			continue
		}
		return k, err
	}

	log15.Info("Generating new Sourcegraph ID key")
	k, err := idkey.Generate()
	if err != nil {
		return nil, err
	}

	err = fileStore.Put(k)
	return k, err
}

type idStoreGet interface {
	// Get returns the key. If the key is not set nil is returned
	Get() (*idkey.IDKey, error)
}

type idStorePut interface {
	// Put stores the key
	Put(*idkey.IDKey) error
}

type stringStore struct {
	IDKeyData string
}

func (s *stringStore) Get() (*idkey.IDKey, error) {
	if s.IDKeyData == "" {
		return nil, nil
	}
	log15.Debug("Reading ID key from environment (or CLI flag).")
	return idkey.FromString(s.IDKeyData)
}

type fileStore struct {
	IDKeyFile string
}

func (s *fileStore) Get() (*idkey.IDKey, error) {
	if data, err := ioutil.ReadFile(s.path()); err == nil {
		// File exists.
		return idkey.New(data)
	} else if os.IsNotExist(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

func (s *fileStore) Put(k *idkey.IDKey) error {
	idKeyFile := s.path()
	log15.Info("Storing new Sourcegraph ID key", "path", idKeyFile)
	data, err := k.MarshalText()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(idKeyFile), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(idKeyFile, data, 0600)
}

func (s *fileStore) path() string {
	// Fallback to old file based location for non-platform storage uses
	idKeyFile := s.IDKeyFile
	if idKeyFile == "" {
		idKeyFile = "$SGPATH/id.pem"
	}
	return os.ExpandEnv(idKeyFile)
}
