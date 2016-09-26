package vfsutil

import "testing"

func TestGitHubRepoVFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	// Any public repo will work.
	fs := &GitHubRepoVFS{
		repo: "github.com/gorilla/schema",
		rev:  "0164a00ab4cd01d814d8cd5bf63fd9fcea30e23b",
	}
	want := map[string]string{
		"/LICENSE":         "...",
		"/README.md":       "schema...",
		"/cache.go":        "// Copyright...",
		"/converter.go":    "// Copyright...",
		"/decoder.go":      "// Copyright...",
		"/decoder_test.go": "// Copyright...",
		"/doc.go":          "// Copyright...",
		"/.travis.yml":     "...",
	}

	testVFS(t, fs, want)
}

func TestGitHubRepoVFS_subtree(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	// Any public repo will work.
	fs := &GitHubRepoVFS{
		repo:    "github.com/gorilla/rpc",
		rev:     "e592e2e099465ae27afa66ec089d570904cd2d53",
		subtree: "protorpc",
	}
	want := map[string]string{
		"/doc.go":           "// Copyright 2...",
		"/protorpc_test.go": "// Copyright 2...",
		"/server.go":        "// Copyright 2...",
	}

	testVFS(t, fs, want)
}
