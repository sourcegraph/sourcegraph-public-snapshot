package main

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/graph"
	"github.com/sourcegraph/sourcegraph/dev/depgraph/root"
)

// loadDependencyGraph reads a cached dependency graph or recalculates one if no cache file
// is present. If a new dependency graph is loaded it will be re-serialized to the cache file.
func loadDependencyGraph() (*graph.DependencyGraph, error) {
	if graph, err := loadDependencyGraphFromCache(); graph != nil || err != nil {
		return graph, err
	}

	graph, err := graph.Load()
	if err != nil {
		return nil, err
	}

	return graph, writeDependencyGraphCache(graph)
}

// loadDependencyGraphFromCache reads the cache file (if it exists) and decodes a dependency
// graph from it. If no cache file exists, a nil graph is returned.
func loadDependencyGraphFromCache() (graph *graph.DependencyGraph, _ error) {
	cacheFile, err := cacheFile()
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	if err := gob.NewDecoder(bytes.NewReader(contents)).Decode(&graph); err != nil {
		return nil, err
	}

	return graph, nil
}

// writeDependencyGraphCache gob-encodes the dependency graph and writes it to a cache file.
func writeDependencyGraphCache(graph *graph.DependencyGraph) error {
	buffer := &bytes.Buffer{}
	if err := gob.NewEncoder(buffer).Encode(graph); err != nil {
		return err
	}

	cacheFile, err := cacheFile()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(cacheFile, buffer.Bytes(), os.ModePerm)
}

// clearCache removes the cache file, if it exists.
func clearCache() error {
	cacheFile, err := cacheFile()
	if err != nil {
		return err
	}

	if err := os.Remove(cacheFile); !os.IsNotExist(err) {
		return err
	}

	return nil
}

const cacheFileName = "depgraph.cache"

// cacheFile returns the absolute path to the cache file.
func cacheFile() (string, error) {
	root, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, cacheFileName), nil
}
