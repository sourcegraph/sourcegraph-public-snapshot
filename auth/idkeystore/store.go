package idkeystore

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"

	"gopkg.in/inconshreveable/log15.v2"
)

// GenerateOrGetIDKey reads the server's ID key (or creates one on-demand)
func GenerateOrGetIDKey(ctx context.Context, idKeyData string, idKeyFile string) (*idkey.IDKey, error) {
	stringStore := &stringStore{idKeyData}
	fileStore := &fileStore{idKeyFile}
	platformStore := &platformStore{ctx}

	getStores := []idStoreGet{stringStore, fileStore, platformStore}
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

	var p idStorePut
	if idKeyFile == "" {
		// If this is not specified, default to platformStore
		p = platformStore
	} else {
		p = fileStore
	}
	err = p.Put(k)

	if err != nil && idKeyFile == "" && grpc.Code(err) == codes.AlreadyExists {
		log15.Info("Key generation race detected. Falling back to first generated ID key")
		return platformStore.Get()
	}

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

type platformStore struct {
	ctx context.Context
}

func (s *platformStore) Get() (*idkey.IDKey, error) {
	key := s.key()
	v, err := store.StorageFromContext(s.ctx).Get(s.ctx, &key)
	if err != nil {
		if grpc.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	return idkey.New(v.Value)
}

func (s *platformStore) Put(k *idkey.IDKey) error {
	data, err := k.MarshalText()
	if err != nil {
		return err
	}
	_, err = store.StorageFromContext(s.ctx).PutNoOverwrite(s.ctx, &sourcegraph.StoragePutOp{
		Key:   s.key(),
		Value: data,
	})
	return err
}

func (s *platformStore) key() sourcegraph.StorageKey {
	return sourcegraph.StorageKey{
		Bucket: &sourcegraph.StorageBucket{
			AppName: "core.serve",
			Name:    "auth",
		},
		Key: "id.pem",
	}
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
