// Package usercontent contains a store for user uploaded content.
package usercontent

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/webdav"
)

// Store for user uploaded content.
var Store = func() webdav.FileSystem {
	dir := filepath.Join(os.Getenv("SGPATH"), "usercontent")
	log.Println("usercontent.Store path:", dir)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		// TODO: Error-prone things should happen elsewhere where it can be handled better.
		log.Fatalf("Error creating directory %q: %v.\n", dir, err)
	}
	return webdav.Dir(dir)
}()
