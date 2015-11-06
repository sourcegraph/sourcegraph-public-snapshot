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
// "myconfig.json").
func OpenFileSystem(configName string, fs vfs.FileSystem) (*Store, error) {
	f, err := fs.Open(configName)
	if os.IsNotExist(err) {
		return &Store{
			fs:         fs,
			configName: configName,
			Data:       make(map[string]interface{}),
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	// Unmarshal the config.
	s := &Store{fs: fs, configName: configName}
	return s, json.NewDecoder(f).Decode(&s.Data)
}

// Store represents a storage for keys and values.
type Store struct {
	fs vfs.FileSystem

	// configName is the name of the config file.
	configName string

	// Data is the dataset which is JSON-encoded.
	Data map[string]interface{}
}

// Close closes the store, saving all of it's contents.
func (s *Store) Close() error {
	// Create the file, truncating it if it already exists.
	f, err := s.fs.Create(s.configName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode to JSON.
	return json.NewEncoder(f).Encode(s.Data)
}
