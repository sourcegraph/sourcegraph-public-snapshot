// Package config implements per-repository and global configuration on platform storage.
//
// The keys are strings and should be human-readable, as they may be displayed
// in e.g. a configuration UI.
//
// The values are simple values also designed to be specified by humans, they
// must be of type bool, string, or int64.
//
// An application may have one global configuration, and one configuration per
// repository URI.
//
// Example Usage
//
//  // Open the config store specific to "myapp" and local to "my/repo".
//  cfg, err := config.Open(ctx, "myapp", "my/repo")
//  if err != nil {
//    log.Fatal(err)
//  }
//
//  // Check if a config value exists.
//  if v, ok := cfg.Data["mykey"]; ok {
//    fmt.Println("mykey has a value of", v)
//  }
//
//  // Store some values.
//  cfg.Data["mykey"] = "hello world!"
//  cfg.Data["mysecondkey"] = 5
//
//  // Save the config.
//  if err := cfg.Close(); err != nil {
//    log.Fatal(err)
//  }
//
package config

import (
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

const (
	bucket = "config"
	key    = "data"
)

// Open is short-hand for:
//
//  OpenWith(storage.Namespace(ctx, appName+"-config", repo))
//
func Open(ctx context.Context, appName, repo string) (*Store, error) {
	return OpenWith(storage.Namespace(ctx, appName+"-config", repo))
}

// OpenWith opens a configuration store with the storage system.
func OpenWith(sys storage.System) (*Store, error) {
	s := &Store{sys: sys}
	err := storage.GetJSON(sys, bucket, key, &s.Data)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	return s, nil
}

// Store represents a storage for keys and values.
type Store struct {
	// Data is the dataset which is JSON-encoded.
	Data map[string]interface{}

	sys storage.System
}

// Close closes the store, saving all of it's contents.
func (s *Store) Close() error {
	return storage.PutJSON(s.sys, bucket, key, s.Data)
}
