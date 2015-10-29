// Package config implements a basic configuration store on platform storage.
//
// Example Usage
//
//  // Open/create repo-local config.
//  cfg, err := config.Open("myappname", "my-config-vars.json", ctx, repo)
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
//  // Close/save config.
//  if err := cfg.Close(); err != nil {
//    log.Fatal(err)
//  }
//
package config

import (
	"encoding/json"
	"os"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

// Open is short-hand for:
//
//  OpenFileSystem(configName, storage.Namespace(ctx, appName+"-config", repo))
//
func Open(appName, configName string, ctx context.Context, repo *sourcegraph.RepoSpec) *Store {
	return OpenFileSystem(configName, storage.Namespace(ctx, appName+"-config", repo))
}

// OpenFileSystem opens a configuration store with the given filename (e.g.
// "myconfig.json") creating it if needed on the given filesystem.
func OpenFileSystem(configName string, fs storage.FileSystem) (*Store, error) {
	f, err := fs.Open(configName)
	if err == os.ErrNotExist {
		// Create the config file then.
		f, err := fs.Create(configName)
		if err != nil {
			return nil, err
		}
		return &Store{fs: fs, f: f}, nil
	} else if err != nil {
		return nil, err
	}
	return &Store{fs: fs, f: f}, nil
}

// Store represents a storage for keys and values.
type Store struct {
	f  storage.File
	fs storage.FileSystem

	// Data is the dataset which is JSON-encoded.
	Data map[interface{}]interface{}
}

// Close closes the store, saving all of it's contents.
func (s *Store) Close() error {
	// Recreate the file.
	//
	// TODO(slimsag): once we implement File.Truncate, we can avoid this.
	if err := s.f.Close(); err != nil {
		return err
	}
	s.f, err = s.fs.Create(s.f.Name())
	if err != nil {
		return err
	}

	// Encode to JSON.
	if err := json.NewEncoder(s.f).Encode(s.Data); err != nil {
		return err
	}
	return s.f.Close()
}
