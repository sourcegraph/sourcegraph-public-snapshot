package vfsutil

import "testing"

func TestGitRepoVFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	fs := &GitRepoVFS{
		CloneURL: "git://github.com/gorilla/schema",
		Rev:      "0164a00ab4cd01d814d8cd5bf63fd9fcea30e23b",
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

func TestGitRepoVFS_subtree(t *testing.T) {
	if testing.Short() {
		t.Skip("skip network-intensive test")
	}

	// Any public repo will work.
	fs := &GitRepoVFS{
		CloneURL: "git://github.com/gorilla/rpc",
		Rev:      "e592e2e099465ae27afa66ec089d570904cd2d53",
		Subtree:  "protorpc",
	}
	want := map[string]string{
		"/doc.go":           "// Copyright 2...",
		"/protorpc_test.go": "// Copyright 2...",
		"/server.go":        "// Copyright 2...",
	}

	testVFS(t, fs, want)
}
