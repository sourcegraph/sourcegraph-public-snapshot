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
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/vfs"
)

// Open is short-hand for:
//
//  OpenFileSystem(configName, storage.Namespace(ctx, appName+"-config", repo))
//
func Open(ctx context.Context, appName, configName, repo string) (*Store, error) {
	return OpenFileSystem(configName, storage.Namespace(ctx, appName+"-config", repo))
}

// OpenFileSystem opens a configuration store with the given filename (e.g.
// "myconfig.json") creating it if needed on the given filesystem.
func OpenFileSystem(configName string, fs vfs.FileSystem) (*Store, error) {
	f, err := fs.Open(configName)
	if os.IsNotExist(err) {
		// Create the config file then.
		f, err := fs.Create(configName)
		if err != nil {
			return nil, err
		}
		return &Store{
			fs:   fs,
			f:    f,
			Data: make(map[string]interface{}),
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	// Unmarshal the config.
	s := &Store{fs: fs, f: f}
	return s, json.NewDecoder(f).Decode(&s.Data)
}

// Store represents a storage for keys and values.
type Store struct {
	f  vfs.File
	fs vfs.FileSystem

	// Data is the dataset which is JSON-encoded.
	Data map[string]interface{}
}

// Close closes the store, saving all of it's contents.
func (s *Store) Close() error {
	// Recreate the file.
	//
	// TODO(slimsag): once we implement File.Truncate, we can avoid this.
	var err error
	if err = s.f.Close(); err != nil {
		return err
	}
	s.f, err = s.fs.Create(s.f.Name())
	if err != nil {
		return err
	}
	defer s.f.Close()

	// Encode to JSON.
	return json.NewEncoder(s.f).Encode(s.Data)
}
